package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"embroidery-designs/internal/utils"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Repository struct {
	db     *Postgres
	logger *zap.Logger
}

func NewRepository(db *Postgres) *Repository {
	return &Repository{
		db:     db,
		logger: utils.GetLogger(),
	}
}

// Task operations
func (r *Repository) CreateTask(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (name, url, type, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Status = TaskStatusPending

	err := r.db.pool.QueryRow(ctx, query,
		task.Name, task.URL, task.Type, task.Status, task.Config, task.CreatedAt, task.UpdatedAt,
	).Scan(&task.ID)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *Repository) GetTask(ctx context.Context, id int64) (*Task, error) {
	query := `
		SELECT id, name, url, type, status, config, created_at, updated_at, started_at, completed_at
		FROM tasks
		WHERE id = $1
	`

	task := &Task{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Name, &task.URL, &task.Type, &task.Status, &task.Config,
		&task.CreatedAt, &task.UpdatedAt, &task.StartedAt, &task.CompletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (r *Repository) ListTasks(ctx context.Context, limit, offset int) ([]*Task, error) {
	query := `
		SELECT id, name, url, type, status, config, created_at, updated_at, started_at, completed_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(
			&task.ID, &task.Name, &task.URL, &task.Type, &task.Status, &task.Config,
			&task.CreatedAt, &task.UpdatedAt, &task.StartedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, id int64, status TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.pool.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (r *Repository) UpdateTask(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks
		SET name = $1, url = $2, config = $3, updated_at = $4
		WHERE id = $5
	`

	task.UpdatedAt = time.Now()
	_, err := r.db.pool.Exec(ctx, query, task.Name, task.URL, task.Config, task.UpdatedAt, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (r *Repository) UpdateTaskConfig(ctx context.Context, id int64, config string) error {
	query := `
		UPDATE tasks
		SET config = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.pool.Exec(ctx, query, config, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task config: %w", err)
	}

	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// CrawlResult operations
func (r *Repository) CreateCrawlResult(ctx context.Context, result *CrawlResult) error {
	query := `
		INSERT INTO crawl_results (task_id, url, method, status_code, headers, body, response_time, proxy_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	result.CreatedAt = time.Now()

	err := r.db.pool.QueryRow(ctx, query,
		result.TaskID, result.URL, result.Method, result.StatusCode,
		result.Headers, result.Body, result.ResponseTime, result.ProxyUsed, result.CreatedAt,
	).Scan(&result.ID)

	if err != nil {
		return fmt.Errorf("failed to create crawl result: %w", err)
	}

	return nil
}

func (r *Repository) GetCrawlResults(ctx context.Context, taskID int64, limit, offset int) ([]*CrawlResult, error) {
	query := `
		SELECT id, task_id, url, method, status_code, headers, body, response_time, proxy_used, created_at
		FROM crawl_results
		WHERE task_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.pool.Query(ctx, query, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get crawl results: %w", err)
	}
	defer rows.Close()

	var results []*CrawlResult
	for rows.Next() {
		result := &CrawlResult{}
		err := rows.Scan(
			&result.ID, &result.TaskID, &result.URL, &result.Method, &result.StatusCode,
			&result.Headers, &result.Body, &result.ResponseTime, &result.ProxyUsed, &result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan crawl result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *Repository) DeleteCrawlResults(ctx context.Context, taskID int64) error {
	query := `DELETE FROM crawl_results WHERE task_id = $1`

	_, err := r.db.pool.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete crawl results: %w", err)
	}

	return nil
}

// Proxy operations
func (r *Repository) CreateProxy(ctx context.Context, proxy *Proxy) error {
	query := `
		INSERT INTO proxies (host, port, type, username, password, is_active, failure_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now()
	proxy.CreatedAt = now
	proxy.UpdatedAt = now
	if !proxy.IsActive {
		proxy.IsActive = true
	}

	err := r.db.pool.QueryRow(ctx, query,
		proxy.Host, proxy.Port, proxy.Type, proxy.Username, proxy.Password,
		proxy.IsActive, proxy.FailureCount, proxy.CreatedAt, proxy.UpdatedAt,
	).Scan(&proxy.ID)

	if err != nil {
		return fmt.Errorf("failed to create proxy: %w", err)
	}

	return nil
}

func (r *Repository) ListProxies(ctx context.Context) ([]*Proxy, error) {
	query := `
		SELECT id, host, port, type, username, password, is_active, failure_count, last_checked, created_at, updated_at
		FROM proxies
		ORDER BY created_at DESC
	`

	rows, err := r.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list proxies: %w", err)
	}
	defer rows.Close()

	var proxies []*Proxy
	for rows.Next() {
		proxy := &Proxy{}
		err := rows.Scan(
			&proxy.ID, &proxy.Host, &proxy.Port, &proxy.Type,
			&proxy.Username, &proxy.Password, &proxy.IsActive, &proxy.FailureCount,
			&proxy.LastChecked, &proxy.CreatedAt, &proxy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %w", err)
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func (r *Repository) GetActiveProxies(ctx context.Context) ([]*Proxy, error) {
	query := `
		SELECT id, host, port, type, username, password, is_active, failure_count, last_checked, created_at, updated_at
		FROM proxies
		WHERE is_active = true
		ORDER BY failure_count ASC, last_checked ASC NULLS FIRST
	`

	rows, err := r.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active proxies: %w", err)
	}
	defer rows.Close()

	var proxies []*Proxy
	for rows.Next() {
		proxy := &Proxy{}
		err := rows.Scan(
			&proxy.ID, &proxy.Host, &proxy.Port, &proxy.Type,
			&proxy.Username, &proxy.Password, &proxy.IsActive, &proxy.FailureCount,
			&proxy.LastChecked, &proxy.CreatedAt, &proxy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %w", err)
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func (r *Repository) UpdateProxyHealth(ctx context.Context, id int64, isHealthy bool) error {
	var query string
	if isHealthy {
		query = `
			UPDATE proxies
			SET is_active = true,
				failure_count = 0,
				last_checked = $1,
				updated_at = $1
			WHERE id = $2
		`
	} else {
		query = `
			UPDATE proxies
			SET is_active = false,
				failure_count = failure_count + 1,
				last_checked = $1,
				updated_at = $1
			WHERE id = $2
		`
	}

	_, err := r.db.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update proxy health: %w", err)
	}

	return nil
}

func (r *Repository) DeleteProxy(ctx context.Context, id int64) error {
	query := `DELETE FROM proxies WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete proxy: %w", err)
	}

	return nil
}

// CrawlLog operations
func (r *Repository) CreateLog(ctx context.Context, log *CrawlLog) error {
	query := `
		INSERT INTO crawl_logs (task_id, level, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	log.CreatedAt = time.Now()

	err := r.db.pool.QueryRow(ctx, query,
		log.TaskID, log.Level, log.Message, log.Metadata, log.CreatedAt,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

func (r *Repository) GetLogs(ctx context.Context, taskID *int64, limit, offset int) ([]*CrawlLog, error) {
	var query string
	var args []interface{}

	if taskID != nil {
		query = `
			SELECT id, task_id, level, message, metadata, created_at
			FROM crawl_logs
			WHERE task_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*taskID, limit, offset}
	} else {
		query = `
			SELECT id, task_id, level, message, metadata, created_at
			FROM crawl_logs
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var logs []*CrawlLog
	for rows.Next() {
		log := &CrawlLog{}
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.Level, &log.Message, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// Product operations
func (r *Repository) UpsertProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (
			elastic_id, product_id, item_id, name, brand, catalog, artist,
			rating, list_price, sale_price, club_price, sale_rank, customer_interest_index,
			in_stock, is_active, is_buyable, licensed, is_applique, is_cross_stitch,
			is_pdf_available, is_fsl, is_heat_transfer, is_design_used_in_project,
			in_custom_pack, definition_name, product_type, gtin, color_sequence,
			design_keywords, categories, categories_list, keywords, sales, sales_list,
			sale_end_date, year_created, applied_discount_id, is_multiple_variants_available,
			variants, raw_data, status, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39, $40, $41, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (elastic_id) 
		DO UPDATE SET
			product_id = EXCLUDED.product_id,
			item_id = EXCLUDED.item_id,
			name = EXCLUDED.name,
			brand = EXCLUDED.brand,
			catalog = EXCLUDED.catalog,
			artist = EXCLUDED.artist,
			rating = EXCLUDED.rating,
			list_price = EXCLUDED.list_price,
			sale_price = EXCLUDED.sale_price,
			club_price = EXCLUDED.club_price,
			sale_rank = EXCLUDED.sale_rank,
			customer_interest_index = EXCLUDED.customer_interest_index,
			in_stock = EXCLUDED.in_stock,
			is_active = EXCLUDED.is_active,
			is_buyable = EXCLUDED.is_buyable,
			licensed = EXCLUDED.licensed,
			is_applique = EXCLUDED.is_applique,
			is_cross_stitch = EXCLUDED.is_cross_stitch,
			is_pdf_available = EXCLUDED.is_pdf_available,
			is_fsl = EXCLUDED.is_fsl,
			is_heat_transfer = EXCLUDED.is_heat_transfer,
			is_design_used_in_project = EXCLUDED.is_design_used_in_project,
			in_custom_pack = EXCLUDED.in_custom_pack,
			definition_name = EXCLUDED.definition_name,
			product_type = EXCLUDED.product_type,
			gtin = EXCLUDED.gtin,
			color_sequence = EXCLUDED.color_sequence,
			design_keywords = EXCLUDED.design_keywords,
			categories = EXCLUDED.categories,
			categories_list = EXCLUDED.categories_list,
			keywords = EXCLUDED.keywords,
			sales = EXCLUDED.sales,
			sales_list = EXCLUDED.sales_list,
			sale_end_date = EXCLUDED.sale_end_date,
			year_created = EXCLUDED.year_created,
			applied_discount_id = EXCLUDED.applied_discount_id,
			is_multiple_variants_available = EXCLUDED.is_multiple_variants_available,
			variants = EXCLUDED.variants,
			raw_data = EXCLUDED.raw_data,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	if !product.Status.IsValid() {
		product.Status = ProductStatusPending
	}

	err := r.db.pool.QueryRow(ctx, query,
		product.ElasticID, product.ProductID, product.ItemID, product.Name,
		product.Brand, product.Catalog, product.Artist, product.Rating,
		product.ListPrice, product.SalePrice, product.ClubPrice, product.SaleRank,
		product.CustomerInterestIndex, product.InStock, product.IsActive,
		product.IsBuyable, product.Licensed, product.IsApplique, product.IsCrossStitch,
		product.IsPDFAvailable, product.IsFSL, product.IsHeatTransfer,
		product.IsDesignUsedInProject, product.InCustomPack, product.DefinitionName,
		product.ProductType, product.GTIN, product.ColorSequence, product.DesignKeywords,
		product.Categories, product.CategoriesList, product.Keywords, product.Sales,
		product.SalesList, product.SaleEndDate, product.YearCreated,
		product.AppliedDiscountID, product.IsMultipleVariantsAvailable,
		product.Variants, product.RawData, product.Status,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert product: %w", err)
	}

	return nil
}

func (r *Repository) GetProduct(ctx context.Context, id int64) (*Product, error) {
	query := `
		SELECT id, elastic_id, product_id, item_id, name, brand, catalog, artist,
			rating, list_price, sale_price, club_price, sale_rank, customer_interest_index,
			in_stock, is_active, is_buyable, licensed, is_applique, is_cross_stitch,
			is_pdf_available, is_fsl, is_heat_transfer, is_design_used_in_project,
			in_custom_pack, definition_name, product_type, gtin, color_sequence,
			design_keywords, categories, categories_list, keywords, sales, sales_list,
			sale_end_date, year_created, applied_discount_id, is_multiple_variants_available,
			variants, raw_data, status, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &Product{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.ElasticID, &product.ProductID, &product.ItemID,
		&product.Name, &product.Brand, &product.Catalog, &product.Artist,
		&product.Rating, &product.ListPrice, &product.SalePrice, &product.ClubPrice,
		&product.SaleRank, &product.CustomerInterestIndex, &product.InStock,
		&product.IsActive, &product.IsBuyable, &product.Licensed, &product.IsApplique,
		&product.IsCrossStitch, &product.IsPDFAvailable, &product.IsFSL,
		&product.IsHeatTransfer, &product.IsDesignUsedInProject, &product.InCustomPack,
		&product.DefinitionName, &product.ProductType, &product.GTIN,
		&product.ColorSequence, &product.DesignKeywords, &product.Categories,
		&product.CategoriesList, &product.Keywords, &product.Sales, &product.SalesList,
		&product.SaleEndDate, &product.YearCreated, &product.AppliedDiscountID,
		&product.IsMultipleVariantsAvailable, &product.Variants, &product.RawData,
		&product.Status,
		&product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (r *Repository) GetProductByElasticID(ctx context.Context, elasticID string) (*Product, error) {
	query := `
		SELECT id, elastic_id, product_id, item_id, name, brand, catalog, artist,
			rating, list_price, sale_price, club_price, sale_rank, customer_interest_index,
			in_stock, is_active, is_buyable, licensed, is_applique, is_cross_stitch,
			is_pdf_available, is_fsl, is_heat_transfer, is_design_used_in_project,
			in_custom_pack, definition_name, product_type, gtin, color_sequence,
			design_keywords, categories, categories_list, keywords, sales, sales_list,
			sale_end_date, year_created, applied_discount_id, is_multiple_variants_available,
			variants, raw_data, status, created_at, updated_at
		FROM products
		WHERE elastic_id = $1
	`

	product := &Product{}
	err := r.db.pool.QueryRow(ctx, query, elasticID).Scan(
		&product.ID, &product.ElasticID, &product.ProductID, &product.ItemID,
		&product.Name, &product.Brand, &product.Catalog, &product.Artist,
		&product.Rating, &product.ListPrice, &product.SalePrice, &product.ClubPrice,
		&product.SaleRank, &product.CustomerInterestIndex, &product.InStock,
		&product.IsActive, &product.IsBuyable, &product.Licensed, &product.IsApplique,
		&product.IsCrossStitch, &product.IsPDFAvailable, &product.IsFSL,
		&product.IsHeatTransfer, &product.IsDesignUsedInProject, &product.InCustomPack,
		&product.DefinitionName, &product.ProductType, &product.GTIN,
		&product.ColorSequence, &product.DesignKeywords, &product.Categories,
		&product.CategoriesList, &product.Keywords, &product.Sales, &product.SalesList,
		&product.SaleEndDate, &product.YearCreated, &product.AppliedDiscountID,
		&product.IsMultipleVariantsAvailable, &product.Variants, &product.RawData,
		&product.Status,
		&product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (r *Repository) ListProducts(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*Product, int64, error) {
	// Build WHERE clause
	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if brand, ok := filters["brand"]; ok && brand != nil {
		whereClause += fmt.Sprintf(" AND brand = $%d", argIndex)
		args = append(args, brand)
		argIndex++
	}

	if catalog, ok := filters["catalog"]; ok && catalog != nil {
		whereClause += fmt.Sprintf(" AND catalog = $%d", argIndex)
		args = append(args, catalog)
		argIndex++
	}

	if inStock, ok := filters["in_stock"]; ok && inStock != nil {
		whereClause += fmt.Sprintf(" AND in_stock = $%d", argIndex)
		args = append(args, inStock)
		argIndex++
	}

	if search, ok := filters["search"]; ok && search != nil {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR design_keywords ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search.(string)+"%")
		argIndex++
	}

	if statuses, ok := filters["statuses"]; ok && statuses != nil {
		statusList, _ := statuses.([]ProductStatus)
		if len(statusList) > 0 {
			whereClause += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			stringStatuses := make([]string, 0, len(statusList))
			for _, st := range statusList {
				stringStatuses = append(stringStatuses, string(st))
			}
			args = append(args, stringStatuses)
			argIndex++
		}
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products WHERE %s", whereClause)
	var total int64
	err := r.db.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get products
	query := fmt.Sprintf(`
		SELECT id, elastic_id, product_id, item_id, name, brand, catalog, artist,
			rating, list_price, sale_price, club_price, sale_rank, customer_interest_index,
			in_stock, is_active, is_buyable, licensed, is_applique, is_cross_stitch,
			is_pdf_available, is_fsl, is_heat_transfer, is_design_used_in_project,
			in_custom_pack, definition_name, product_type, gtin, color_sequence,
			design_keywords, categories, categories_list, keywords, sales, sales_list,
			sale_end_date, year_created, applied_discount_id, is_multiple_variants_available,
			variants, raw_data, status, created_at, updated_at
		FROM products
		WHERE %s
		ORDER BY sale_rank DESC NULLS LAST, rating DESC NULLS LAST, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		err := rows.Scan(
			&product.ID, &product.ElasticID, &product.ProductID, &product.ItemID,
			&product.Name, &product.Brand, &product.Catalog, &product.Artist,
			&product.Rating, &product.ListPrice, &product.SalePrice, &product.ClubPrice,
			&product.SaleRank, &product.CustomerInterestIndex, &product.InStock,
			&product.IsActive, &product.IsBuyable, &product.Licensed, &product.IsApplique,
			&product.IsCrossStitch, &product.IsPDFAvailable, &product.IsFSL,
			&product.IsHeatTransfer, &product.IsDesignUsedInProject, &product.InCustomPack,
			&product.DefinitionName, &product.ProductType, &product.GTIN,
			&product.ColorSequence, &product.DesignKeywords, &product.Categories,
			&product.CategoriesList, &product.Keywords, &product.Sales, &product.SalesList,
			&product.SaleEndDate, &product.YearCreated, &product.AppliedDiscountID,
			&product.IsMultipleVariantsAvailable, &product.Variants, &product.RawData,
			&product.Status,
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	return products, total, nil
}

func (r *Repository) DeleteProduct(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (r *Repository) UpdateProductStatus(ctx context.Context, id int64, status ProductStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid product status: %s", status)
	}

	commandTag, err := r.db.pool.Exec(ctx, `
		UPDATE products
		SET status = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update product status: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *Repository) GetProductStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total products
	var total int64
	err := r.db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total"] = total
	stats["total_products"] = total

	// In stock count
	var inStock int64
	err = r.db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE in_stock = true").Scan(&inStock)
	if err != nil {
		return nil, fmt.Errorf("failed to get in stock count: %w", err)
	}
	stats["in_stock"] = inStock

	// Brands count
	var brandsCount int64
	err = r.db.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT brand) FROM products WHERE brand IS NOT NULL").Scan(&brandsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get brands count: %w", err)
	}
	stats["brands_count"] = brandsCount

	// Status distribution
	rows, err := r.db.pool.Query(ctx, "SELECT status, COUNT(*) FROM products GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status breakdown: %w", err)
	}
	defer rows.Close()

	statusBreakdown := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if scanErr := rows.Scan(&status, &count); scanErr != nil {
			return nil, fmt.Errorf("failed to scan status breakdown: %w", scanErr)
		}
		statusBreakdown[status] = count
	}
	stats["status_breakdown"] = statusBreakdown

	return stats, nil
}

// User operations
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &User{}
	err := r.db.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.pool.QueryRow(ctx, query,
		user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// API Token operations
func (r *Repository) CreateAPIToken(ctx context.Context, token *APIToken) error {
	query := `
		INSERT INTO api_tokens (user_id, token_name, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	token.CreatedAt = time.Now()

	err := r.db.pool.QueryRow(ctx, query,
		token.UserID, token.TokenName, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create API token: %w", err)
	}

	return nil
}

func (r *Repository) GetAPITokenByHash(ctx context.Context, tokenHash string) (*APIToken, error) {
	query := `
		SELECT id, user_id, token_name, token_hash, expires_at, created_at, last_used_at
		FROM api_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	token := &APIToken{}
	err := r.db.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenName, &token.TokenHash,
		&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("token not found or expired: %w", err)
		}
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}

	return token, nil
}

func (r *Repository) UpdateTokenLastUsed(ctx context.Context, tokenID int64) error {
	query := `
		UPDATE api_tokens
		SET last_used_at = $1
		WHERE id = $2
	`

	_, err := r.db.pool.Exec(ctx, query, time.Now(), tokenID)
	if err != nil {
		return fmt.Errorf("failed to update token last used: %w", err)
	}

	return nil
}

func (r *Repository) ListAPITokens(ctx context.Context, userID int64) ([]*APIToken, error) {
	query := `
		SELECT id, user_id, token_name, token_hash, expires_at, created_at, last_used_at
		FROM api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*APIToken
	for rows.Next() {
		token := &APIToken{}
		err := rows.Scan(
			&token.ID, &token.UserID, &token.TokenName, &token.TokenHash,
			&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API token: %w", err)
		}
		// Don't expose token hash
		token.TokenHash = ""
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *Repository) DeleteAPIToken(ctx context.Context, tokenID int64) error {
	query := `DELETE FROM api_tokens WHERE id = $1`

	_, err := r.db.pool.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete API token: %w", err)
	}

	return nil
}

// Refresh token operations
func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	token.CreatedAt = time.Now()
	err := r.db.pool.QueryRow(ctx, query,
		token.UserID, token.TokenHash, token.UserAgent, token.IPAddress, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *Repository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`

	token := &RefreshToken{}
	err := r.db.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.UserAgent, &token.IPAddress,
		&token.ExpiresAt, &token.CreatedAt, &token.RevokedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found or expired: %w", err)
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE id = $2 AND revoked_at IS NULL
	`

	_, err := r.db.pool.Exec(ctx, query, time.Now(), tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (r *Repository) RevokeRefreshTokensByUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL
	`

	_, err := r.db.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

// Embroidery payload overrides
func (r *Repository) GetEmbroideryPayloadOverrides(ctx context.Context) (map[string]interface{}, *time.Time, error) {
	query := `
		SELECT value, updated_at
		FROM crawler_settings
		WHERE settings_key = $1
	`

	var raw json.RawMessage
	var updatedAt time.Time

	err := r.db.pool.QueryRow(ctx, query, EmbroideryPayloadOverridesKey).Scan(&raw, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return map[string]interface{}{}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to load embroidery payload overrides: %w", err)
	}

	var payload map[string]interface{}
	if len(raw) > 0 {
		if unmarshalErr := json.Unmarshal(raw, &payload); unmarshalErr != nil {
			return nil, nil, fmt.Errorf("failed to parse embroidery payload overrides: %w", unmarshalErr)
		}
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	return payload, &updatedAt, nil
}

func (r *Repository) UpdateEmbroideryPayloadOverrides(ctx context.Context, overrides map[string]interface{}) error {
	if overrides == nil {
		overrides = map[string]interface{}{}
	}

	payloadBytes, err := json.Marshal(overrides)
	if err != nil {
		return fmt.Errorf("failed to marshal embroidery payload overrides: %w", err)
	}

	query := `
		INSERT INTO crawler_settings (settings_key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (settings_key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`

	_, err = r.db.pool.Exec(ctx, query, EmbroideryPayloadOverridesKey, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to upsert embroidery payload overrides: %w", err)
	}

	return nil
}
