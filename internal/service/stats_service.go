package service

import (
    "context"
    "time"

    "github.com/lindaprotocol/grpc-gateway/api"
    "gorm.io/gorm"
)

type StatsService struct {
    scan.UnimplementedScanServiceServer
    walletClient   api.WalletClient
    solidityClient api.WalletSolidityClient
    db             *gorm.DB
}

func NewStatsService(walletClient api.WalletClient, solidityClient api.WalletSolidityClient, db *gorm.DB) *StatsService {
    return &StatsService{
        walletClient:   walletClient,
        solidityClient: solidityClient,
        db:             db,
    }
}

// Statistics endpoints implementation