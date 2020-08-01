package rustydht

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
)

// TemperatureUnit is the temperature unit wanted, either Celsius or Fahrenheit
type TemperatureUnit int

const (
	// Celsius temperature unit
	Celsius TemperatureUnit = iota
	// Fahrenheit temperature unit
	Fahrenheit
)

// DHT struct to interface with the sensor.
// Call NewDHT to create a new one.
type DHT struct {
	pin             gpio.PinIO
	temperatureUnit TemperatureUnit
	sensorType      string
	numErrors       int
	lastRead        time.Time
}

// NewDHT to create a new DHT struct.
// sensorType is dht11 for DHT11, anything else for AM2302 / DHT22.
func NewDHT(pin gpio.PinIO, temperatureUnit TemperatureUnit, sensorType string) (*DHT, error) {
	dht := &DHT{temperatureUnit: temperatureUnit, pin: pin}

	// set sensorType
	sensorType = strings.ToLower(sensorType)
	if sensorType == "dht11" {
		dht.sensorType = "dht11"
	}

	// set pin to high so ready for first read
	err := dht.pin.Out(gpio.High)
	if err != nil {
		return nil, fmt.Errorf("pin out high error: %v", err)
	}

	// set lastRead a second before to give the pin a second to warm up
	dht.lastRead = time.Now().Add(-1 * time.Second)

	return dht, nil
}

const maxCycles = 1000
const TIMEOUT = time.Minute

// readBits will get the bits for humidity and temperature
func (dht *DHT) readBits() ([]int, error) {
	// create variables ahead of time before critical timing part
	var err error

	waitLevel := func(wantLevel gpio.Level) time.Duration {
		startTime := time.Now()
		// loopCnt := 0
		for {
			gotLevel := dht.pin.Read()
			if gotLevel == wantLevel {
				break
			}
			// loopCnt++
			// if loopCnt == maxCycles {
			// 	return TIMEOUT
			// }
		}
		return time.Since(startTime)
	}

	// set lastRead so do not read more than once every 2 seconds
	dht.lastRead = time.Now()

	// disable garbage collection during critical timing part
	gcPercent := debug.SetGCPercent(-1)

	defer func() {
		// release the bus
		// set pin to high so ready for next time
		if err := dht.pin.Out(gpio.High); err != nil {
			fmt.Printf("pin out high error: %v", err)
		}
	}()

	// {
	// 	// Go into high impedence state to let pull-up raise data line level and
	// 	// start the reading process.
	// 	err = dht.pin.Out(gpio.High)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("pin out high error: %v", err)
	// 	}
	// 	time.Sleep(time.Millisecond)
	// 	// err = dht.pin.In(gpio.PullUp, gpio.NoEdge)
	// 	// if err != nil {
	// 	// 	return nil, fmt.Errorf("pin out high error: %v", err)
	// 	// }
	// }

	// send start signal
	{
		// First set data line low for a period according to sensor type
		err = dht.pin.Out(gpio.Low)
		if err != nil {
			return nil, fmt.Errorf("pin out low error: %v", err)
		}
		time.Sleep(18 * time.Millisecond) // data sheet says at least 18ms, 20ms just to be safe

		// End the start signal by setting data line high for 40 microseconds.
		err = dht.pin.Out(gpio.High)
		if err != nil {
			return nil, fmt.Errorf("pin out high error: %v", err)
		}
		// Delay a moment to let sensor pull data line low.
		// time.Sleep(40 * time.Microsecond)
	}

	// get data from sensor
	var initCycles []time.Duration = make([]time.Duration, 4)
	var cycles []time.Duration = make([]time.Duration, 80)
	{
		err = dht.pin.In(gpio.PullUp, gpio.NoEdge)
		if err != nil {
			return nil, fmt.Errorf("pin in error: %v", err)
		}

		initCycles[0] = waitLevel(gpio.Low)
		initCycles[1] = waitLevel(gpio.High) // 50us
		initCycles[2] = waitLevel(gpio.Low)  // 50us
		// initCycles[0] = waitLevel(gpio.Low)
		// initCycles[1] = waitLevel(gpio.High)
		for i := 0; i < 80; i += 2 {
			cycles[i] = waitLevel(gpio.High)
			cycles[i+1] = waitLevel(gpio.Low)
			// cycles[i] = waitLevel(gpio.Low)
			// cycles[i+1] = waitLevel(gpio.High)
		}
		initCycles[3] = waitLevel(gpio.High)
	}

	fmt.Println(initCycles)
	fmt.Println(cycles)

	// enable garbage collection, done with critical part
	debug.SetGCPercent(gcPercent)

	var data []int = make([]int, 5)
	for i := 0; i < 40; i++ {
		lowDur := cycles[2*i]
		highDur := cycles[2*i+1]

		if lowDur == TIMEOUT || highDur == TIMEOUT {
			fmt.Println("Timeout", i)
			// 	return nil, errors.New("Timeout")
		}
		if lowDur < 40*time.Microsecond || lowDur > 60*time.Microsecond {
			fmt.Printf("low duration is not around 50us (%s)\n", lowDur)
		}

		data[i/8] <<= 1
		// Now compare the low and high cycle times to see if the bit is a 0 or 1.
		if highDur > lowDur {
			// High cycles are greater than 50us low cycle count, must be a 1.
			data[i/8] |= 1
		}
	}

	return data, nil
}

// bitsToValues will convert the bits into humidity and temperature values
func (dht *DHT) bitsToValues(data []int) (humidity float64, temperature float64, err error) {
	var sumTotal int = 0
	for i, b := range data[0:4] {
		fmt.Printf("%3X\t%3d\t<- %d\n", b, b, i)
		sumTotal += b
	}
	fmt.Printf("%3X\t%3d\t<- Checksum\n", data[4], data[4])
	fmt.Println(data)
	fmt.Printf("%3X\t%3d\t<- Calculated checksum\n", sumTotal, sumTotal)

	humidityInt := data[0]
	temperatureInt := data[2]
	checkSum := data[4]

	// humidity is between 0 % to 100 %
	if humidityInt < 0 || humidityInt > 100 {
		err = fmt.Errorf("bad data - humidity: %v", humidityInt)
		return
	}
	// temperature between 0 C to 50 C
	if temperatureInt < 0 || temperatureInt > 50 {
		err = fmt.Errorf("bad data - temperature: %v", temperatureInt)
		return
	}
	// check checkSum
	if checkSum != sumTotal {
		err = fmt.Errorf("bad data - check sum fail")
		return
	}

	humidity = float64(humidityInt)
	if dht.temperatureUnit == Celsius {
		temperature = float64(temperatureInt)
	} else {
		temperature = float64(temperatureInt)*9.0/5.0 + 32.0
	}
	return
}

// HostInit calls periph.io host.Init(). This needs to be done before DHT can be used.
func HostInit() error {
	state, err := host.Init()
	// Prints the loaded driver.
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}

	// Prints the driver that were skipped as irrelevant on the platform.
	fmt.Printf("Drivers skipped:\n")
	for _, failure := range state.Skipped {
		fmt.Printf("- %s: %s\n", failure.D, failure.Err)
	}

	// Having drivers failing to load may not require process termination. It
	// is possible to continue to run in partial failure mode.
	fmt.Printf("Drivers failed to load:\n")
	for _, failure := range state.Failed {
		fmt.Printf("- %s: %v\n", failure.D, failure.Err)
	}
	return err
}

// Read reads the sensor once, returing humidity and temperature, or an error.
// Note that Read will sleep for at least 2 seconds between last call.
// Each reads error adds a half second to sleep time to max of 30 seconds.
func (dht *DHT) Read() (humidity float64, temperature float64, err error) {
	// // set sleepTime
	// var sleepTime time.Duration
	// if dht.numErrors < 57 {
	// 	sleepTime = (2 * time.Second) + (time.Duration(dht.numErrors) * 500 * time.Millisecond)
	// } else {
	// 	// sleep max of 30 seconds
	// 	sleepTime = 30 * time.Second
	// }
	// sleepTime -= time.Since(dht.lastRead)

	// // sleep between 2 and 30 seconds
	// time.Sleep(sleepTime)

	// read bits from sensor
	var bits []int
	bits, err = dht.readBits()
	if err != nil {
		return
	}

	// covert bits to humidity and temperature
	humidity, temperature, err = dht.bitsToValues(bits)

	return
}

// ReadRetry will call Read until there is no errors or the maxRetries is hit.
// Suggest maxRetries to be set around 11.
func (dht *DHT) ReadRetry(maxRetries int) (humidity float64, temperature float64, err error) {
	for i := 0; i < maxRetries; i++ {
		humidity, temperature, err = dht.Read()
		if err == nil {
			return
		}
		fmt.Println("Warning - error: %v", err)
	}
	return
}

// ReadBackground it meant to be run in the background, run as a Goroutine.
// sleepDuration is how long it will try to sleep between reads.
// If there is ongoing read errors there will be no notice except that the values will not be updated.
// Will continue to read sensor until stop is closed.
// After it has been stopped, the stopped chan will be closed.
// Will panic if humidity, temperature, or stop are nil.
func (dht *DHT) ReadBackground(humidity *float64, temperature *float64, sleepDuration time.Duration, stop chan struct{}, stopped chan struct{}) {
	var humidityTemp float64
	var temperatureTemp float64
	var err error
	var startTime time.Time

Loop:
	for {
		startTime = time.Now()
		humidityTemp, temperatureTemp, err = dht.Read()
		if err == nil {
			// no read error, save result
			*humidity = humidityTemp
			*temperature = temperatureTemp
			// wait for sleepDuration or stop
			select {
			case <-time.After(sleepDuration - time.Since(startTime)):
			case <-stop:
				break Loop
			}
		} else {
			// read error, just check for stop
			select {
			case <-stop:
				break Loop
			default:
			}
		}
	}

	close(stopped)
}
