./basic.sh

echo "PAUSING ALL ONE BY ONE"
./leader_election.sh
./basic.sh

echo "TURNING ALL OFF"
./turn_off.sh
sleep 15
./basic.sh
sleep 5

echo "PAUSE 1"
./pause.sh
./basic.sh
sleep 5

echo "PAUSE 2"
./pause_another.sh
./basic.sh