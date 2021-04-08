package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	_"go/types"
	"log"
	"net"
	"os"
	"reflect"
	_"reflect"
	"strings"
	"sync"
	"time"
)

var (
	netName string = "api.warframe.com"
	/*	ipArr   []string = []string{"184.51.176.169", "23.74.133.167", "104.95.79.154", "104.126.22.203",
		"23.79.190.230", "23.12.224.185", "104.86.4.131", "23.33.36.79",
		"104.76.1.18", "23.198.102.81", "23.41.68.161", "96.16.180.231",
		"104.94.227.206", "104.127.99.156", "23.39.250.104", "72.247.104.185",
		"23.39.78.6", "96.17.217.36", "2.21.127.142", "23.194.74.82",
		"104.89.9.82", "23.66.16.169", "104.87.134.99", "104.74.221.247",
		"23.5.242.238", "23.4.113.204", "184.30.212.47", "23.76.83.250", "23.51.129.181"}*/
	ipArr     []string = []string{"72.247.104.185", "104.103.116.152", "104.117.201.77", "23.72.65.27", "104.99.90.187", "23.33.36.79", "184.51.176.169", "104.127.215.220", "104.126.22.203", "96.17.43.31", "104.94.227.206", "96.16.180.231", "104.121.93.36", "23.208.32.39", "104.103.226.202", "23.74.133.167", "23.211.165.59", "23.41.68.161", "96.16.137.94", "96.16.225.173", "104.86.4.131"}
	pingTimes int      = 5
	maxTime   int64    = 9999
)

type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)
}
func getICMP(seq uint16) ICMP {
	icmp := ICMP{
		Type:        8,
		Code:        0,
		CheckSum:    0,
		Identifier:  0,
		SequenceNum: seq,
	}

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.CheckSum = CheckSum(buffer.Bytes())
	buffer.Reset()
	return icmp
}
func getPingTimes(ip string, resultChan chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("ip4:icmp", ip)
	conn.SetReadDeadline((time.Now().Add(time.Second * 1)))
	var sumTime int64 = 0
	var avgTime int64 = 0
	if err != nil {
		fmt.Println(err)
		return
	}
	var realTimes = 0
	for   i := uint16(0); int(i) < pingTimes; i++ {
		icmp := getICMP(i)
		err = nil
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.BigEndian, icmp)
		if _, err := conn.Write(buffer.Bytes()); err != nil {
			fmt.Println(err)
			continue
		}
		//	buffer.Reset();
		tStart := time.Now()
		recv := make([]byte, 1024)
		receiveCnt, err := conn.Read(recv)

		if err != nil {
			fmt.Println(err)
			continue
		}
		tEnd := time.Now()
		//duration := tEnd.Sub(tStart).Nanoseconds() / 1e6
		duration := tEnd.Sub(tStart).Milliseconds()
		sumTime += duration
		realTimes++
		fmt.Printf("%d bytes from %s: seq=%d time=%d\n", receiveCnt, ip, icmp.SequenceNum, duration)
	}
	if realTimes == 0 {
		avgTime = maxTime
	} else {
		avgTime = sumTime / int64(realTimes)
	}
	if avgTime == maxTime {
		sumTime = maxTime
	}
	result := Result{
		ip:       ip,
		avgTime:  avgTime,
		sumTime:  sumTime,
		loseRate: fmt.Sprintf("%.2f  ",float32(pingTimes - realTimes) / float32(pingTimes)),
	}
	resultChan <- result
	//return avgTime,sumTime
}

func chan2Slice(boolchan chan bool) {
	for v := range resultChan {
		resultArr = append(resultArr, v)
	}
	boolchan <- true
}

type Result struct {
	ip       string
	avgTime  int64
	sumTime  int64
	loseRate string
}

func getMinResult(resultArr []Result) (int, Result) {
	var (
		index  int    = -1
		result Result = Result{
			ip:      "",
			avgTime: 9999,
			sumTime: 9999,
		}
	)
	for i, v := range resultArr {
		if result.avgTime > v.avgTime {
			index = i
			result = v
		}
	}
	return index, result
}

var resultArr = make([]Result, 0, len(ipArr))
var resultChan chan Result

func write2File(result Result, f float64, filePathPointer *string) {
	var data string = ""
	file, err := os.OpenFile(*filePathPointer, os.O_RDWR, 0777)
	if err != nil {
		log.Fatalln("errat write2file:", err)
		return
	}
	defer file.Close()
	bufScanner := bufio.NewScanner(file)

	for bufScanner.Scan() {
		line := bufScanner.Text()
		if !strings.Contains(line, netName) {
			data += line + "\n"
		}
	}
	err = bufScanner.Err()
	if err != nil {
		log.Fatalln("errat write2file:", err)
		return
	}
	data = fmt.Sprintf("%s%s  %s #写入时间 %s 延迟 %d ms 运行时间 %f s", data, result.ip, netName, time.Now().String(), result.avgTime, f)
	err = file.Truncate(0)
	if err != nil {
		log.Fatalln("errat write2file:", err)
		return
	}
	//重置游标,truncate不会重置游标位置
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatalln("errat write2file:", err)
		return
	}
	file.Write([]byte(data))
	if err != nil {
		log.Fatalln("errat write2file:", err)
		return
	}
}
func toString(a interface{}) string{
	var str string="{";
	if reflect.TypeOf(a).Kind()==reflect.Struct{
		for i:=0;i<reflect.ValueOf(a).NumField();i++{
			if reflect.ValueOf(a).Field(i).Kind()==reflect.Int64{
				key:=reflect.TypeOf(a).Field(i).Name

				data:=reflect.ValueOf(a).Field(i).Int()

				str=fmt.Sprintf("%s %s:%d ,",str,key,data)

			}else{
				key:=reflect.TypeOf(a).Field(i).Name


				data:=reflect.ValueOf(a).Field(i).String()

				str=fmt.Sprintf("%s %s:%s ,",str,key,data)
			}
		}
		str=strings.TrimRight(str,",")
		str=str+"}"
	}
	return str
}
func main() {
	filePathPointer := flag.String("filepath", "C:\\Windows\\System32\\drivers\\etc\\hosts", "desc a filepath")
	//filePathPointer := flag.String("filepath", "./lines.txt", "desc a filepath")
	flag.Parse()
	for {
		start := time.Now()
		var wg sync.WaitGroup
		done := make(chan bool)
		resultChan = make(chan Result, len(ipArr))
		for _, v := range ipArr {
			wg.Add(1)
			go getPingTimes(v, resultChan, &wg)
		}
		go chan2Slice(done)
		wg.Wait()
		close(resultChan)
		<-done
		_, resutl := getMinResult(resultArr)
		fmt.Println(toString(resutl))
		fmt.Println(resultArr)
		end := time.Now()
		ms := (end.Sub(start).Milliseconds())
		s := (end.Sub(start).Seconds())
		write2File(resutl, s, filePathPointer)
		fmt.Printf("总共耗时 %d ms (%f s) \n", ms, s)

		time.Sleep(time.Second * 40)
	}
}
