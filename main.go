package main

import (
	"encoding/json"
	"fmt"
	//"time"
	//"log"
	"net/http"
	"strconv"
	//"strings"
	//"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/influxdata/influxdb/client/v2"
	//	"golang.org/x/crypto/acme/autocert"
)

// Check Error
func checkErr(e error) {
	if e != nil {
		fmt.Println(e)
		//panic(e)
	}
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func setIDW() {
	res, err := queryDB(ifx, "SELECT Humidity,Latitude,Longitude,Pm25,Temperature FROM sensor WHERE time > now() - 5m")
	if err == nil {
		PtRes2 = &res
	}
}

func GetSensorDev(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println(params)

	var query = map[string]string{
		"0": "SELECT * FROM sensor WHERE Id = '0' AND time > now() - 5m",
		"1": "SELECT * FROM sensor WHERE Id = '1' AND time > now() - 5m",
		"2": "SELECT * FROM sensor WHERE Id = '2' AND time > now() - 5m",
		"3": "SELECT * FROM sensor WHERE Id = '3' AND time > now() - 5m",
		"4": "SELECT * FROM sensor WHERE Id = '4' AND time > now() - 5m",
		"5": "SELECT * FROM sensor WHERE Id = '5' AND time > now() - 5m",
	}

	q, ok := query[params["device"]]
	if !ok {
		json.NewEncoder(w).Encode("")
		return
	}

	res, err := queryDB(ifx, q)
	//fmt.Println(res[0].Series[0]);
	if err != nil {
		json.NewEncoder(w).Encode("")
	} else {
		json.NewEncoder(w).Encode(res)
	}
}

func GetHistory(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	json.NewEncoder(w).Encode(&PtRes)
}
func GetIdw(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	json.NewEncoder(w).Encode(&PtRes2)
}

// InfluxDB Client Initial
func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
	checkErr(err)

	return c
}

//資料庫寫
func createMetrics(c client.Client, v Sensor) {

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "m",
	})

	if err != nil {
		checkErr(err)
	}

	ids := strconv.Itoa(int(v.Id))
	tags := map[string]string{
		"Id":        ids,
		"Device_id": v.Device_id,
	}
	fields := map[string]interface{}{
		"App":         v.App,
		"Device":      v.Device,
		"SiteName":    v.SiteName,
		"Longitude":   v.Longitude,
		"Latitude":    v.Latitude,
		"Speed_kmph":  v.Speed_kmph,
		"Timestamp":   v.Timestamp,
		"Temperature": v.Temperature,
		"Barometer":   v.Barometer,
		"Pm1":         v.Pm1,
		"Pm25":        v.Pm25,
		"Pm10":        v.Pm10,
		"Humidity":    v.Humidity,
		"Satellites":  v.Satellites,
		"Voltage":     v.Voltage,
	}
	//udbtime, _ := time.Parse(time.RFC3339,strconv.Itoa(int(v.Timestamp)))
	//	pt, err := client.NewPoint("sensor", tags, fields, time.Now())
	pt, err := client.NewPoint("sensor", tags, fields, v.Timestamp)
	if err != nil {
		checkErr(err)
	}

	bp.AddPoint(pt)

	err = c.Write(bp)
	if err != nil {
		checkErr(err)
	}
}

//資料庫讀
func queryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

//處理websocket
var PtDev string
var PtRes *[]client.Result
var PtRes2 *[]client.Result

var database, username, password string
var ifx client.Client

//main function
func main() {

	database = "test1"
	username = "u9923001"
	password = "zephyrus2"

	myRouter := mux.NewRouter()
	sess := NewSession()

	//定時抓資料
	lcache := NewLassCache()
	go GetLassData(sess, lcache)

	ifx = influxDBClient()
	//LASS感測器資料
	myRouter.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
		//ReadBufferSize:  1024,
		//WriteBufferSize: 1024,
		}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		list := lcache.GetAll()
		fmt.Println("[ws]", len(list))
		for _, buf := range list {
			conn.WriteMessage(1, buf)
		}

		wsconn := sess.Add(conn)
		defer sess.Del(wsconn)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			mlen := len(message)
			com := message[0]
			sb := message[2:mlen]
			//fmt.Println(message)
			switch com {
			case '0': //history
				str := string(sb)
				res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Device_id = '"+str+"' AND time > now() - 1d")
				if err != nil {

				} else {
					PtRes = &res
					_ = conn.WriteMessage(messageType, message)
				}
			}

		}
	})
	//myRouter.HandleFunc("/sensor/{device}", GetSensorDev).Methods("GET")
	myRouter.HandleFunc("/sensor/history/{device}", GetHistory).Methods("GET")
	myRouter.HandleFunc("/sensor/idw/", GetIdw).Methods("GET")
	myRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//定時算IDW
	go setIDW()
	fmt.Println("Listen Port : *3001")

	//--HTTP
	http.ListenAndServe(":3001", myRouter)

	//--HTTPS-1
	//http.ListenAndServeTLS(":3001", "server.crt", "server.key", myRouter)

	//--HTTPS-2
	/*	certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("u9923001.myddns.me"),
			Cache:      autocert.DirCache("certs"),
		}
		htServer := &http.Server{
			Addr:    ":3001",
			Handler: myRouter,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}
		htServer.ListenAndServeTLS("", "")*/
}
