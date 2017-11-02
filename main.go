package main

import (
	"encoding/json"
	"fmt"
	"time"
	//"log"
	"net/http"
	"strconv"
	//"strings"
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/influxdata/influxdb/client/v2"
	"golang.org/x/crypto/acme/autocert"
)

// Sensor struct
type Sensor struct {
	Id          uint8
	Device      string
	App         string
	Device_id   string
	SiteName    string
	Latitude    float32
	Longitude   float32
	Timestamp   time.Time
	Speed_kmph  float32
	Temperature float32
	Barometer   float32
	Pm1         float32
	Pm25        float32
	Pm10        float32
	Humidity    float32
	Satellites  float32
	Voltage     float32
}

type Lass struct {
	Feeds []LassSensor `json:"feeds"`
}

type LassSensor struct {
	SiteName    string    `json:"SiteName,omitempty"`
	App         string    `json:"app,omitempty"`
	Device      string    `json:"device,omitempty"`
	Device_id   string    `json:"device_id,omitempty"`
	Longitude   float32   `json:"gps_lon,omitempty"`     //緯度
	Latitude    float32   `json:"gps_lat,omitempty"`     //經度
	Timestamp   time.Time `json:"timestamp,omitempty"`   //時間
	Temperature float32   `json:"Temperature,omitempty"` //溫度
	S_t0        float32   `json:"s_t0,omitempty"`        //溫度
	S_t2        float32   `json:"s_t2,omitempty"`        //溫度
	S_t4        float32   `json:"s_t4,omitempty"`        //溫度
	S_b2        float32   `json:"s_b2,omitempty"`        //氣壓
	S_b0        float32   `json:"s_b0,omitempty"`        //氣壓
	Pm1         float32   `json:"s_d2,omitempty"`        //PM1
	Pm10        float32   `json:"s_d1,omitempty"`        //PM10
	S_d0        float32   `json:"s_d0,omitempty"`        //PM25
	Pm25        float32   `json:"PM25,omitempty"`        //PM25
	Humidity    float32   `json:"Humidity,omitempty"`    //濕度
	S_h0        float32   `json:"s_h0,omitempty"`        //濕度
	S_h2        float32   `json:"s_h2,omitempty"`        //濕度
	S_h4        float32   `json:"s_h4,omitempty"`        //濕度
	Satellites  float32   `json:"gps_num,omitempty"`     //衛星
	Voltage     float32   `json:"s_1,omitempty"`         //電量
}

func (v *LassSensor) GetTemp() float32 {
	if v.Temperature > 0 {
		return v.Temperature
	} else if v.S_t0 > 0 {
		return v.S_t0
	} else if v.S_t2 > 0 {
		return v.S_t2
	} else if v.S_t4 > 0 {
		return v.S_t4
	} else {
		return 0.0
	}
}

func (v *LassSensor) GetBaro() float32 {
	if v.S_b2 > 0 {
		return v.S_b2
	} else if v.S_b0 > 0 {
		return v.S_b0
	} else {
		return 0.0
	}
}

func (v *LassSensor) Getpm1() float32 {
	if v.Pm1 > 0 {
		return v.Pm1
	} else {
		return 0.0
	}
}

func (v *LassSensor) Getpm10() float32 {
	if v.Pm10 > 0 {
		return v.Pm10
	} else {
		return 0.0
	}
}

func (v *LassSensor) Getpm25() float32 {
	if v.S_d0 > 0 {
		return v.S_d0
	} else if v.Pm25 > 0 {
		return v.Pm25
	} else {
		return 0.0
	}
}

func (v *LassSensor) GetHum() float32 {
	if v.Humidity > 0 {
		return v.Humidity
	} else if v.S_h0 > 0 {
		return v.S_h0
	} else if v.S_h2 > 0 {
		return v.S_h2
	} else if v.S_h4 > 0 {
		return v.S_h4
	} else {
		return 0.0
	}
}

func (v *LassSensor) GetSate() float32 {
	if v.Satellites > 0 {
		return v.Satellites
	} else {
		return 0.0
	}
}

func (v *LassSensor) GetVol() float32 {
	if v.Voltage > 0 {
		return v.Voltage
	} else {
		return 0.0
	}
}

// Check Error
func checkErr(e error) {
	if e != nil {
		//fmt.Println(e)
		panic(e)
	}
}

func getJson(url string, id uint8) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("xxxxxxxxxxxxxxx\r\n")
		fmt.Println(err)
		fmt.Printf("xxxxxxxxxxxxxxx\r\n")
		return
	}
	defer r.Body.Close()

	var t Lass
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Printf("xxxxxxxxxxxxxxx\r\n")
		fmt.Println(err)
		fmt.Printf("xxxxxxxxxxxxxxx\r\n")
	}

	m := t.Feeds
	for _, v := range m {
		var res Sensor
		res.Id = id
		res.SiteName = v.SiteName
		res.App = v.App
		res.Device = v.Device
		res.Device_id = v.Device_id
		res.Longitude = v.Longitude
		res.Latitude = v.Latitude
		res.Timestamp = v.Timestamp
		res.Temperature = v.GetTemp()
		res.Barometer = v.GetBaro()
		res.Pm1 = v.Getpm1()
		res.Pm25 = v.Getpm25()
		res.Pm10 = v.Getpm10()
		res.Humidity = v.GetHum()
		res.Satellites = v.GetSate()
		res.Voltage = v.GetVol()
		createMetrics(ifx, res)
	}

}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func GetLassMux(url map[int]string) {
	///*
	getJson(url[0], 0)
	getJson(url[1], 1)
	getJson(url[2], 2)
	getJson(url[3], 3)
	getJson(url[4], 4)
	getJson(url[5], 5)
	//*/
	//PrettyPrint(LassData[4]);
	//PrettyPrint(LassData[3]);
	loc, _ := time.LoadLocation("Asia/Taipei")
	now := time.Now().In(loc)
	fmt.Printf("============\r\n")
	fmt.Printf("Get %d:%d:%d\r\n", now.Hour(), now.Minute(), now.Second())
	fmt.Printf("============\r\n")
}

//var LassData [6]interface{}
func GetLassData() {

	var lassUrl = map[int]string{
		0: "https://pm25.lass-net.org/data/last-all-airbox.json",
		1: "https://pm25.lass-net.org/data/last-all-maps.json",
		2: "https://pm25.lass-net.org/data/last-all-lass.json",
		3: "https://pm25.lass-net.org/data/last-all-lass4u.json",
		4: "https://pm25.lass-net.org/data/last-all-indie.json",
		5: "https://pm25.lass-net.org/data/last-all-probecube.json",
	}

	GetLassMux(lassUrl)
	go setIDW()
	t1 := time.NewTicker(time.Second * 300)
	for _ = range t1.C {
		GetLassMux(lassUrl)
	}
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
	switch params["device"] {
	case "0":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '0' AND time > now() - 5m")
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
	case "1":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '1' AND time > now() - 5m")
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
	case "2":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '2' AND time > now() - 5m")
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
	case "3":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '3' AND time > now() - 5m")
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
	case "4":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '4' AND time > now() - 5m")
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
	case "5":
		res, err := queryDB(ifx, "SELECT * FROM sensor WHERE Id = '5' AND time > now() - 5m")
		//fmt.Println(res[0].Series[0]);
		if err != nil {
			json.NewEncoder(w).Encode("")
		} else {
			json.NewEncoder(w).Encode(res)
		}
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
	pt, err := client.NewPoint("sensor", tags, fields, time.Now())
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

func echo(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
	//ReadBufferSize:  1024,
	//WriteBufferSize: 1024,
	}
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

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
		case '1': //IWD

			_ = conn.WriteMessage(messageType, message)

		case '5':
			_ = conn.WriteMessage(messageType, message)
		}

	}
}

var database, username, password string
var ifx client.Client

//main function
func main() {

	database = "test1"
	username = "u9923001"
	password = "zephyrus2"

	myRouter := mux.NewRouter()

	ifx = influxDBClient()
	//LASS感測器資料
	myRouter.HandleFunc("/socket", echo)
	myRouter.HandleFunc("/sensor/{device}", GetSensorDev).Methods("GET")
	myRouter.HandleFunc("/sensor/history/{device}", GetHistory).Methods("GET")
	myRouter.HandleFunc("/sensor/idw/", GetIdw).Methods("GET")
	myRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//定時抓資料
	go GetLassData()
	//定時算IDW
	fmt.Println("Listen Port : *3001")

	//--HTTP
	//http.ListenAndServe(":3001", myRouter)

	//--HTTPS-1
	//http.ListenAndServeTLS(":3001", "server.crt", "server.key", myRouter)

	//--HTTPS-2
	certManager := autocert.Manager{
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
	htServer.ListenAndServeTLS("", "")
}
