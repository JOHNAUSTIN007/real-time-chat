<!-- To install required packages -->
go mod tidy

<!-- Prerequisites -->
Need redis in the server or system where this is deployed or running

<!-- Command to activate redis server -->
redis-server


<!-- To run the application -->
go run main.go

<!-- Packages used -->
used gorilla for websocket connections
used redis to store data
