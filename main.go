package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var barkKey = os.Getenv("BARK_KEY")
var airKey = os.Getenv("AIR_KEY")

// Response 定义结构体（与 API 返回的 JSON 数据一致）
type Response struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	AQI          int           `json:"aqi"`
	Idx          int           `json:"idx"`
	Attributions []Attribution `json:"attributions"`
	City         City          `json:"city"`
	Dominentpol  string        `json:"dominentpol"`
	IAQI         IAQI          `json:"iaqi"`
	Time         Time          `json:"time"`
	Forecast     Forecast      `json:"forecast"`
	Debug        Debug         `json:"debug"`
}

type Attribution struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type City struct {
	Geo      []float64 `json:"geo"`
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Location string    `json:"location"`
}

type IAQI struct {
	CO   Value `json:"co"`
	H    Value `json:"h"`
	NO2  Value `json:"no2"`
	O3   Value `json:"o3"`
	P    Value `json:"p"`
	PM10 Value `json:"pm10"`
	PM25 Value `json:"pm25"`
	SO2  Value `json:"so2"`
	T    Value `json:"t"`
	W    Value `json:"w"`
}

type Value struct {
	V float64 `json:"v"`
}

type Time struct {
	S   string `json:"s"`
	TZ  string `json:"tz"`
	V   int64  `json:"v"`
	ISO string `json:"iso"`
}

type Forecast struct {
	Daily Daily `json:"daily"`
}

type Daily struct {
	O3   []Pollutant `json:"o3"`
	PM10 []Pollutant `json:"pm10"`
	PM25 []Pollutant `json:"pm25"`
	UVI  []Pollutant `json:"uvi"`
}

type Pollutant struct {
	Avg int    `json:"avg"`
	Day string `json:"day"`
	Max int    `json:"max"`
	Min int    `json:"min"`
}

type Debug struct {
	Sync string `json:"sync"`
}

// pushAnnounce 使用 Bark API 推送通知
func pushAnnounce(title string, message string) {
	barkApi := "https://api.day.app/" + barkKey

	data := map[string]interface{}{
		"body":  message,
		"title": title,
		"badge": 1,
		"sound": "minuet",
		"icon":  "https://aqicn.org/images/logo/regular.png",
		"group": "Weather",
		"url":   "https://air.juniortree.com",
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post(barkApi, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("推送通知失败:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	fmt.Println("通知状态码:", resp.StatusCode)
}

// getData 从 API 获取 JSON 数据，并解析出 AQI 和 IAQI 信息
func getData() (int, IAQI) {
	// 使用实际的 API 地址，此处的 token 使用 airKey 环境变量
	airApi := "https://api.waqi.info/feed/@1451/?token=" + airKey
	resp, err := http.Get(airApi)
	if err != nil {
		fmt.Println("请求失败:", err)
		return 0, IAQI{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return 0, IAQI{}
	}

	var apiResp Response
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		fmt.Println("JSON 解析失败:", err)
		return 0, IAQI{}
	}

	if apiResp.Status != "ok" {
		fmt.Println("API 返回状态错误:", apiResp.Status)
		return 0, IAQI{}
	}

	return apiResp.Data.AQI, apiResp.Data.IAQI
}

// analyze 根据 AQI 值分析空气质量级别
func analyze(aqi int) int {
	if aqi <= 50 {
		return 1
	} else if aqi <= 100 {
		return 2
	} else if aqi <= 150 {
		return 3
	} else if aqi <= 200 {
		return 4
	} else if aqi <= 250 {
		return 5
	} else {
		return 6
	}
}

func main() {
	if airKey == "" {
		fmt.Println("请配置API Key")
	} else if barkKey == "" {
		fmt.Println("请配置Bark Key")
	} else {
		for {
			AQI, _ := getData()
			quality := analyze(AQI)
			if quality == 1 {
				fmt.Println("空气质量正常")
				fmt.Printf("AQI: %d, 质量级别: %d\n", AQI, quality)
			} else if quality >= 2 && quality <= 4 {
				pushAnnounce("健康提醒", "当前空气质量为轻度或中度污染，出门请佩戴口罩，当前AQI为："+strconv.Itoa(AQI))
			} else if quality >= 5 {
				pushAnnounce("出行建议", "当前空气质量为重度或严重污染，建议尽量不要外出，当前AQI为："+strconv.Itoa(AQI))
			}
			time.Sleep(1 * time.Second)
			fmt.Println("开始下一次检查")
		}
	}
}
