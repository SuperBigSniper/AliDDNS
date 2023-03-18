package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var commandModel CommandModel
var configModel ConfigurationModel

// 上一次的公有ip
var lastPublicIp string

func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	log.Println(addrs)
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		ipv6 := regexp.MustCompile(`(\w+:){7}\w+`).FindString(addr.String())
		if strings.Count(ipv6, ":") == 7 {
			return ipv6
		}
	}
	return ""
}

func main() {

	initCommandModel()
	loadConfig()

	if commandModel.Interval == nil || *commandModel.Interval == 0 {
		update()
		return
	}

	intervalFunction()
}

func update() {
	publicIp := getPublicIp()
	log.Println("公网ip: " + publicIp)
	// log.Println("本地获取的公网Ip: ", getLocalIp())

	if lastPublicIp == publicIp {
		log.Println("Ip地址没有发生改变, 不进行更新")
		return
	}
	subDomains := getSubDomains()
	for _, sub := range subDomains {
		if sub.Value != publicIp {
			// 更新域名绑定的 IP 地址。
			sub.Value = publicIp
			sub.TTL = linq.From(*configModel.SubDomains).FirstWith(func(subDomain interface{}) bool {
				return subDomain.(SubDomainModel).Name == sub.RR
			}).(SubDomainModel).Interval
			updateSubDomain(&sub)
		}
	}

	lastPublicIp = publicIp

	log.Printf("域名记录更新成功...")
}

func intervalFunction() {
	tick := time.Tick(time.Second * time.Duration(*commandModel.Interval))
	for {
		<-tick
		update()
	}
}

func initCommandModel() {
	commandModel.FilePath = flag.String("f", "", "指定自定义的配置文件，请传入配置文件的路径。")
	commandModel.Interval = flag.Int("i", 0, "指定程序的自动检测周期，单位是秒。")

	flag.Parse()
}

func loadConfig() {
	var configFile string
	if *commandModel.FilePath == "" {
		dir, _ := os.Getwd()
		configFile = path.Join(dir, "settings.json")
	} else {
		configFile = *commandModel.FilePath
	}

	// 打开配置文件，并进行反序列化。
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("无法打开文件：%s", err)
		os.Exit(-1)
	}
	defer f.Close()
	data, _ := ioutil.ReadAll(f)

	if err := json.Unmarshal(data, &configModel); err != nil {
		log.Fatalf("数据反序列化失败：%s", err)
		os.Exit(-1)
	}
}

func getPublicIp() string {
	publicUrl := GetPublicIpUrl
	if configModel.GetPublicIpUrl != "" {
		publicUrl = configModel.GetPublicIpUrl
	}
	resp, err := http.Get(publicUrl)
	if err != nil {
		log.Fatalf("获取公网 IP 出现错误，错误信息：%s", err)
		os.Exit(-1)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	return strings.Replace(string(bytes), "\n", "", -1)
}

func getSubDomains() []alidns.Record {

	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", configModel.AccessId, configModel.AccessKey)
	if err != nil {
		log.Println(err.Error())
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"

	request.DomainName = configModel.MainDomain

	if err != nil {
		log.Fatalln("连接阿里云失败, 请检查你的AccessId和AccessKey是否正确", err)
	}

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		log.Println(err.Error())
	}

	// 过滤符合条件的子域名信息。
	var queryResult []alidns.Record
	linq.From(response.DomainRecords.Record).Where(func(c interface{}) bool {
		return linq.From(*configModel.SubDomains).Select(func(x interface{}) interface{} {
			return x.(SubDomainModel).Name
		}).Contains(c.(alidns.Record).RR)
	}).ToSlice(&queryResult)

	return queryResult
}

func updateSubDomain(subDomain *alidns.Record) {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = subDomain.RecordId
	request.RR = subDomain.RR
	request.Type = subDomain.Type
	request.Value = subDomain.Value
	request.TTL = requests.NewInteger64(subDomain.TTL)

	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", configModel.AccessId, configModel.AccessKey)
	if err != nil {
		log.Fatalln("连接阿里云失败, 请检查你的AccessId和AccessKey是否正确", err)
	}

	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		log.Print(err.Error())
	}
}
