package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tarm/serial"

	log "github.com/google/logger"
)

var (
	listen      = flag.String("listen", "localhost:9101", "listen address")
	metricsPath = flag.String("metrics_path", "/metrics", "path under which metrics are served")
	// verbose     = flag.Bool("verbose", false, "print info level logs to stdout")
	serialPort = flag.String("port", "", "serial port")
)

type server struct {
	serial     *serial.Port
	serialRead chan string
}

func (s *server) serialReadLoop() {
	line := ""
	for {
		buf := make([]byte, 128)
		n, err := s.serial.Read(buf)
		if err != nil {
			log.Error(err)
		}
		strs := strings.Split(string(buf[:n]), "\n")
		if len(strs) > 1 {
			line += strs[0]
			fmt.Println(line)
			s.serialRead <- strings.ReplaceAll(line, "$", "\n")
			line = ""
		}
		line += strs[len(strs)-1]
	}
}

func (s *server) sendSerialCommand(command string) error {
	_, err := s.serial.Write([]byte(command + "\n"))
	return err
}

func (s *server) serialReadLine(ctx context.Context, prefix string) string {
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	prefix = prefix + ": "
	for {
		select {
		case line := <-s.serialRead:
			if strings.HasPrefix(line, prefix) {
				return strings.ReplaceAll(strings.TrimPrefix(line, prefix), "$", "\n")
			}
		case <-timer.C:
			return ""
		case <-ctx.Done():
			return ""
		}
	}
}

func (s *server) handleArduino(w http.ResponseWriter, r *http.Request) {
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if err := s.sendSerialCommand(fmt.Sprintf("%s=%s", key, value)); err != nil {
				log.Error(err)
				w.WriteHeader(500)
				w.Write([]byte(`KO`))
				return
			}
		}
	}
	w.Write([]byte(`OK`))
}

func (s *server) ledBlink(duration int) {
	if err := s.sendSerialCommand(fmt.Sprintf("led=%d", duration)); err != nil {
		log.Error(err)
	}
}

func (s *server) metrics(w http.ResponseWriter, r *http.Request) {
	s.ledBlink(-1)
	defer s.ledBlink(0)
	if err := s.sendSerialCommand("metrics"); err != nil {
		log.Error(err)
		w.Write([]byte(""))
		return
	}
	metrics := s.serialReadLine(r.Context(), "METRICS")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metrics))
}

func main() {
	flag.Parse()
	defer log.Init("ArduinoExporter", true, true, ioutil.Discard).Close()

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

	s.serialRead = make(chan string)
	go s.serialReadLoop()

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
