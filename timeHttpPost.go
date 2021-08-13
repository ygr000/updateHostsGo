package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)
func httpPostUrl(url string,index int){
	begin:=time.Now();
	urlArr:=strings.Split(url,"####")

	response, err := http.Post(urlArr[0],"application/x-www-form-urlencoded",strings.NewReader(urlArr[1]))
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Fatalln(err)
		return
	}
	end:=time.Now();
	log.Println("第 {",index,"} 次访问",url,"耗时:",end.Sub(begin).Milliseconds(),"返回结果:",string(body))
}
func httpPostUrlWithAgent(url string,userAgent string,index int){
	begin:=time.Now();
	urlArr:=strings.Split(url,"####")
	client := &http.Client{}
	req, err := http.NewRequest("POST", urlArr[0], strings.NewReader(urlArr[1]))
	if err != nil {
		log.Fatalln(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("sec-ch-ua", "'Chromium';v='92', ' Not A;Brand';v='99', 'Google Chrome';v='92'")

	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Fatalln(err)
		return
	}
	end:=time.Now();
	log.Println("第 {",index,"} 次访问",url,"耗时:",end.Sub(begin).Milliseconds(),"返回结果:",string(body))
}
func main() {
	filePathPoint:=flag.String("fpath","d:\\urlPostPath.txt","desc: url in file ")
	flag.Parse()
	file,err:=os.OpenFile(*filePathPoint,os.O_RDWR,0777)
	if err != nil {
		log.Fatalln(err)
		return
	}
	userAgentFile,err:=os.OpenFile("d:\\userAgent.txt",os.O_RDWR,0777)
	defer file.Close();
	data,err:=ioutil.ReadAll(file)
	userAgentData,err:=ioutil.ReadAll(userAgentFile)
	if err != nil {
		log.Fatalln(err)
		return
	}
	userAgents:=strings.Split(string(userAgentData),"\r\n")
	urls:=strings.Split(string(data),"\r\n")
	index:=1;
	for {

		for _,url:=range urls{
			for _,userAgent:=range userAgents{
				go httpPostUrlWithAgent(url,userAgent,index);
			}
		}
		index++
		time.Sleep(5.1*60*time.Second)
	}

}
