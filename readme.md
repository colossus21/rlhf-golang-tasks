### Instructions

1. Running samples:
```
cd task_id
go run -tags v1 v1.go
```

2. Running tests:
```
go run -tags v1 v1.go
go run -tags v2 v2.go
go test -tags=v1
go test -tags=v2
go test -tags=v1 -bench=.
go test -tags=v1 -bench=.
```