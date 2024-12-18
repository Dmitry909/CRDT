### One node

```
go run app/main.go 0 8000 '8000'

curl -i -X GET http://localhost:8000/read?key=key1
# 404, key not found

curl -i -X PATCH http://localhost:8000/set -H "Content-Type: application/json" -d '{"values": {"key1": "value1"}}'
# 200

curl -i -X GET http://localhost:8000/read?key=key1
# value1
```

### Two nodes

##### simple

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'

curl -i -X PATCH http://localhost:8001/set -H "Content-Type: application/json" -d '{"values": {"key1": "value1"}}'

curl -i -X GET http://localhost:8000/read?key=key1
# value1

curl -i -X GET http://localhost:8001/read?key=key1
# value1
```

##### network partition, different keys

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'

sudo iptables -A INPUT -p tcp --dport 8000 -j DROP; sudo iptables -A INPUT -p udp --dport 8000 -j DROP; sudo iptables -A OUTPUT -p tcp --sport 8000 -j DROP; sudo iptables -A OUTPUT -p udp --sport 8000 -j DROP

curl -i -X PATCH http://localhost:8001/set -H "Content-Type: application/json" -d '{"values": {"key0": "value0"}}'

sudo iptables -F

sudo iptables -A INPUT -p tcp --dport 8001 -j DROP; sudo iptables -A INPUT -p udp --dport 8001 -j DROP; sudo iptables -A OUTPUT -p tcp --sport 8001 -j DROP; sudo iptables -A OUTPUT -p udp --sport 8001 -j DROP

curl -i -X PATCH http://localhost:8000/set -H "Content-Type: application/json" -d '{"values": {"key1": "value1"}}'

sudo iptables -F

curl -i -X GET http://localhost:8000/read?key=key0
# value0
curl -i -X GET http://localhost:8000/read?key=key1
# value1

curl -i -X GET http://localhost:8001/read?key=key0
# value0
curl -i -X GET http://localhost:8001/read?key=key1
# value1
```

##### network partition, equal keys

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'

sudo iptables -A INPUT -p tcp --dport 8000 -j DROP; sudo iptables -A INPUT -p udp --dport 8000 -j DROP; sudo iptables -A OUTPUT -p tcp --sport 8000 -j DROP; sudo iptables -A OUTPUT -p udp --sport 8000 -j DROP

curl -i -X PATCH http://localhost:8001/set -H "Content-Type: application/json" -d '{"values": {"key0": "value0"}}'

sudo iptables -F

sudo iptables -A INPUT -p tcp --dport 8001 -j DROP; sudo iptables -A INPUT -p udp --dport 8001 -j DROP; sudo iptables -A OUTPUT -p tcp --sport 8001 -j DROP; sudo iptables -A OUTPUT -p udp --sport 8001 -j DROP

curl -i -X PATCH http://localhost:8000/set -H "Content-Type: application/json" -d '{"values": {"key0": "value1"}}'

sudo iptables -F

curl -i -X GET http://localhost:8000/read?key=key0
# value0

curl -i -X GET http://localhost:8001/read?key=key0
# value1
```
