package service

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/lindaprotocol/grpc-gateway/api"
    "github.com/lindaprotocol/grpc-gateway/api/scan"
    "github.com/lindaprotocol/grpc-gateway/internal/storage"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/emptypb"
    "gorm.io/gorm"
)

type ScanService struct {
    scan.UnimplementedScanServiceServer
    walletClient    api.WalletClient
    solidityClient  api.WalletSolidityClient
    db              *gorm.DB
    httpClient      *http.Client
    cmcAPIKey       string
}

func NewScanService(walletClient api.WalletClient, solidityClient api.WalletSolidityClient, db *gorm.DB) *ScanService {
    return &ScanService{
        walletClient:   walletClient,
        solidityClient: solidityClient,
        db:             db,
        httpClient:     &http.Client{Timeout: 10 * time.Second},
        cmcAPIKey:      "your-coinmarketcap-api-key", // Configure this
    }
}

func (s *ScanService) GetHomepageBundle(ctx context.Context, req *emptypb.Empty) (*scan.HomepageBundle, error) {
    // Get blockchain stats
    nowBlock, err := s.walletClient.GetNowBlock(ctx, &emptypb.Empty{})
    if err != nil {
        return nil, err
    }

    // Get price from CoinMarketCap via proxy
    price, marketCap, volume := s.getMarketData()

    // Get recent blocks (last 10)
    var recentBlocks []*api.Block
    for i := int64(0); i < 10; i++ {
        blockNum := nowBlock.BlockHeader.RawData.Number - i
        if blockNum > 0 {
            block, err := s.walletClient.GetBlockByNum(ctx, &api.NumberMessage{Num: blockNum})
            if err == nil {
                recentBlocks = append(recentBlocks, block)
            }
        }
    }

    // Get recent transactions
    var recentTransactions []*api.Transaction
    // Implementation to fetch recent transactions

    return &scan.HomepageBundle{
        TotalBlocks:        nowBlock.BlockHeader.RawData.Number,
        TotalTransactions:  s.getTotalTransactions(ctx),
        TotalAccounts:      s.getTotalAccounts(ctx),
        TotalContracts:     s.getTotalContracts(ctx),
        TotalTokens:        s.getTotalTokens(ctx),
        PriceUSD:           price,
        MarketCap:          marketCap,
        Volume24h:          volume,
        RecentBlocks:       recentBlocks,
        RecentTransactions: recentTransactions,
    }, nil
}

func (s *ScanService) GetNodeMap(ctx context.Context, req *emptypb.Empty) (*scan.NodeMapResponse, error) {
    // Get nodes from the network
    nodes, err := s.walletClient.ListNodes(ctx, &emptypb.Empty{})
    if err != nil {
        return nil, err
    }

    var nodeInfos []*scan.NodeInfo
    for _, node := range nodes.Nodes {
        // Enhance with geolocation data
        host := string(node.Address.Host)
        country, city, lat, lng := s.geolocateIP(host)

        nodeInfos = append(nodeInfos, &scan.NodeInfo{
            Ip:        host,
            Host:      host,
            Port:      node.Address.Port,
            Country:   country,
            City:      city,
            Latitude:  lat,
            Longitude: lng,
            NodeType:  "FullNode",
        })
    }

    return &scan.NodeMapResponse{Nodes: nodeInfos}, nil
}

func (s *ScanService) GetTop10(ctx context.Context, req *scan.Top10Request) (*scan.Top10Response, error) {
    // Implementation for top accounts/witnesses
    // This would typically come from a database or cache
    return &scan.Top10Response{}, nil
}

func (s *ScanService) ProxyRequest(ctx context.Context, req *scan.ProxyRequestMessage) (*scan.ProxyResponse, error) {
    // Create HTTP request
    httpReq, err := http.NewRequest(req.Method, req.Url, strings.NewReader(string(req.Body)))
    if err != nil {
        return nil, err
    }

    // Add headers
    for k, v := range req.Headers {
        httpReq.Header.Add(k, v)
    }

    // Execute request
    resp, err := s.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Build response headers
    headers := make(map[string]string)
    for k, v := range resp.Header {
        if len(v) > 0 {
            headers[k] = v[0]
        }
    }

    return &scan.ProxyResponse{
        StatusCode: int32(resp.StatusCode),
        Headers:    headers,
        Body:       body,
    }, nil
}

func (s *ScanService) GetLRC20Tokens(ctx context.Context, req *scan.LRC20TokenRequest) (*scan.LRC20TokenListResponse, error) {
    // This would typically query a database that indexes LRC20 tokens
    var tokens []storage.LRC20Token
    query := s.db.Offset(int(req.Start)).Limit(int(req.Limit))

    if req.Contract != "" {
        query = query.Where("contract = ?", req.Contract)
    }

    if req.Sort != "" {
        query = query.Order(req.Sort)
    }

    result := query.Find(&tokens)

    var tokenInfos []*scan.LRC20TokenInfo
    for _, token := range tokens {
        tokenInfos = append(tokenInfos, &scan.LRC20TokenInfo{
            Contract:    token.Contract,
            Name:        token.Name,
            Symbol:      token.Symbol,
            Decimals:    int32(token.Decimals),
            TotalSupply: token.TotalSupply,
            Owner:       token.Owner,
            IssueTime:   token.IssueTime,
            Holders:     token.Holders,
            Transfers:   token.Transfers,
        })
    }

    return &scan.LRC20TokenListResponse{
        Tokens: tokenInfos,
        Total:  result.RowsAffected,
    }, nil
}

func (s *ScanService) GetTokenHolders(ctx context.Context, req *scan.TokenHoldersRequest) (*scan.TokenHoldersResponse, error) {
    // Query token holders from database
    var holders []storage.TokenHolder
    query := s.db.Where("contract = ?", req.Contract).Offset(int(req.Start)).Limit(int(req.Limit))

    if req.Format == "csv" {
        // Handle CSV export
    }

    query.Order("balance desc").Find(&holders)

    var holderInfos []*scan.TokenHolder
    var total int64

    for i, holder := range holders {
        holderInfos = append(holderInfos, &scan.TokenHolder{
            Address:    holder.Address,
            Balance:    holder.Balance,
            Percentage: holder.Percentage,
            Rank:       int64(i + 1 + int(req.Start)),
        })
    }

    s.db.Model(&storage.TokenHolder{}).Where("contract = ?", req.Contract).Count(&total)

    return &scan.TokenHoldersResponse{
        Holders: holderInfos,
        Total:   total,
    }, nil
}

func (s *ScanService) GetAccountList(ctx context.Context, req *scan.AccountListRequest) (*scan.AccountListResponse, error) {
    var accounts []storage.Account
    query := s.db.Offset(int(req.Start)).Limit(int(req.Limit))

    if req.Address != "" {
        query = query.Where("address = ?", req.Address)
    }

    if req.Sort != "" {
        query = query.Order(req.Sort)
    }

    query.Find(&accounts)

    var accountInfos []*scan.AccountInfo
    for _, acc := range accounts {
        accountInfos = append(accountInfos, &scan.AccountInfo{
            Address:      acc.Address,
            Balance:      acc.Balance,
            Transactions: acc.Transactions,
            AccountType:  acc.AccountType,
            Bandwidth:    acc.Bandwidth,
            Energy:       acc.Energy,
        })
    }

    var total int64
    s.db.Model(&storage.Account{}).Count(&total)

    return &scan.AccountListResponse{
        Accounts: accountInfos,
        Total:    total,
    }, nil
}

func (s *ScanService) GetTags(ctx context.Context, req *scan.TagRequest) (*scan.TagListResponse, error) {
    var tags []storage.Tag
    query := s.db.Offset(int(req.Start)).Limit(int(req.Limit))

    if req.Address != "" {
        query = query.Where("address = ?", req.Address)
    }

    query.Order("votes desc").Find(&tags)

    var tagInfos []*scan.Tag
    for _, tag := range tags {
        tagInfos = append(tagInfos, &scan.Tag{
            Id:          int32(tag.ID),
            Address:     tag.Address,
            Tag:         tag.Tag,
            Description: tag.Description,
            Owner:       tag.Owner,
            CreatedAt:   tag.CreatedAt,
            Votes:       int32(tag.Votes),
        })
    }

    var total int64
    s.db.Model(&storage.Tag{}).Count(&total)

    return &scan.TagListResponse{
        Tags:  tagInfos,
        Total: total,
    }, nil
}

func (s *ScanService) InsertTag(ctx context.Context, req *scan.TagInsertRequest) (*scan.TagResponse, error) {
    // Verify signature (implement proper verification)
    if !s.verifySignature(req.Address, req.Signature, req.Tag) {
        return &scan.TagResponse{
            Success: false,
            Message: "Invalid signature",
        }, nil
    }

    tag := storage.Tag{
        Address:     req.Address,
        Tag:         req.Tag,
        Description: req.Description,
        Owner:       req.Owner,
        CreatedAt:   time.Now().Unix(),
        Votes:       0,
    }

    result := s.db.Create(&tag)
    if result.Error != nil {
        return &scan.TagResponse{
            Success: false,
            Message: result.Error.Error(),
        }, nil
    }

    return &scan.TagResponse{
        Success: true,
        Message: "Tag created successfully",
        Id:      int32(tag.ID),
    }, nil
}

func (s *ScanService) Search(ctx context.Context, req *scan.SearchRequest) (*scan.SearchResponse, error) {
    query := req.Query
    var results []*scan.SearchResult

    // Search by block number
    if num, err := strconv.ParseInt(query, 10, 64); err == nil {
        results = append(results, &scan.SearchResult{
            Type:        "block",
            Id:          query,
            Name:        fmt.Sprintf("Block #%d", num),
            Url:         fmt.Sprintf("#/block/%d", num),
            Description: fmt.Sprintf("Block at height %d", num),
        })
    }

    // Search by transaction hash (64 chars hex)
    if len(query) == 64 {
        results = append(results, &scan.SearchResult{
            Type:        "transaction",
            Id:          query,
            Name:        fmt.Sprintf("Transaction %s", query[:8]),
            Url:         fmt.Sprintf("#/transaction/%s", query),
            Description: fmt.Sprintf("Transaction with hash %s", query),
        })
    }

    // Search by address (34 chars base58)
    if len(query) == 34 {
        results = append(results, &scan.SearchResult{
            Type:        "address",
            Id:          query,
            Name:        fmt.Sprintf("Address %s", query[:8]),
            Url:         fmt.Sprintf("#/address/%s", query),
            Description: fmt.Sprintf("Account with address %s", query),
        })
    }

    // Search tokens by name/symbol
    var tokens []storage.LRC20Token
    s.db.Where("name ILIKE ? OR symbol ILIKE ?", "%"+query+"%", "%"+query+"%").Limit(5).Find(&tokens)
    for _, token := range tokens {
        results = append(results, &scan.SearchResult{
            Type:        "token",
            Id:          token.Contract,
            Name:        token.Symbol,
            Url:         fmt.Sprintf("#/token/%s", token.Contract),
            Description: token.Name,
        })
    }

    return &scan.SearchResponse{Results: results}, nil
}

// Helper methods
func (s *ScanService) getMarketData() (price, marketCap, volume int64) {
    // Use CoinMarketCap API via proxy
    resp, err := s.httpClient.Get("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=LIND&convert=USD")
    if err != nil {
        return 0, 0, 0
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    // Parse response (simplified)
    if data, ok := result["data"].(map[string]interface{}); ok {
        if lindData, ok := data["LIND"].(map[string]interface{}); ok {
            if quote, ok := lindData["quote"].(map[string]interface{}); ok {
                if usd, ok := quote["USD"].(map[string]interface{}); ok {
                    price = int64(usd["price"].(float64) * 1000000) // Store as micro
                    marketCap = int64(usd["market_cap"].(float64))
                    volume = int64(usd["volume_24h"].(float64))
                }
            }
        }
    }

    return price, marketCap, volume
}

func (s *ScanService) geolocateIP(ip string) (country, city string, lat, lng float64) {
    // Implement IP geolocation (use a service like ipapi.co or MaxMind)
    // This is a placeholder
    return "Unknown", "Unknown", 0, 0
}

func (s *ScanService) verifySignature(address, signature, data string) bool {
    // Implement signature verification
    // This should match the frontend's signing method
    return true
}

func (s *ScanService) getTotalTransactions(ctx context.Context) int64 {
    // Get from database or node
    return 0
}

func (s *ScanService) getTotalAccounts(ctx context.Context) int64 {
    return 0
}

func (s *ScanService) getTotalContracts(ctx context.Context) int64 {
    return 0
}

func (s *ScanService) getTotalTokens(ctx context.Context) int64 {
    return 0
}