#!/bin/bash -e

SSH_USER=${SSH_USER:-pi}
SSH_HOST=${SSH_HOST:-192.168.1.5}
SSH_PORT=${SSH_PORT:-2222}

HOST="$SSH_USER@$SSH_HOST"

# --inplace is needed to preserve docker volumes (for file volume)
rsync -acvz --delete --no-owner --no-group --inplace -e "ssh -p $SSH_PORT" . "$HOST:~/dev/rpi_exporter/"

ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
cd ~/dev/rpi_exporter

# go build .
/usr/local/go/bin/go build -o rpi_exporter ./d2r2/
sudo ./rpi_exporter
EOF
