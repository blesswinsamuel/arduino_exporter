package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/MichaelS11/go-dht"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"golang.org/x/sync/errgroup"
)

var (
	listen = flag.String("listen",
		"localhost:9153",
		"listen address")
	metricsPath = flag.String("metrics_path",
		"/metrics",
		"path under which metrics are served")
)

var (
	temperatureMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "temperature",
		Help: "Temperature from DHT sensor",
	})
	humidityMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "humidity",
		Help: "Humidity from DHT sensor",
	})
)

func init() {
	prometheus.MustRegister(temperatureMetric)
	prometheus.MustRegister(humidityMetric)
}

type server struct {
	promHandler http.Handler
	dht         *dht.DHT
}

func (s *server) metrics(w http.ResponseWriter, r *http.Request) {
	var eg errgroup.Group

	eg.Go(func() error {
		temperature, humidity, err := s.dht.ReadRetry(11)
		if err != nil {
			return fmt.Errorf("failed to read temperature: %v", err)
		}
		temperatureMetric.Set(temperature)
		humidityMetric.Set(humidity)
		return nil
	})

	if err := eg.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.promHandler.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	err := dht.HostInit()
	if err != nil {
		log.Fatal("HostInit error:", err)
	}
	dht, err := dht.NewDHT("GPIO4", dht.Celsius, "dht11")
	if err != nil {
		log.Fatal("NewDHT error:", err)
	}
	temperature, humidity, err := dht.ReadRetry(6)
	if err != nil {
		log.Fatalf("failed to read temperature: %v", err)
	}
	log.Infof("Temperature = %v*C, Humidity = %v%%\n", temperature, humidity)
	s := &server{
		promHandler: promhttp.Handler(),
		dht:         dht,
	}
	http.HandleFunc(*metricsPath, s.metrics)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Dnsmasq Exporter</title></head>
			<body>
			<h1>Dnsmasq Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body></html>`))
	})
	log.Infoln("Listening on", *listen)
	log.Infoln("Serving metrics under", *metricsPath)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
