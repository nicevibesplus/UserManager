# UserManager
Web Interface to manage Users in OpenLDAP

## Installation
```
cp config.conf.sample config.conf
vi config.conf # configure the application to the specific setup
```

## Run as service
### systemd
```
vi userManager.service #Change Installation Path
sudo cp userManager.service /etc/systemd/system/
sudo systemctl enable userManager
sudo systemctl start userManager
```

## license
see `LICENSE`
