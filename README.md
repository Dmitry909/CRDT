# CRDT

Examples:

```
go run app/main.go 0 8000 '8000'
```

```
go run app/main.go 0 8000 '8000,8001'
go run app/main.go 1 8001 '8000,8001'
```

tc:

```
sudo tc qdisc add dev wlp1s0 root handle 1: prio

sudo tc filter add dev wlp1s0 protocol ip parent 1: prio 1 u32 match ip sport 8000 match ip dport 8001 police rate 0kb burst 10kb drop flowid :1
```
