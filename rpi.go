package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/blesswinsamuel/rpi_exporter/dht"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

var (
	listen = flag.String("listen",
		"localhost:9101",
		"listen address")
	metricsPath = flag.String("metrics_path",
		"/metrics",
		"path under which metrics are served")
)

var (
	dhtTemperatureMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pi_dht_temperature",
		Help: "Temperature from DHT sensor",
	})
	dhtHumidityMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pi_dht_humidity",
		Help: "Humidity from DHT sensor",
	})
	dhtRetriesMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pi_dht_retries",
		Help: "Retries from DHT sensor",
	})
	dhtFailuresMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pi_dht_failed",
		Help: "Failures from DHT sensor",
	})
	dhtReadingsCountMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pi_dht_readings",
		Help: "Successful readings from DHT sensor",
	})
)

func init() {
	prometheus.MustRegister(dhtTemperatureMetric)
	prometheus.MustRegister(dhtHumidityMetric)
	prometheus.MustRegister(dhtRetriesMetric)
	prometheus.MustRegister(dhtFailuresMetric)
	prometheus.MustRegister(dhtReadingsCountMetric)
}

type server struct {
	promHandler http.Handler
	dht         *dht.DHT
}

func (s *server) calcTemperatureInBackground() {
	for {
		humidity, temperature, retries, err := s.dht.ReadRetry(11)
		if err != nil {
			dhtFailuresMetric.Add(1)
			log.Errorf("failed to read temperature: %v", err)
			continue
		}
		dhtReadingsCountMetric.Inc()
		dhtTemperatureMetric.Set(temperature)
		dhtHumidityMetric.Set(humidity)
		dhtRetriesMetric.Add(float64(retries))
		time.Sleep(5 * time.Second)
	}
}

func (s *server) metrics(w http.ResponseWriter, r *http.Request) {
	s.promHandler.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	if _, err := host.Init(); err != nil {
		log.Fatal("HostInit error:", err)
	}
	dht, err := dht.NewDHT(rpi.P1_7, dht.Celsius, "dht11")
	if err != nil {
		log.Fatal("NewDHT error:", err)
	}
	s := &server{
		promHandler: promhttp.Handler(),
		dht:         dht,
	}
	go s.calcTemperatureInBackground()

	http.HandleFunc(*metricsPath, s.metrics)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Raspberry Pi Exporter</title></head>
			<body>
			<h1>Raspberry Pi Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body></html>`))
	})
	log.Infoln("Listening on", *listen)
	log.Infoln("Serving metrics under", *metricsPath)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
