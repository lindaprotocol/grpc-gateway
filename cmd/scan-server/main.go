package main

import (
    "context"
    "database/sql"
    "flag"
    "fmt"
    "log"
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/lindaprotocol/grpc-gateway/api"
    "github.com/lindaprotocol/grpc-gateway/api/scan"
    "github.com/lindaprotocol/grpc-gateway/internal/service"
    "github.com/lindaprotocol/grpc-gateway/internal/storage"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/reflection"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var (
    grpcPort          = flag.Int("grpc-port", 50051, "gRPC port")
    httpPort          = flag.Int("http-port", 8080, "HTTP port")
    lindaNodeEndpoint = flag.String("linda-node", "localhost:50051", "Linda node endpoint")
    dbConnection      = flag.String("db", "postgresql://user:pass@localhost/lindascan?sslmode=disable", "Database connection")
)

func main() {
    flag.Parse()

    // Initialize database
    gormDB, err := initDatabase(*dbConnection)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Initialize Linda node client
    nodeConn, err := grpc.Dial(*lindaNodeEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to Linda node: %v", err)
    }
    defer nodeConn.Close()

    walletClient := api.NewWalletClient(nodeConn)
    solidityClient := api.NewWalletSolidityClient(nodeConn)

    // Initialize services
    scanService := service.NewScanService(walletClient, solidityClient, gormDB)
    tagService := service.NewTagService(gormDB)
    statsService := service.NewStatsService(walletClient, solidityClient, gormDB)

    // Start gRPC server
    go startGRPCServer(scanService, tagService, statsService)

    // Start HTTP gateway
    startHTTPGateway()
}

func initDatabase(connStr string) (*gorm.DB, error) {
    // Parse connection string
    // For PostgreSQL
    db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Auto migrate schemas
    db.AutoMigrate(&storage.Tag{}, &storage.Statistic{}, &storage.MarketData{})

    return db, nil
}

func startGRPCServer(scanService *service.ScanService, tagService *service.TagService, statsService *service.StatsService) {
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    s := grpc.NewServer()
    scan.RegisterScanServiceServer(s, scanService)
    reflection.Register(s)

    log.Printf("gRPC server listening on :%d", *grpcPort)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func startHTTPGateway() {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    mux := runtime.NewServeMux(
        runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
            MarshalOptions: protojson.MarshalOptions{
                UseProtoNames:   true,
                EmitUnpopulated: true,
            },
            UnmarshalOptions: protojson.UnmarshalOptions{
                DiscardUnknown: true,
            },
        }),
    )

    opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
    
    // Register scan service
    err := scan.RegisterScanServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", *grpcPort), opts)
    if err != nil {
        log.Fatalf("Failed to register gateway: %v", err)
    }

    // Add CSV support
    mux = addCSVSupport(mux)

    log.Printf("HTTP gateway listening on :%d", *httpPort)
    if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), mux); err != nil {
        log.Fatalf("Failed to serve HTTP: %v", err)
    }
}

func addCSVSupport(mux *runtime.ServeMux) *runtime.ServeMux {
    // Custom handler for CSV format
    return mux
}