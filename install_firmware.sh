#!/bin/bash -e

# wget -P /tmp/firmware https://raw.githubusercontent.com/blesswinsamuel/arduino_exporter/master/firmware/firmware.ino
# arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/ttyACM0 /tmp/firmware

SSH_USER=${SSH_USER:-pi}
SSH_HOST=${SSH_HOST:-192.168.1.5}
SSH_PORT=${SSH_PORT:-2222}

HOST="$SSH_USER@$SSH_HOST"

ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
docker stop arduino-exporter
mkdir -p /tmp/firmware
EOF
scp -P "$SSH_PORT" ./firmware/firmware.ino $HOST:/tmp/firmware/firmware.ino
ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/ttyACM0 /tmp/firmware
docker start arduino-exporter
EOF
