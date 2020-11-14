#!/bin/bash -e

SSH_USER=${SSH_USER:-pi}
SSH_HOST=${SSH_HOST:-192.168.1.5}
SSH_PORT=${SSH_PORT:-2222}

HOST="$SSH_USER@$SSH_HOST"

export GOOS=linux
export GOARCH=arm
# go build -o /tmp/home-controller ./maintest/
go build -o /tmp/home-controller .

ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
sudo killall -9 home-controller || true
EOF
scp -P "$SSH_PORT" /tmp/home-controller $HOST:/tmp/home-controller
ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
sudo /tmp/home-controller
EOF

# # --inplace is needed to preserve docker volumes (for file volume)
# rsync -acvz --delete --no-owner --no-group --inplace -e "ssh -p $SSH_PORT" . "$HOST:~/dev/home-controller/"

# ssh -p "$SSH_PORT" "$HOST" 'bash -ex' <<EOF
# cd ~/dev/home-controller

# # /usr/local/go/bin/go build -o home-controller .
# /usr/local/go/bin/go build -o home-controller ./maintest/
# sudo ./home-controller
# EOF
