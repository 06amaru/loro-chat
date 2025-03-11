## LORO Chat

Minimal chat terminal on real time. The project is a template to share simple message using websocket through the terminal so this can be a bootstrap to your personal terminal chat. Basic features are create users, create chat and send messages. 

### Instructions

1. Run Postgres. If you want do it with docker here is the following command: ```docker run --name some-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres```
2. Go to loro-server folder and follow the instructions from README. By default the server port is 8081.
3. Go to loro-tui folder and then use loro terminal: ```go run main.go -server http://localhost:8081``` 
