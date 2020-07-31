package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/d2r2/go-dht"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"golang.org/x/sync/errgroup"
)

var (
	listen = flag.String("listen",
		"localhost:9153",
		"listen address")
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
	prometheus.MustRegister(leases)
}

type server struct {
	promHandler http.Handler
}

func (s *server) metrics(w http.ResponseWriter, r *http.Request) {
	var eg errgroup.Group

	eg.Go(func() error {
		temperature, humidity, retried, err := dht.ReadDHTxxWithRetry(dht.DHT11, 4, false, 5)
		if err != nil {
			return fmt.Errorf("failed to read temperature (retried %d times): %v", retried, err)
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
	s := &server{
		promHandler: promhttp.Handler(),
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
