package storage

import (
    "time"
)

type Tag struct {
    ID          uint      `gorm:"primarykey"`
    Address     string    `gorm:"index;type:varchar(34)"`
    Tag         string    `gorm:"type:varchar(100)"`
    Description string    `gorm:"type:text"`
    Owner       string    `gorm:"type:varchar(34)"`
    Signature   string    `gorm:"type:text"`
    CreatedAt   int64     `gorm:"index"`
    Votes       int       `gorm:"default:0"`
}

type LRC20Token struct {
    ID          uint      `gorm:"primarykey"`
    Contract    string    `gorm:"uniqueIndex;type:varchar(34)"`
    Name        string    `gorm:"type:varchar(100)"`
    Symbol      string    `gorm:"type:varchar(20)"`
    Decimals    int       `gorm:"default:18"`
    TotalSupply string    `gorm:"type:varchar(100)"`
    Owner       string    `gorm:"type:varchar(34)"`
    IssueTime   int64     `gorm:"index"`
    Holders     int64     `gorm:"default:0"`
    Transfers   string    `gorm:"type:text"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TokenHolder struct {
    ID         uint      `gorm:"primarykey"`
    Contract   string    `gorm:"index;type:varchar(34)"`
    Address    string    `gorm:"index;type:varchar(34)"`
    Balance    string    `gorm:"type:varchar(100)"`
    Percentage float64   `gorm:"type:decimal(10,4)"`
    UpdatedAt  time.Time
}

type Account struct {
    Address      string    `gorm:"primaryKey;type:varchar(34)"`
    Balance      string    `gorm:"type:varchar(100)"`
    Transactions int64     `gorm:"default:0"`
    AccountType  string    `gorm:"type:varchar(20)"`
    Bandwidth    int64     `gorm:"default:0"`
    Energy       int64     `gorm:"default:0"`
    UpdatedAt    time.Time
}

type Statistic struct {
    ID          uint      `gorm:"primarykey"`
    Type        string    `gorm:"index;type:varchar(50)"`
    Value       string    `gorm:"type:text"`
    Timestamp   int64     `gorm:"index"`
}

type MarketData struct {
    ID        uint      `gorm:"primarykey"`
    Pair      string    `gorm:"index;type:varchar(20)"`
    Price     string    `gorm:"type:varchar(100)"`
    Volume24h string    `gorm:"type:varchar(100)"`
    High24h   string    `gorm:"type:varchar(100)"`
    Low24h    string    `gorm:"type:varchar(100)"`
    Change24h string    `gorm:"type:varchar(20)"`
    Timestamp int64     `gorm:"index"`
}