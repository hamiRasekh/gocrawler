package storage

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusPaused    TaskStatus = "paused"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusStopped   TaskStatus = "stopped"
)

type ProductStatus string

const (
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusApproved ProductStatus = "approved"
	ProductStatusRejected ProductStatus = "rejected"
)

const EmbroideryPayloadOverridesKey = "embroidery_payload_overrides"

func (ps ProductStatus) IsValid() bool {
	switch ps {
	case ProductStatusPending, ProductStatusApproved, ProductStatusRejected:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	URL         string     `json:"url" db:"url"`
	Type        string     `json:"type" db:"type"` // "api" or "web"
	Status      TaskStatus `json:"status" db:"status"`
	Config      string     `json:"config" db:"config"` // JSON config
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type CrawlResult struct {
	ID          int64     `json:"id" db:"id"`
	TaskID      int64     `json:"task_id" db:"task_id"`
	URL         string    `json:"url" db:"url"`
	Method      string    `json:"method" db:"method"`
	StatusCode  int       `json:"status_code" db:"status_code"`
	Headers     string    `json:"headers" db:"headers"` // JSON
	Body        string    `json:"body" db:"body"`
	ResponseTime int      `json:"response_time" db:"response_time"` // milliseconds
	ProxyUsed   *string   `json:"proxy_used,omitempty" db:"proxy_used"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Proxy struct {
	ID          int64     `json:"id" db:"id"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	Type        string    `json:"type" db:"type"` // "http", "https", "socks5"
	Username    *string   `json:"username,omitempty" db:"username"`
	Password    *string   `json:"password,omitempty" db:"password"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	FailureCount int      `json:"failure_count" db:"failure_count"`
	LastChecked *time.Time `json:"last_checked,omitempty" db:"last_checked"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CrawlLog struct {
	ID        int64     `json:"id" db:"id"`
	TaskID    *int64    `json:"task_id,omitempty" db:"task_id"`
	Level     string    `json:"level" db:"level"` // "info", "warn", "error"
	Message   string    `json:"message" db:"message"`
	Metadata  string    `json:"metadata" db:"metadata"` // JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Product struct {
	ID                          int64      `json:"id" db:"id"`
	ElasticID                   string     `json:"elastic_id" db:"elastic_id"`
	ProductID                   *string    `json:"product_id,omitempty" db:"product_id"`
	ItemID                      *string    `json:"item_id,omitempty" db:"item_id"`
	Name                        *string    `json:"name,omitempty" db:"name"`
	Brand                       *string    `json:"brand,omitempty" db:"brand"`
	Catalog                     *string    `json:"catalog,omitempty" db:"catalog"`
	Artist                      *string    `json:"artist,omitempty" db:"artist"`
	Rating                      *float64   `json:"rating,omitempty" db:"rating"`
	ListPrice                   *float64   `json:"list_price,omitempty" db:"list_price"`
	SalePrice                   *float64   `json:"sale_price,omitempty" db:"sale_price"`
	ClubPrice                   *float64   `json:"club_price,omitempty" db:"club_price"`
	SaleRank                    *int       `json:"sale_rank,omitempty" db:"sale_rank"`
	CustomerInterestIndex       *int       `json:"customer_interest_index,omitempty" db:"customer_interest_index"`
	InStock                     bool       `json:"in_stock" db:"in_stock"`
	IsActive                    bool       `json:"is_active" db:"is_active"`
	IsBuyable                   bool       `json:"is_buyable" db:"is_buyable"`
	Licensed                    bool       `json:"licensed" db:"licensed"`
	IsApplique                  bool       `json:"is_applique" db:"is_applique"`
	IsCrossStitch               bool       `json:"is_cross_stitch" db:"is_cross_stitch"`
	IsPDFAvailable              bool       `json:"is_pdf_available" db:"is_pdf_available"`
	IsFSL                       bool        `json:"is_fsl" db:"is_fsl"`
	IsHeatTransfer              bool        `json:"is_heat_transfer" db:"is_heat_transfer"`
	IsDesignUsedInProject       bool        `json:"is_design_used_in_project" db:"is_design_used_in_project"`
	InCustomPack                bool        `json:"in_custom_pack" db:"in_custom_pack"`
	DefinitionName              *string     `json:"definition_name,omitempty" db:"definition_name"`
	ProductType                 *string     `json:"product_type,omitempty" db:"product_type"`
	GTIN                        *string     `json:"gtin,omitempty" db:"gtin"`
	ColorSequence               *string     `json:"color_sequence,omitempty" db:"color_sequence"`
	DesignKeywords              *string     `json:"design_keywords,omitempty" db:"design_keywords"`
	Categories                  *string     `json:"categories,omitempty" db:"categories"`
	CategoriesList              *string     `json:"categories_list,omitempty" db:"categories_list"` // JSON
	Keywords                    *string     `json:"keywords,omitempty" db:"keywords"` // JSON
	Sales                       *string     `json:"sales,omitempty" db:"sales"`
	SalesList                   *string     `json:"sales_list,omitempty" db:"sales_list"` // JSON
	SaleEndDate                 *time.Time  `json:"sale_end_date,omitempty" db:"sale_end_date"`
	YearCreated                 *time.Time  `json:"year_created,omitempty" db:"year_created"`
	AppliedDiscountID           *int        `json:"applied_discount_id,omitempty" db:"applied_discount_id"`
	IsMultipleVariantsAvailable bool        `json:"is_multiple_variants_available" db:"is_multiple_variants_available"`
	Variants                    *string     `json:"variants,omitempty" db:"variants"` // JSON
	RawData                     *string     `json:"raw_data,omitempty" db:"raw_data"` // Full JSON
	Status                      ProductStatus `json:"status" db:"status"`
	CreatedAt                   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt                   time.Time   `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type APIToken struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	TokenName  string     `json:"token_name" db:"token_name"`
	TokenHash  string     `json:"-" db:"token_hash"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

type RefreshToken struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	TokenHash  string     `json:"-" db:"token_hash"`
	UserAgent  *string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress  *string    `json:"ip_address,omitempty" db:"ip_address"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

