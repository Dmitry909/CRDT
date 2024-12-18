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

##### different keys

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'
```

##### equal keys

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'
```
