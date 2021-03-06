package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ua               = "Mozilla/5.0 (Linux; Android 13; Pixel 6 Build/HWA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Mobile Safari/537.36"
	SecondsPerMinute = 60
	SecondsPerHour   = SecondsPerMinute * 60
	SecondsPerDay    = SecondsPerHour * 24
)

var (
	config                                   = &Config{}
	client                                   = resty.New()
	login                                    = resty.New()
	cookies                                  []*http.Cookie
	orderId                                  string
	itemName                                 string
	strStartTime                             string
	oauthKey                                 string
	qrCodeUrl                                string
	fileName                                 string
	startTime, waitTime, errorTime, fastTime int64
	bp, price                                float64
	rankInfo                                 *Rank
)

type GetLoginUrl struct {
	Code   int  `json:"code"`
	Status bool `json:"status"`
	Ts     int  `json:"ts"`
	Data   struct {
		Url      string `json:"url"`
		OauthKey string `json:"oauthKey"`
	} `json:"data"`
}

type GetLoginInfo struct {
	Status bool `json:"status"`
	//Data    string `json:"data"`
	Message string `json:"message"`
}

type Config struct {
	BpEnough    bool   `json:"bp_enough"`
	BuyNum      string `json:"buy_num"`
	CouponToken string `json:"coupon_token"`
	Device      string `json:"device"`
	ItemId      string `json:"item_id"`
	TimeBefore  int    `json:"time_before"`
	Cookies     struct {
		SESSDATA        string `json:"SESSDATA"`
		BiliJct         string `json:"bili_jct"`
		DedeUserID      string `json:"DedeUserID"`
		DedeUserIDCkMd5 string `json:"DedeUserID__ckMd5"`
	} `json:"cookies"`
}

//type Static struct {
//	AppId    int    `json:"appId"`
//	Platform int    `json:"platform"`
//	Version  string `json:"version"`
//	Abtest   string `json:"abtest"`
//}

type Details struct {
	Data struct {
		Name       string `json:"name"`
		Properties struct {
			SaleTimeBegin    string `json:"sale_time_begin"`
			SaleBpForeverRaw string `json:"sale_bp_forever_raw"`
		}
		CurrentActivity struct {
			PriceBpForever float64 `json:"price_bp_forever"`
		} `json:"current_activity"`
	} `json:"data"`
}

type Now struct {
	Data struct {
		Now int64 `json:"now"`
	} `json:"data"`
}

type Navs struct {
	Code int `json:"code"`
	Data struct {
		Wallet struct {
			BcoinBalance float64 `json:"bcoin_balance"`
		} `json:"wallet"`
		Uname string `json:"uname"`
	} `json:"data"`
}

type Asset struct {
	Data struct {
		Id   int `json:"id"`
		Item struct {
			ItemId int `json:"item_id"`
		} `json:"item"`
	} `json:"data"`
}

type Rank struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Rank []struct {
			Mid      int    `json:"mid"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
			Number   int    `json:"number"`
		} `json:"rank"`
	} `json:"data"`
}

type Create struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		OrderId  string `json:"order_id"`
		State    string `json:"state"`
		BpEnough int    `json:"bp_enough"`
	} `json:"data"`
}

type Query struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		OrderId  string `json:"order_id"`
		Mid      int    `json:"mid"`
		Platform string `json:"platform"`
		ItemId   int    `json:"item_id"`
		PayId    string `json:"pay_id"`
		State    string `json:"state"`
	} `json:"data"`
}

type Wallet struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		BcoinBalance  float64 `json:"bcoin_balance"`
		CouponBalance int     `json:"coupon_balance"`
	} `json:"data"`
}

type SuitAsset struct {
	Data struct {
		Fan struct {
			IsFan      bool   `json:"is_fan"`
			Token      string `json:"token"`
			Number     int    `json:"number"`
			Color      string `json:"color"`
			Name       string `json:"name"`
			LuckItemId int    `json:"luck_item_id"`
			Date       string `json:"date"`
		} `json:"fan"`
	} `json:"data"`
}

// ????????????
func webLogin() {
	var mode int
	log.Println("??????????????? SESSDATA, ???????????????????????????????????????????????? (????????????).")
	fmt.Println("\t1. ????????????????????????.")
	fmt.Println("\t2. ????????????????????????????????????.")
	fmt.Println("\t3. APP ?????? URL ??????.")

Loop:
	_, err := fmt.Scanf("%v", &mode)
	checkErr(err)

	switch mode {
	case 1:
		getLoginUrl()
		obj := qrcodeTerminal.New()
		obj.Get(qrCodeUrl).Print()
	case 2:
		getLoginUrl()
		err = qrcode.WriteFile(qrCodeUrl, qrcode.Medium, 256, "./login.png")
		log.Println("????????????????????????????????????????????????.")
		checkErr(err)
	case 3:
		getLoginUrl()
		log.Println("????????? URL ??? APP ?????????:")
		fmt.Println(qrCodeUrl)
	default:
		fmt.Printf("???????????????: ")
		goto Loop
	}
	getLoginInfo()
}

// ?????????????????????token
func getLoginUrl() {
	g := &GetLoginUrl{}
	_, err := login.R().
		SetResult(g).
		Get("/qrcode/getLoginUrl")

	checkErr(err)
	qrCodeUrl = g.Data.Url
	oauthKey = g.Data.OauthKey
	//fmt.Println(r)
	//fmt.Println(qrCodeUrl)
	//fmt.Println(oauthKey)
}

// ?????????????????????
func getLoginInfo() {
	for {
		task := time.NewTimer(3 * time.Second)
		data := map[string]string{
			"oauthKey": oauthKey,
		}

		g := &GetLoginInfo{}
		r, err := login.R().
			SetFormData(data).
			SetResult(g).
			Post("/qrcode/getLoginInfo")

		checkErr(err)
		//fmt.Println(g)

		if g.Status == true {
			cookies = r.Cookies()
			for _, cookie := range cookies {
				switch cookie.Name {
				case "SESSDATA":
					config.Cookies.SESSDATA = cookie.Value
				case "bili_jct":
					config.Cookies.BiliJct = cookie.Value
				case "DedeUserID":
					config.Cookies.DedeUserID = cookie.Value
				case "DedeUserID__ckMd5":
					config.Cookies.DedeUserIDCkMd5 = cookie.Value
				}
			}

			result, err := json.MarshalIndent(config, "", " ")
			checkErr(err)

			err = ioutil.WriteFile(fileName, result, 644)
			checkErr(err)

			break
		}
		<-task.C
	}

}

func nav() {
	params := map[string]string{
		"csrf": config.Cookies.BiliJct,
	}

	navs := &Navs{}
	_, err := client.R().
		SetResult(navs).
		SetQueryParams(params).
		Get("/web-interface/nav")
	checkErr(err)
	if navs.Code == -101 {
		log.Fatalln("???????????????????????????cookies.")
	}
	bp = navs.Data.Wallet.BcoinBalance
	uname := navs.Data.Uname
	log.Printf("????????????, ????????????: %v, B????????????: %v.", uname, bp)
}

func popup() {
	params := map[string]string{
		"csrf": config.Cookies.BiliJct,
	}

	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/popup")
	checkErr(err)
	//fmt.Println(r)
}

func detail() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}

	details := &Details{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(details).
		Get("/garb/v2/mall/suit/detail")
	checkErr(err)
	itemName = details.Data.Name
	strStartTime = details.Data.Properties.SaleTimeBegin
	startTime, err = strconv.ParseInt(strStartTime, 10, 64)
	checkErr(err)
	if details.Data.CurrentActivity.PriceBpForever == 0 {
		p, _ := strconv.ParseFloat(details.Data.Properties.SaleBpForeverRaw, 64)
		price = p / 100
	} else {
		price = details.Data.CurrentActivity.PriceBpForever / 100
	}
	log.Printf("????????????: %v???????????????: %v.", details.Data.Name, startTime)
	if config.BpEnough == true {
		if price > bp {
			log.Fatalf("???????????????????????????????????????????????? %.2f B???.", price)
		}
	} else if config.BpEnough == false {
		if price > bp {
			log.Printf("???????????????????????????????????????????????? %.2f B???.\n", price)
		}
	}
}

func asset() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}

	assetData := &Asset{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(assetData).
		Get("/garb/user/asset")
	checkErr(err)
	//fmt.Println(r)
}

func state() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}

	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/user/reserve/state")
	checkErr(err)
}

func rank() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}

	ranks := &Rank{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(ranks).
		Get("/garb/rank/fan/recent")
	checkErr(err)
	rankInfo = ranks
	//fmt.Println(r)
}

func stat() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}

	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/order/user/stat")
	checkErr(err)
	//fmt.Println(r)
}

func coupon() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}

	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/coupon/usable")
	checkErr(err)
	//fmt.Println(r)
}

func create() {
Loop:
	for {
		// 1s ????????????
		task := time.NewTimer(1 * time.Second)
		//params := map[string]string{
		//	"add_month":    "-1",
		//	"buy_num":      config.BuyNum,
		//	"coupon_token": "",
		//	"csrf":         config.Cookies.BiliJct,
		//	"currency":     "bp",
		//	"item_id":      config.ItemId,
		//	"platform":     config.Device,
		//}

		data := map[string]string{
			"add_month":    "-1",
			"buy_num":      config.BuyNum,
			"coupon_token": "",
			"csrf":         config.Cookies.BiliJct,
			"currency":     "bp",
			"item_id":      config.ItemId,
			"platform":     config.Device,
		}

		//s := sign(data)
		//data["sign"] = s

		creates := &Create{}
		r, err := client.R().
			SetFormData(data).
			//SetQueryParams(params).
			SetResult(creates).
			EnableTrace().
			Post("/garb/v2/trade/create")
		checkErr(err)
		log.Printf("??????????????????: %v.", r.Request.TraceInfo().TotalTime)
		switch creates.Code {
		case 0: // ??????????????????????????????????????????
			if creates.Data.BpEnough == -1 {
				log.Println(r)
				log.Fatalln("????????????.")
			}
			orderId = creates.Data.OrderId
			if creates.Data.State != "paying" {
				log.Println(r)
			}
			break Loop
		case -400:
			log.Fatalln(r)
		case -403: //????????????
			log.Fatalln("???????????????.")
		case 26102: //??????????????????????????????????????????????????????????????????
			errorTime += 1
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
			log.Println(r)
			task.Reset(0)
			create()
		case 26106: //????????????????????????
			log.Fatalln(r)
		case 26120: //?????????????????????????????????????????????????????????????????????
			fastTime++
			if fastTime >= 5 {
				log.Println(r)
				log.Fatalln("???????????????????????????????????????????????????????????????...")
			}
			log.Println(r)
		case 26113: //????????????
			log.Fatalln("????????????/??????/???????????????????????????????????????.")
		case 26134: //?????????????????????????????????????????????26135?????????????????????
			errorTime += 1
			if errorTime >= 5 {
				log.Println(r)
				log.Fatalln("??????????????????????????????????????????...")
			}
			log.Println(r)
			task.Reset(500 * time.Millisecond)
			create()
		case 26135: //?????????????????????????????????????????????????????????????????????
			errorTime += 1
			if errorTime >= 5 {
				log.Println(r)
				log.Fatalln("??????????????????????????????????????????...")
			}
			log.Println(r)
			task.Reset(500 * time.Millisecond)
			create()
		case 69949: //????????????????????????????????????
			errorTime += 1
			log.Println(r)
			log.Println("?????????69949.")
			go coupon()
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
		default:
			errorTime += 1
			log.Println(r)
			go coupon()
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
		}
		<-task.C
	}
}

func tradeQuery() {
Loop:
	for {
		task := time.NewTimer(500 * time.Millisecond)
		params := map[string]string{
			"csrf":     config.Cookies.BiliJct,
			"order_id": orderId,
		}
		query := &Query{}
		r, err := client.R().
			SetQueryParams(params).
			SetResult(query).
			Get("/garb/trade/query")
		checkErr(err)
		//log.Println(r)

		if query.Code == 0 {
			switch query.Data.State {
			case "paid":
				log.Println("???????????????.")
				break Loop
			case "paying":
				log.Println("?????????????????????...")
			default:
				errorTime += 1
				log.Println(r)
				if errorTime >= 5 {
					log.Fatalln("??????????????????????????????????????????...")
				}
			}
		} else {
			errorTime += 1
			log.Println(r)
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
		}
		<-task.C
	}
}

func wallet() {
	params := map[string]string{
		"platform": "android",
	}
	response := &Wallet{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(response).
		Get("/garb/user/wallet?platform")
	checkErr(err)
	log.Printf("?????????????????????: %v.", response.Data.BcoinBalance)
}

func suitAsset() {
	params := map[string]string{
		"item_id": config.ItemId,
		"part":    "suit",
		"trial":   "0",
	}
	response := &SuitAsset{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(response).
		Get("garb/user/suit/asset")
	checkErr(err)
	//fmt.Println(r)
	log.Printf("??????: %v ??????: %v.", itemName, response.Data.Fan.Number)
}

func now() {
	result := &Now{}
	clock := resty.New()
	for {
		r, err := clock.R().
			SetResult(result).
			EnableTrace().
			SetHeader("user-agent", ua).
			Get("http://api.bilibili.com/x/report/click/now")
		checkErr(err)
		if result.Data.Now >= startTime-28 {
			waitTime = r.Request.TraceInfo().ServerTime.Milliseconds()
			break
		}
	}
}

/*
func clientInfo() {
	test := resty.New()
	resp, err := test.R().
		EnableTrace().
		SetHeader("user-agent", ua).
		Get("https://api.bilibili.com/client_info")
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println("  Headers    :", resp.Header())
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr.String())
}
*/

// ??? params ????????????
func sign(params map[string]string) string {
	var query string
	var buffer bytes.Buffer

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		//query += fmt.Sprintf("%v=%v&", k, params[k])
		buffer.WriteString(k)
		buffer.WriteString("=")
		buffer.WriteString(params[k])
		buffer.WriteString("&")
	}
	query = strings.TrimRight(buffer.String(), "&")

	// ???????????????????????????????????????????????????????????????
	/*
		v := url.Values{}
		for key, value := range params {
			v.Add(key, value)
		}
		query, err := url.QueryUnescape(v.Encode())
		checkErr(err)
	*/

	s := strMd5(fmt.Sprintf("%v%v", query, "560c52ccd288fed045859ed18bffd973"))
	//fmt.Println(sign)
	return s
}

// ?????? MD5
func strMd5(str string) (retMd5 string) {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func outPutRank() {
	if rankInfo.Data.Rank == nil {
		log.Println("???????????????????????????????????????????????????")
		return
	}
	log.Println("??????????????????:")
	fmt.Println("")
	for _, x := range rankInfo.Data.Rank {
		fmt.Printf("\t??????: %v\t?????????: %v\n", x.Number, x.Nickname)
	}
	fmt.Println("")
}

func waitToStart() {
	log.Println("??????????????????...")
	for {
		task := time.NewTimer(1 * time.Millisecond)
		t := time.Now().Unix()
		fmt.Printf("?????????: %v.\r", formatSecond(startTime-t))
		if t >= startTime-30 {
			log.Println("?????????????????????")
			task.Reset(0)
			break
		}
		<-task.C
	}
}

func formatSecond(seconds int64) string {
	var d, h, m, s int64
	var msg string

	if seconds > SecondsPerDay {
		d = seconds / SecondsPerDay
		h = seconds % SecondsPerDay / SecondsPerHour
		m = seconds % SecondsPerDay % SecondsPerHour / SecondsPerMinute
		s = seconds % 60
		msg = fmt.Sprintf("%v???%v??????%v???%v???", d, h, m, s)
	} else if seconds > SecondsPerHour {
		h = seconds / SecondsPerHour
		m = seconds % SecondsPerHour / SecondsPerMinute
		s = seconds % 60
		msg = fmt.Sprintf("%v??????%v???%v???", h, m, s)
	} else if seconds > SecondsPerMinute {
		m = seconds / SecondsPerMinute
		s = seconds % 60
		msg = fmt.Sprintf("%v???%v???", m, s)
	} else {
		s = seconds
		msg = fmt.Sprintf("%v???", s)
	}
	return msg
}

func init() {
	flag.StringVar(&fileName, "c", "./config.json", "Path to config file.")
	flag.Parse()

	// ??????????????????
	jsonFile, err := ioutil.ReadFile(fileName)
	checkErr(err)
	err = json.Unmarshal(jsonFile, config)
	checkErr(err)

	// ??????
	if config.Cookies.SESSDATA == "" {
		login.SetHeader("user-agent", ua)
		login.SetBaseURL("https://passport.bilibili.com")
		webLogin()
	}

	cookies = []*http.Cookie{
		{Name: "SESSDATA", Value: config.Cookies.SESSDATA},
		{Name: "bili_jct", Value: config.Cookies.BiliJct},
		{Name: "DedeUserID", Value: config.Cookies.DedeUserID},
		{Name: "DedeUserID__ckMd5", Value: config.Cookies.DedeUserIDCkMd5},
	}

	headers := map[string]string{
		"user-agent":         ua,
		"native_api_from":    "h5",
		"refer":              "https://www.bilibili.com",
		"x-bili-aurora-zone": "",
		"bili-bridge-engine": "cronet",
	}

	client.SetHeaders(headers)
	client.SetBaseURL("https://api.bilibili.com/x")
	client.SetCookies(cookies)
}

func main() {
	// ?????????log
	f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	checkErr(err)
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)

	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	// ??????
	//clientInfo()
	//os.Exit(0)

	// ????????????
	nav()

	// ??????
	popup()

	// ??????????????????
	detail()

	// ???????????????!!
	asset()
	state()
	rank()
	stat()
	coupon()

	// ??????????????????
	outPutRank()

	// ????????????
	//startTime = time.Now().Unix() + 10

	// ??????????????????????????????????????????????????????????????????
	waitToStart()

	// ??????b????????????????????????????????????????????????
	now()

	// ?????????
	time.Sleep((27000 - time.Duration(waitTime) - time.Duration(config.TimeBefore)) * time.Millisecond)

	// ???????????????
	start := time.NewTimer(1000 * time.Millisecond)
	detail()
	go asset()
	go state()
	go rank()
	go stat()
	go coupon()
	<-start.C

	// ????????????
	create()

	// ????????????
	tradeQuery()

	// ????????????
	nav()
	wallet()

	// ????????????
	suitAsset()
}

func checkErr(err error) {
	if err != err {
		log.Fatalln(err)
	}
}
