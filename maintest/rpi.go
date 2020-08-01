package main

import (
	"fmt"
	"log"

	// "github.com/blesswinsamuel/rpi_exporter/d2r2/dht"

	"github.com/blesswinsamuel/rpi_exporter/dht"
	"periph.io/x/periph/host/rpi"
)

func main() {
	// temperature, humidity, retried, err :=
	// 	dht.ReadDHTxxWithRetry(dht.DHT11, rpi.P1_7, false, 10)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	if err := dht.HostInit(); err != nil {
		log.Fatal(err)
	}
	d, err := dht.NewDHT(rpi.P1_7, dht.Celsius, "dht11")
	if err != nil {
		log.Fatal(err)
	}
	humidity, temperature, err := d.ReadRetry(11)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	fmt.Printf("humidity: %v\n", humidity)
	fmt.Printf("temperature: %v\n", temperature)

	// temperature, humidity, retried, err :=
	// 	dht.ReadDHTxxWithRetry(dht.DHT11, rpi.P1_7, false, 10)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // Print temperature and humidity
	// fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
	// 	temperature, humidity, retried)

	// t := time.NewTicker(500 * time.Millisecond)
	// for l := gpio.Low; ; l = !l {
	// 	// Lookup a pin by its location on the board:
	// 	if err := rpi.P1_7.Out(l); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	<-t.C
	// }
}
