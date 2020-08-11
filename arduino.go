package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarm/serial"

	log "github.com/google/logger"
)

var (
	listen      = flag.String("listen", "localhost:9101", "listen address")
	metricsPath = flag.String("metrics_path", "/metrics", "path under which metrics are served")
	verbose     = flag.Bool("verbose", false, "print info level logs to stdout")
	serialPort  = flag.String("port", "", "serial port")
)

type server struct {
	serial *serial.Port
}

var lastRead = ""

func (s *server) readSerial() {
	tmpStr := ""
	for {
		buf := make([]byte, 128)
		n, err := s.serial.Read(buf)
		if err != nil {
			log.Error(err)
		}
		strs := strings.Split(string(buf[:n]), "\n")
		if len(strs) > 1 {
			tmpStr += strs[0]
			fmt.Println(tmpStr)
			if strings.HasPrefix(tmpStr, "METRICS: ") {
				lastRead = strings.ReplaceAll(tmpStr, "$", "\n")
			}
			tmpStr = ""
		}
		tmpStr += strs[len(strs)-1]
	}
}

func (s *server) handleArduino(w http.ResponseWriter, r *http.Request) {
	for key, values := range r.URL.Query() {
		for _, value := range values {
			_, err := s.serial.Write([]byte(fmt.Sprintf("%s=%s\n", key, value)))
			if err != nil {
				log.Error(err)
			}
		}
	}
	w.Write([]byte(`OK`))
}

func (s *server) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(lastRead))
}

func main() {
	flag.Parse()
	defer log.Init("ArduinoExporter", *verbose, true, ioutil.Discard).Close()

	if *serialPort == "" {
		flag.Usage()
		log.Fatal("port is required")
	}

	s := &server{}
	c := &serial.Config{Name: *serialPort, Baud: 9600}

	var err error
	s.serial, err = serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	go s.readSerial()

	http.HandleFunc(*metricsPath, s.metrics)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Raspberry Pi Exporter</title></head>
			<body>
			<h1>Raspberry Pi Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body></html>`))
	})
	http.HandleFunc("/arduino", s.handleArduino)
	log.Infoln("Listening on", *listen)
	log.Infoln("Serving metrics under", *metricsPath)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
