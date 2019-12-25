package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

type ipSoruce struct {
	IP                string  `json:"ip"`
	CountryCode       string  `json:"country_code"`
	CountryName       string  `json:"country_name"`
	RegionCode        string  `json:"region_code"`
	RegionName        string  `json:"region_name"`
	City              string  `json:"city"`
	ZipCode           string  `json:"zip_code"`
	TimeZone          string  `json:"time_zone"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	MetroCode         int     `json:"metro_code"`
	SuspiciousFactors struct {
		IsProxy      bool `json:"is_proxy"`
		IsTorNode    bool `json:"is_tor_node"`
		IsSpam       bool `json:"is_spam"`
		IsSuspicious bool `json:"is_suspicious"`
	} `json:"suspicious_factors"`
}

type myResponse struct {
	Args struct {
	} `json:"args"`
	Headers struct {
		Accept                  string `json:"Accept"`
		AcceptEncoding          string `json:"Accept-Encoding"`
		AcceptLanguage          string `json:"Accept-Language"`
		CacheControl            string `json:"Cache-Control"`
		Connection              string `json:"Connection"`
		Host                    string `json:"Host"`
		IfModifiedSince         string `json:"If-Modified-Since"`
		UpgradeInsecureRequests string `json:"Upgrade-Insecure-Requests"`
		UserAgent               string `json:"User-Agent"`
	} `json:"headers"`
	Origin string `json:"origin"`
	URL    string `json:"url"`
}

func getInfo(ip string) ipSoruce {
	var ips ipSoruce
	rs, err := http.Get("https://ip-api.io/json/" + ip)
	if err != nil {
		panic(err)
	}
	defer rs.Body.Close()
	ipInfo, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		panic(err)
	} else {
		json.Unmarshal(ipInfo, &ips)
		return ips
	}
}

func outputProxy(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {

	var res myResponse
	color.Yellow(`
	
          ::::::::  :::::::::  :::::::::   ::::::::  :::    ::: :::   ::: 
          :+:    :+: :+:    :+: :+:    :+: :+:    :+: :+:    :+: :+:   :+: 
          +:+        +:+    +:+ +:+    +:+ +:+    +:+  +:+  +:+   +:+ +:+  
          :#:        +#++:++#+  +#++:++#:  +#+    +:+   +#++:+     +#++:   
          +#+   +#+# +#+        +#+    +#+ +#+    +#+  +#+  +#+     +#+    
          #+#    #+# #+#        #+#    #+# #+#    #+# #+#    #+#    #+#    
           ########  ###        ###    ###  ########  ###    ###    ###    
	              Gpr0xy by: Abdullah Hussam(@abdulahhusam)
	`)

	/*
	   go run gproxy.go -file mylist.txt
	   go run gproxy.go -file mylist.txt -active
	   go run gproxy.go -file mylist.txt -active -o output.txt

	*/

	filename := flag.String("file", "", "Path to prorxy list file e.g: mylist.txt.")
	output := flag.String("o", "output.txt", "Output filename e.g:output.txt.")
	showActive := flag.Bool("active", false, "Show only active porxies.")

	flag.Parse()

	if *filename == "" {
		fmt.Println()
		color.Yellow("File path is required! Use -file {{path}} e.g: -file mylist.txt")

		os.Exit(1)
	}

	color.Yellow("\tIP\t\t\t Port\t\t Country\t\t TOR\t\t Spam")

	content, err := ioutil.ReadFile(*filename)

	if err != nil {
		color.Yellow("File path is invaild!")
		os.Exit(2)
	}
	lines := strings.Split(string(content), "\r\n")
	arrayOfWorkingProxy := []string{}

	for i := 0; i < len(lines); i++ {
		proxyAddress := "http://" + lines[i]
		proxy, err := url.Parse(proxyAddress)

		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		target := "http://httpbin.org/get"
		url, err := url.Parse(target)
		if err != nil {
			color.Yellow("Unable to parse target URL!")
			i++
		}

		transportObj := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client := &http.Client{
			Transport: transportObj,
			Timeout:   time.Duration(2 * time.Second),
		}

		request, err := http.NewRequest("GET", url.String(), nil)
		if err != nil {
			color.Yellow("Unable to send request!")
			i++
		}

		response, err := client.Do(request)
		if err != nil {
			if !*showActive {
				failedIp := strings.Split(lines[i], ":")[0]
				color.Red("\t" + failedIp + "\t\t " + "N/A" + "\t\t " + "N/A" + "\t\t\t" + "N/A" + "\t\t" + "N/A" + "\n")

			}
		} else {
			arrayOfWorkingProxy = append(arrayOfWorkingProxy, lines[i])
			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println(err)
				os.Exit(7)
			}
			json.Unmarshal(data, &res)
			addr := strings.Split(lines[i], ":")
			IP := addr[0]
			info := getInfo(IP)
			Port := addr[1]
			tor := "false"
			if info.SuspiciousFactors.IsTorNode != false {
				tor = "True"
			}
			spam := "false"
			if info.SuspiciousFactors.IsTorNode != false {
				spam = "True"
			}
			color.Green("\t" + IP + "\t\t " + Port + "\t\t " + info.CountryName + "\t\t" + tor + "\t\t" + spam + "\n")

		}
	}
	if err := outputProxy(arrayOfWorkingProxy, *output); err != nil {
		log.Fatalf("outputProxy: %s", err)
	}

}
