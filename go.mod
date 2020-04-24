module github.com/condensat/bank-accounting

go 1.14

require (
	github.com/condensat/bank-core v0.0.2-0.20200422130000-e959bcdf1c1d
	github.com/google/uuid v1.1.2 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/nats-io/nats.go v1.10.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/btcsuite/btcd => github.com/condensat/btcd v0.20.1-beta.0.20200424100000-5dc523e373e2
