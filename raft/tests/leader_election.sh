cd ..
sudo docker compose pause server1
sleep 4
sudo docker compose unpause server1
sleep 4
sudo docker compose pause server2
sleep 4
sudo docker compose unpause server2
sleep 4
sudo docker compose pause server3
sleep 4
sudo docker compose unpause server3
cd tests