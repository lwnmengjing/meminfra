package model

import (
	"time"
)

type Resource struct {
	ID           uint      `gorm:"primaryKey"`
	ResourceKey  string    `gorm:"uniqueIndex;not null"`
	Kind         string    `gorm:"not null"`
	Hostname     string    `gorm:"index"`
	IPv4         string    `gorm:"column:ipv4;index"`
	IPv6         string    `gorm:"column:ipv6;index"`
	Provider     string    `gorm:"index"`
	Region       string    `gorm:"index"`
	Source       string    `gorm:"index"`
	MetadataJSON string    `gorm:"column:metadata_json;type:text"`
	FirstSeen    time.Time `gorm:"not null"`
	LastSeen     time.Time `gorm:"not null;index"`
}

type Observation struct {
	ID           uint      `gorm:"primaryKey"`
	ResourceID   uint      `gorm:"not null;index"`
	Resource     Resource  `gorm:"constraint:OnDelete:CASCADE"`
	Metric       string    `gorm:"not null;index"`
	Value        float64   `gorm:"not null"`
	Unit         string    `gorm:"index"`
	Source       string    `gorm:"index"`
	MetadataJSON string    `gorm:"column:metadata_json;type:text"`
	ObservedAt   time.Time `gorm:"not null;index"`
}

type Event struct {
	ID            uint      `gorm:"primaryKey"`
	ResourceID    uint      `gorm:"not null;index"`
	Resource      Resource  `gorm:"constraint:OnDelete:CASCADE"`
	EventType     string    `gorm:"not null;index"`
	EventDataJSON string    `gorm:"column:event_data_json;type:text"`
	Source        string    `gorm:"index"`
	CreatedAt     time.Time `gorm:"not null;index"`
}

type MemoryDocument struct {
	ID        uint      `gorm:"primaryKey"`
	DocType   string    `gorm:"not null;index:idx_memory_ref,unique"`
	RefID     uint      `gorm:"not null;index:idx_memory_ref,unique"`
	Title     string    `gorm:"not null"`
	Body      string    `gorm:"not null"`
	Tags      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

type SearchResult struct {
	ID      uint
	DocType string
	RefID   uint
	Title   string
	Body    string
	Tags    string
	Rank    float64
}
