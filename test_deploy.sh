#!/bin/bash -e

SSH_USER=${SSH_USER:-pi}
SSH_HOST=${SSH_HOST:-192.168.1.5}
SSH_PORT=${SSH_PORT:-2222}

HOST="$SSH_USER@$SSH_HOST"

export GOOS=linux
export GOARCH=arm
go build -o /tmp/rpi_exporter ./maintest/

ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
sudo killall -9 rpi_exporter || true
EOF
scp -P "$SSH_PORT" /tmp/rpi_exporter $HOST:/tmp/rpi_exporter
ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
sudo /tmp/rpi_exporter
EOF

# # --inplace is needed to preserve docker volumes (for file volume)
# rsync -acvz --delete --no-owner --no-group --inplace -e "ssh -p $SSH_PORT" . "$HOST:~/dev/rpi_exporter/"

# ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
# cd ~/dev/rpi_exporter

# # /usr/local/go/bin/go build -o rpi_exporter .
# /usr/local/go/bin/go build -o rpi_exporter ./maintest/
# sudo ./rpi_exporter
# EOF
