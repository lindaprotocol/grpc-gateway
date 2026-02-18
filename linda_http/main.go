package main

import (
    "flag"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/golang/glog"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime" 
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    gw "github.com/lindaprotocol/grpc-gateway/api"
    "github.com/lindaprotocol/grpc-gateway/api/scan"
)

var (
	port = flag.Int("port", 50051, "port of your linda grpc service")
	host = flag.String("host", "localhost", "host of your linda grpc service")
	listen = flag.Int("listen", 18890, "the port that http server listen")
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
	httpPort           = flag.String("http-port", ":18889", "HTTP port")
)

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "X-Requested-With", "token"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "OPTIONS", "PATCH"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	glog.Infof("preflight request for %s", r.URL.Path)
	return
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a single mux instance
	mux := runtime.NewServeMux()
	
	// Use the same endpoint for all services
	grpcEndpoint := *host + ":" + strconv.Itoa(*port)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	fmt.Printf("grpc server:  %s\n", grpcEndpoint)
	fmt.Printf("http port  :  %d\n", *listen)

	// Register Wallet service
	err := gw.RegisterWalletHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	// Register WalletSolidity service
	err = gw.RegisterWalletSolidityHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	// Register WalletExtension service
	err = gw.RegisterWalletExtensionHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	// Register Scan service
	err = scan.RegisterScanServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server with CORS support
	glog.Info("Starting HTTP server on :", *listen)
	return http.ListenAndServe(":"+strconv.Itoa(*listen), allowCORS(mux))
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}