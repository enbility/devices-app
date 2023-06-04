package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/enbility/devices-app/app"
	"github.com/gorilla/websocket"
)

//go:embed dist
var web embed.FS

const (
	httpdPort int = 7050
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow connection from any host
		return true
	},
}

func serveWs(cem *app.Cem, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	conn := app.NewConnection(cem, ws)
	cem.AddConnection(conn)
}

func usage() {
	fmt.Println("General Usage:")
	fmt.Println("  devicesapp <httpd-port> <eebus-port> <crtfile> <keyfile> <serial>")
	fmt.Println("    <httpd-port> Optional port for the HTTPD server")
	fmt.Println("    <eebus-port> Optional port for the EEBUS service")
	fmt.Println("    <crt-file>   Optional filepath for the cert file")
	fmt.Println("    <key-file>   Option filepath for the key file")
	fmt.Println("    <serial>     Option mDNS serial string")
	fmt.Println()
	fmt.Println("Default values:")
	fmt.Println("  httpd-port:", httpdPort)
	fmt.Println("  eebus-port: 4815")
	fmt.Println("  crt-file:   cert.crt (same folder as executable)")
	fmt.Println("  key-file:   cert.key (same folder as executable)")
	fmt.Println("  serial:     123456789")
	fmt.Println()
	fmt.Println("If no cert-file or key-file parameters are provided and")
	fmt.Println("the files do not exist, they will be created automatically.")
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		usage()
		return
	}

	portHttpd := httpdPort
	if len(os.Args) > 1 {
		if tempPort, err := strconv.Atoi(os.Args[1]); err == nil {
			portHttpd = tempPort
		}
	}
	log.Println("Web Server running at port", portHttpd)

	hems := app.NewHems()
	hems.Run()

	serverRoot, err := fs.Sub(web, "dist")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(serverRoot)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hems, w, r)
	})
	go func() {
		host := fmt.Sprintf("0.0.0.0:%d", portHttpd)
		addr := flag.String("addr", host, "http service address")

		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Clean exit to make sure mdns shutdown is invoked
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	// User exit
}
