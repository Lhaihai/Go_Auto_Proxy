package main

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)
//状态检测结果
type ipresult struct {
	ERRORCODE string `json:"ERRORCODE"`
	RESULT    []struct {
		Position string `json:"position"`
		Port     string `json:"port"`
		Time     string `json:"time"`
		Anony    string `json:"anony"`
		IP       string `json:"ip"`
	} `json:"RESULT"`
}

type StoreProxy []struct {
	Position string `json:"position"`
	Port     string `json:"port"`
	Time     string `json:"time"`
	Anony    string `json:"anony"`
	IP       string `json:"ip"`
}

type Proxy_Pool []struct {
	CheckCount int    `json:"check_count"`
	FailCount  int    `json:"fail_count"`
	LastStatus int    `json:"last_status"`
	LastTime   string `json:"last_time"`
	Proxy      string `json:"proxy"`
	Region     string `json:"region"`
	Source     string `json:"source"`
	Type       string `json:"type"`
}


func main() {

	Url,delaytime,loop := get_args()

	backup := get_DefaultConnectionSettingsValue()
	proxy := getproxy(Url)

	fmt.Println(strconv.Itoa(len(proxy))+"个代理可用")
	for j:=0 ; j < loop ;j++ {
		for i := 0; i <= len(proxy)-1; i++ {
			tmp := proxy[i].IP + ":" + proxy[i].Port
			s := check_proxy("http://www.baidu.com", "http://"+tmp)
			fmt.Println(i, s)
			if s == 200 {
				fmt.Println(tmp, proxy[i].Time, proxy[i].Anony)
				Set_Proxy_ver2(tmp)
			} else {
				continue
			}
			time.Sleep(time.Duration(delaytime) * time.Second)
		}
	}
	//恢复原来设置
	set_proxy("DefaultConnectionSettings",backup)
	set_proxy("SavedLegacySettings",backup)
	fmt.Println("恢复代理原来设置")
}

func get_args()(string,int,int){

	c:=flag.String("c","", "-c cls 重置代理设置为自动代理")
	u:=flag.String("u","", "代理 Url，例如 http://127.0.0.1:5010/get_all")
	t:=flag.Int("t",30, "自动切换代理时间间隔")
	l:=flag.Int("l",1,"循环次数")

	flag.Parse()

	if *c == "cls" {
		clear()
		fmt.Println("重置代理成功")
		os.Exit(3)
	}

	if *u == "" {
		fmt.Println("proxy.exe -u http://127.0.0.1:5010 -t 30")
		os.Exit(3)
	}
	return *u,*t,*l
}

//返回Get请求数据
func Get(url string) string {
	r, err := http.Get(url)
	if err != nil{
		panic(err)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	data := string(body)
	return data
}


//获取代理数据
func getproxy(Url string) StoreProxy {

	rooturl := "http://www.xdaili.cn/ipagent/checkIp/ipList?"

	r, err := http.Get(Url)
	if err != nil{
		fmt.Println("请检查地址是否正确")
		panic(err)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	data := string(body)

	var proxys Proxy_Pool
	_ = json.Unmarshal([]byte(data), &proxys)

	for _, r := range proxys{
		if r.FailCount == 0 {
			rooturl = rooturl+"ip_ports%5B%5D="+r.Proxy+"&"
		}
	}

	data2 := Get(rooturl)
	var f ipresult
	_ = json.Unmarshal([]byte(data2), &f)

	proxy := StoreProxy{}
	for _, r := range f.RESULT {
		if r.Anony != "\"透明\"" && r.Anony != "" && len(r.Time) < 6 {
			proxy = append(proxy,r)
		}
	}
	return proxy
}


func get_DefaultConnectionSettingsValue() string{
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings\Connections`, registry.ALL_ACCESS)
	defer key.Close()
	if err != nil{
		panic(err)
	}

	s, _, _ := key.GetBinaryValue(`DefaultConnectionSettings`)
	d := ""
	for _,x := range s{
		d = d + fmt.Sprintf("%02x",x)
	}
	return d
}

//设置代理
func Set_Proxy_ver2(ip string){
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings\Connections`, registry.ALL_ACCESS)
	defer key.Close()
	if err != nil{
		panic(err)
	}

	s, _, _ := key.GetBinaryValue(`DefaultConnectionSettings`)
	d := ""
	for _,x := range s{
		d = d + fmt.Sprintf("%02x",x)
	}

	p1 := d[:16] //460000003A160000
	switch_proxy :=  "03" // d[16:18]

	leng := fmt.Sprintf("%02x",(len(ip)))  //如果ip长度小于16，前面补0
	iphex := hex.EncodeToString([]byte(ip))

	data := p1 + switch_proxy + "000000" + leng + "000000" + iphex + "070000003c6c6f63616c3e2b000000"

	set_proxy("DefaultConnectionSettings",data)
	set_proxy("SavedLegacySettings",data)
	fmt.Println("代理设置成功")
}

//传入要修改keyname和16进制字符串
func set_proxy(keyname string,data string){
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings\Connections`, registry.ALL_ACCESS)
	defer key.Close()
	if err != nil{
		panic(err)
	}

	//把16进制字符串转为byte切片
	bytedata := []byte{}
	for i:=0 ; i<len(data)-2; i=i+2{
		t := data[i:i+2]
		n, err := strconv.ParseUint(t, 16, 32)
		if err != nil {
			panic(err)
		}
		n2 := byte(n)
		bytedata = append(bytedata,n2)
	}

	err = key.SetBinaryValue(keyname,bytedata)
	if err != nil{
		panic(err)
	}
}

func check_proxy(webUrl, proxyUrl string) int {
	proxy, _ := url.Parse(proxyUrl)
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5, //超时时间
	}
	resp, err := client.Get(webUrl)
	if err != nil {
		fmt.Println("出错了", err)
		return 500
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

//清空代理设置，设置为自动检查代理，如果不慎操作注册表可以用来恢复设置
func clear()  {
	data := "460000003016000009000000010000003a050000006c6f63616c000000000100000000000000000000000000000000000000000000000000000000000000"
	set_proxy("DefaultConnectionSettings",data)
	set_proxy("SavedLegacySettings",data)
}