package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type testHandler struct{}

var config struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Ca   string `json:"ca"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
var ipRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
var portRegExp = regexp.MustCompile(`^\d{1,5}$`)

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	port := r.FormValue("port")
	ip := r.FormValue("ip")
	proto := r.FormValue("proto")
	act := r.FormValue("act")

	switch act {
	case "add", "del":
		if !portRegExp.MatchString(port) {
			http.Error(w, "illegal port", 500)
			return
		}
		if len(ip) == 0 {
			if ipRegExp.MatchString(r.RemoteAddr) {
				ip = ipRegExp.FindString(r.RemoteAddr)
			}
		}
		if len(ip) == 0 {
			http.Error(w, "no source ip found", 500)
			return
		}
		switch proto {
		case "tcp", "udp":
		default:
			http.Error(w, "available protocol is tcp or udp", 500)
			return
		}
		proto = strings.ToLower(proto)

	}

	switch act {
	case "add":
		cmd := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-p", proto, "--dport", port, "-j", "ACCEPT")
		err := cmd.Run()
		if err != nil { //doesn't exist
			cmd = exec.Command("iptables", "-I", "INPUT", "-s", ip, "-p", proto, "--dport", port, "-j", "ACCEPT")
			cmd.Stderr = w
			err = cmd.Run()
			if err != nil {
				w.Write([]byte(strings.Join(cmd.Args, " ")))
				log.Println(err)
			}
		}
	case "del":
		cmd := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-p", proto, "--dport", port, "-j", "ACCEPT")
		err := cmd.Run()
		if err == nil { //exists
			cmd = exec.Command("iptables", "-D", "INPUT", "-s", ip, "-p", proto, "--dport", port, "-j", "ACCEPT")
			cmd.Stderr = w
			err = cmd.Run()
			if err != nil {
				w.Write([]byte(strings.Join(cmd.Args, " ")))
				log.Println(err)
			}
		}
	case "list":
		cmd := exec.Command("iptables", "-S", "INPUT")
		cmd.Stderr = w
		cmd.Stdout = w
		err := cmd.Run()
		if err != nil {
			w.Write([]byte(strings.Join(cmd.Args, " ")))
			log.Println(err)
		}
	default:
		http.Error(w, "unavailable action", 500)
		return
	}
	w.Write([]byte("\r\ndone"))
}

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

	host := flag.String("host", "", "ip")
	port := flag.String("port", "", "port")
	cafile := flag.String("ca", "", "CA")
	certfile := flag.String("cert", "", "server cert")
	keyfile := flag.String("key", "", "server key")
	flag.Parse()

	if len(*host) > 0 {
		config.Host = *host
	}
	if len(*port) > 0 {
		config.Port = *port
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

	//default rules
	var rules [][]string
	rules = append(rules, []string{"-I", "INPUT", "-p", "tcp", "--dport", config.Port, "-j", "ACCEPT"})
	rules = append(rules, []string{"-I", "OUTPUT", "-p", "tcp", "--dport", config.Port, "-j", "ACCEPT"})
	//this rule should be the last one
	rules = append(rules, []string{"-A", "INPUT", "-j", "DROP"})

	cmd := exec.Command("iptables", "-P", "INPUT", "DROP")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Println(cmd.Args)
	}
	for _, rule := range rules {
		cmd = exec.Command("iptables", append([]string{"-C"}, rule[1:]...)...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil { //rule doesn't exist
			cmd = exec.Command("iptables", rule...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				log.Println(cmd.Args)
			}
		}
	}

	//CA
	caPEM, err := ioutil.ReadFile(config.Ca)
	if err != nil {
		log.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	//server
	server := &http.Server{
		Addr:    config.Host + ":" + config.Port,
		Handler: &testHandler{},
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	if err := server.ListenAndServeTLS(config.Cert, config.Key); err != nil {
		log.Fatal(err)
	}
}
