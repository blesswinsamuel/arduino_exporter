#!/bin/bash

wget -P /tmp/firmware https://raw.githubusercontent.com/blesswinsamuel/arduino_exporter/master/firmware/firmware.ino
arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/ttyACM0 /tmp/firmware
