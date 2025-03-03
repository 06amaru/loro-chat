# LORO CHAT

1. Create database 
```
docker run --name some-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
```
2. Run server (follow instructions from loro-server README)
3. Then go to loro-tui folder and run: 
```
go run main.go -server http://localhost:8081
``` 