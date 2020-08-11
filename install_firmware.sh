#!/bin/bash

https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_ARMv7.tar.gz

arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/cu.usbmodem1421301 ./firmware
