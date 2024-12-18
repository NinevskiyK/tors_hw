echo "CREATING: id 1, value 1"
curl -L 127.0.0.1:8081/1 -d 1
sleep 1
echo "\n"

echo "READING 1"
curl -L 127.0.0.1:8081/1 --max-time 1
curl -L 127.0.0.1:8082/1 --max-time 1
curl -L 127.0.0.1:8083/1 --max-time 1
echo "\n"

echo "CREATING id 2, value 2"
curl -L 127.0.0.1:8081/2 -d 2
sleep 1
echo "\n"

echo "CREATING id 2, value 2"
curl -L 127.0.0.1:8081/2 -d 2
sleep 1
echo "\n"

echo "UPDATING id 2 value 22"
curl -L -X PUT 127.0.0.1:8081/2 -d 22
sleep 1
echo "\n"

echo "CAS id 2 old value 22 value 222"
curl -L -X PATCH 127.0.0.1:8081/2 -d "22 222"
sleep 1
echo "\n"

echo "CAS id 2 old value 22 value 222"
curl -L -X PATCH 127.0.0.1:8081/2 -d "22 222"
sleep 1
echo "\n"

echo "READING 2"
curl -L 127.0.0.1:8081/2 --max-time 1
echo ""
curl -L 127.0.0.1:8082/2 --max-time 1
echo ""
curl -L 127.0.0.1:8083/2 --max-time 1
echo "\n"

echo "READING 3"
curl -L 127.0.0.1:8081/3 --max-time 1
echo ""
curl -L 127.0.0.1:8082/3 --max-time 1
echo ""
curl -L 127.0.0.1:8083/3 --max-time 1
echo "\n"