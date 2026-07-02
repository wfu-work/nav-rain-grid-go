sudo chmod +x navRainGridApp

sudo cp raind-grid.service /etc/systemd/system/

sudo systemctl daemon-reload

sudo systemctl enable raind.service

sudo systemctl stop raind.service

sudo systemctl start raind.service
