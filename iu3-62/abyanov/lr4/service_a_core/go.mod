module warehouse_system

go 1.25.0

require (
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	warehouse_microservices v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
)

replace warehouse_microservices => ../
