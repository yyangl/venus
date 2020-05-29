module github.com/yyangl/venus

go 1.14

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

require (
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/grpc v1.29.1
)

replace google.golang.org/grpc v1.29.1 => google.golang.org/grpc v1.26.0
