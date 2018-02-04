package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

var config struct {
	Server string `json:"server"`
	Ca     string `json:"ca"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	Proto  string `json:"proto"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
}

var data = url.Values{}

func main() {
	p1, _ := exec.LookPath(os.Args[0])
	p2, _ := filepath.Abs(p1)
	p3, _ := filepath.Split(p2)
	configFile := filepath.Join(p3, "tinyfw.json")

	var jsonDecoder *json.Decoder
	configFileHandler, err := os.Open(configFile)
	if err == nil {
		jsonDecoder = json.NewDecoder(configFileHandler)
		if err := jsonDecoder.Decode(&config); err != nil {
			log.Fatal(err)
		}
		configFileHandler.Close()
	}

	server := flag.String("server", "", "tinyfw server,need to be url like https://192.168.1.1:8443")
	cafile := flag.String("ca", "", "CA")
	certfile := flag.String("cert", "", "client cert")
	keyfile := flag.String("key", "", "client key")
	add := flag.Bool("add", false, "add a ACCEPT INPUT rule")
	del := flag.Bool("del", false, "delete a ACCEPT INPUT rule")
	list := flag.Bool("list", false, "list ACCEPT INPUT rules")
	proto := flag.String("proto", "", "protocol")
	ip := flag.String("ip", "", "allowed source ip")
	port := flag.String("port", "", "allowed port")
	flag.Parse()

	if len(*server) > 0 {
		config.Server = *server
	}
	if len(*cafile) > 0 {
		config.Ca = *cafile
	}
	if len(*certfile) > 0 {
		config.Cert = *certfile
	}
	if len(*keyfile) > 0 {
		config.Key = *keyfile
	}
	if len(*proto) > 0 {
		config.Proto = *proto
	}
	if len(*ip) > 0 {
		config.IP = *ip
	}
	if len(*port) > 0 {
		config.Port = *port
	}

	switch true {
	case *add:
		data.Set("act", "add")
		break
	case *del:
		data.Set("act", "del")
		break
	case *list:
		data.Set("act", "list")
	}
	data.Set("ip", config.IP)
	data.Set("proto", config.Proto)
	data.Set("port", config.Port)

	//load CA
	caPEM, err := ioutil.ReadFile(config.Ca)
	if err != nil {
		log.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)

	clientCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
	if err != nil {
		log.Fatal(err)
	}

	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{clientCert},
		},
	}

	httpClient := &http.Client{Transport: httpTransport}
	response, err := httpClient.PostForm(config.Server, data)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("%s\r\n", body)
}
