# UserManager 5000 [![Go Report Card](https://goreportcard.com/badge/github.com/fs-geofs/UserManager)](https://goreportcard.com/report/github.com/fs-geofs/UserManager)
Web Interface to manage Users in OpenLDAP.

## Installation
you need go 1.16 or higher!
```sh
# build the server
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .

# configure the application to the specific setup
cp config.conf.sample config.conf
vi config.conf

# generate key pairs, if necessary
# RSA key for JWT auth
openssl genrsa -out keys/jwt.key 4096
openssl rsa -in keys/jwt.key -pubout -out keys/jwt.pub
# self signed TLS cert
openssl req -x509 -sha256 -nodes -newkey rsa:2048 -days 365 -keyout keys/tls.key -out keys/tls.crt
```

For updates of the binary or frontend, there is a deployment script at `build_deploy.sh`.

## Run as service
### systemd
```sh
editor userManager.service # Change Installation Path
sudo cp userManager.service /etc/systemd/system/
sudo systemctl enable userManager
sudo systemctl start userManager
```

## development
```sh
git clone git@github.com:fs-geofs/UserManager.git
cd UserManager
cp config.conf.sample config.conf
go get .
go run *.go
go fmt
```

To make changes to the frontend without rebuilding the backend, browse index.html manually and
change `API_BASE` to something like `https://localhost:8443`.

To run queries against the local API with a self signed TLS cert:
```
curl --cacert keys/tls.crt https://localhost:8443/api/login
```

## license
see `LICENSE`
