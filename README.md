# arduino_exporter

Export readings from Arduino as prometheus metrics.

## Install firmware on Ardunio

```
arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/cu.usbmodem1421301 ./firmware
```
