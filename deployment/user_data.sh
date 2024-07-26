#!/bin/bash
cd /
sudo apt update
sudo apt install unzip
curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip"
sudo unzip awscliv2.zip
sudo ./aws/install
exec > >(tee /var/log/user-data.log | logger -t user-data -s 2>/dev/console) 2>&1
echo "------------ Starting user_data Script (aws cli installed successfully) --------------"
echo "------------ Installing nginx and musl ---------------"
sudo apt install -y nginx musl
echo "----------- Installing certbot with snap -------------"
sudo snap install core
sudo snap refresh core
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
echo "------------- Creating ask_away directory --------------"
sudo mkdir -p /var/www/ask_away
sudo mkdir -p /var/www/ask_away/app
echo "----------- Syncing /app/ from s3 ----------------"
sudo aws s3 sync s3://ask-away-s3-bucket/app /var/www/ask_away/app
echo "----------- Move letsencrypt/ to etc/letsencrypt --------------"
sudo mv /var/www/ask_away/app/letsencrypt /etc/letsencrypt/
echo "----------- Move nginx.conf to sites-available ------------"
sudo mv /var/www/ask_away/app/nginx.conf /etc/nginx/sites-available/ask-away.mechanicalturk.one
echo "----------- Create Symbolic link to sites-enabled/ask-away.mechanicalturk.one --------------"
sudo ln -s /etc/nginx/sites-available/ask-away.mechanicalturk.one /etc/nginx/sites-enabled/ask-away.mechanicalturk.one
echo "----------- Delete default configuration in sites-enabled & sites-available -----------"
cd /etc/nginx/sites-enabled
sudo rm default
cd ../sites-available
sudo rm default
echo "------------- Reloading Nginx -----------------"
sudo systemctl reload nginx
cd /var/www/ask_away/app
echo "------------ Make build executable ---------------"
sudo chmod +x /var/www/ask_away/app/build
echo "------------ Create app.log file -------------"
touch app.log
echo "------------ Create Chron Job to Update Askers Every Night at Midnight ------------"
0 0 * * * curl -X POST http://localhost:8082/update-askers
echo "------------ Running Executable with env vars and pipe output to app.log ------------"
sudo USE_DEV_SQLITE=true ./build >app.log 2>&1 &
