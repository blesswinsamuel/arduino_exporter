package main

import (
	"time"

	"github.com/prometheus/common/log"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

func light(pin gpio.PinIO, ch <-chan bool) {
	for {
		log.Debugf("Light waiting")
		v := gpio.High
		if <-ch {
			v = gpio.High
		} else {
			v = gpio.Low
		}
		log.Debugf("Setting light to %s", v)
		if err := pin.Out(v); err != nil {
			log.Error(err)
		}
	}
}

func buzzer(pin gpio.PinIO, ch <-chan bool) {
	for {
		v := gpio.High
		log.Debugf("Buzzer waiting")
		if <-ch {
			v = gpio.High
		} else {
			v = gpio.Low
		}
		log.Debugf("Setting buzzer to %s", v)
		if err := pin.Out(v); err != nil {
			log.Error(err)
		}
	}
}

// func ldr(ch chan<- bool) {
// 	for l := gpio.Low; ; l = !l {
// 		if err := rpi.P1_13.ADC(); err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// }

func main() {
	log.Base().SetLevel("debug")
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// analog.PinADC{

	// }
	// d, err := dht.NewDHT(rpi.P1_7, dht.Celsius, "dht11")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// humidity, temperature, retries, err := d.ReadRetry(11)
	// if err != nil {
	// 	fmt.Println("Read error:", err)
	// 	return
	// }

	// fmt.Printf("humidity: %v\n", humidity)
	// fmt.Printf("temperature: %v\n", temperature)
	// fmt.Printf("retries: %v\n", retries)

	// t := time.NewTicker(500 * time.Millisecond)
	ledCh := make(chan bool)
	buzzerCh := make(chan bool)

	go light(rpi.P1_11, ledCh)
	go buzzer(rpi.P1_13, buzzerCh)

	ledCh <- true
	buzzerCh <- true
	time.Sleep(time.Second)
	ledCh <- false
	buzzerCh <- false
	time.Sleep(time.Millisecond)
}
