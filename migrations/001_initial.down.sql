-- Drop indexes
DROP INDEX IF EXISTS idx_crawl_logs_level;
DROP INDEX IF EXISTS idx_crawl_logs_created_at;
DROP INDEX IF EXISTS idx_crawl_logs_task_id;
DROP INDEX IF EXISTS idx_proxies_failure_count;
DROP INDEX IF EXISTS idx_proxies_is_active;
DROP INDEX IF EXISTS idx_crawl_results_created_at;
DROP INDEX IF EXISTS idx_crawl_results_task_id;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_status;

-- Drop tables
DROP TABLE IF EXISTS crawl_logs;
DROP TABLE IF EXISTS proxies;
DROP TABLE IF EXISTS crawl_results;
DROP TABLE IF EXISTS tasks;

