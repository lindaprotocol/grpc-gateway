module github.com/lindaprotocol/grpc-gateway

go 1.21

require (
    github.com/go-redis/redis/v8 v8.11.5
    github.com/golang/glog v1.2.4
    github.com/golang/protobuf v1.5.4
    github.com/gorilla/websocket v1.5.3
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
    github.com/lib/pq v1.10.9
    github.com/patrickmn/go-cache v2.1.0+incompatible
    github.com/rogpeppe/fastuuid v1.2.0
    golang.org/x/net v0.37.0
    google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb
    google.golang.org/grpc v1.71.0
    google.golang.org/protobuf v1.36.5
    gopkg.in/resty.v1 v1.12.0
    gorm.io/driver/postgres v1.5.11
    gorm.io/gorm v1.25.12
)

require (
    github.com/cespare/xxhash/v2 v2.3.0 // indirect
    github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
    github.com/jackc/pgpassfile v1.0.0 // indirect
    github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
    github.com/jackc/pgx/v5 v5.7.2 // indirect
    github.com/jackc/puddle/v2 v2.2.2 // indirect
    github.com/jinzhu/inflection v1.0.0 // indirect
    github.com/jinzhu/now v1.1.5 // indirect
    github.com/kr/text v0.2.0 // indirect
    github.com/rogpeppe/fastuuid v1.2.0 // indirect
    golang.org/x/crypto v0.36.0 // indirect
    golang.org/x/sync v0.12.0 // indirect
    golang.org/x/sys v0.31.0 // indirect
    golang.org/x/text v0.23.0 // indirect
    google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
)

replace github.com/go-resty/resty v1.12.0 => gopkg.in/resty.v1 v1.12.0