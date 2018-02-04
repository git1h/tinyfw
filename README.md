# tinyfw

Tinyfw is a ~tiny firewall~ tiny tool for simply using to manage allowed ports on linux servers, on which you've performed:

> iptables -A INPUT -j DROP && iptables -P INPUT DROP

* Build tinyfw

> go get -u github.com/git1h/tinyfw/...

> go build github.com/git1h/tinyfw/server

> go build github.com/git1h/tinyfw/client

* Create a CA key and the certificate

> openssl req -new -x509 -days 365 -nodes -newkey rsa:2048 -keyout ca.key -subj /CN=tinyca -out ca.crt

* Create a server key and the certificate

> SERVERIP=192.168.1.1

> openssl req -newkey rsa:2048 -nodes -keyout s.key -subj /CN=$SERVERIP -out s.csr

> openssl x509 -req -in s.csr -CA ca.crt -CAkey ca.key -CAcreateserial -days 365 -out s.crt -extfile <(printf "subjectAltName=IP:$SERVERIP\nkeyUsage=digitalSignature, keyEncipherment\nextendedKeyUsage=serverAuth")

* Create a client key and the certificate

> openssl req -newkey rsa:2048 -nodes -keyout c.key -subj /CN=tinyfwclient -out c.csr

> openssl x509 -req -in c.csr -CA ca.crt -CAkey ca.key -CAcreateserial -days 365 -out c.crt -extfile <(printf "subjectAltName=IP:$SERVERIP\nkeyUsage=digitalSignature, keyEncipherment\nextendedKeyUsage=clientAuth")

* Server

> ./server -host 192.168.1.1 -port 1122 -ca ca.crt -cert s.crt -key s.key

Configuration file is supported, create a file named tinyfw.json, of wich content is as follow:

> {"host":"192.168.1.1","port":"1122","ca":"ca.crt","cert":"s.crt","key":"s.key"}

then just execute:

> ./server

* Client

list rules

> ./client -server https://192.168.1.1:1122 -ca ca.crt -cert c.crt -key c.key -list

add a rule, if -ip option omited, you own ip will be added

> ./client -server https://192.168.1.1:1122 -ca ca.crt -cert c.crt -key c.key -add [-ip your-client-ip] -port 22

delete a rule

> ./client -server https://192.168.1.1:1122 -ca ca.crt -cert c.crt -key c.key -del [-ip your-client-ip] -port 22

Client configuration file is also supported,create a file named tinyfw.json, of wich content is as follow:

> {"server":"https://192.168.1.1:1122","ca":"ca.crt","cert":"s.crt","key":"s.key"}

then just execute:

> ./server -list

> ./server -add -port 22

> ./server -del -port 22
