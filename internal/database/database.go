package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LicenseActivation 记录许可证激活信息
type LicenseActivation struct {
	ID          int       `json:"id"`
	Customer    string    `json:"customer"`
	Fingerprint string    `json:"fingerprint"`
	License     string    `json:"license"`
	Description string    `json:"description"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ActivatedAt time.Time `json:"activated_at"`
	IsActive    bool      `json:"is_active"`
}

// DB 数据库连接
type DB struct {
	conn *sql.DB
}

// NewDB 创建新的数据库连接
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	db := &DB{conn: conn}

	// 创建表
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	// 运行数据库迁移
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return db, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// createTables 创建数据库表
func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS license_activations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer TEXT NOT NULL,
		fingerprint TEXT NOT NULL,
		license TEXT NOT NULL,
		description TEXT DEFAULT '',
		issued_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		activated_at INTEGER NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		UNIQUE(fingerprint, license)
	);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create license_activations table: %v", err)
	}

	// 创建索引以提高查询性能
	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_fingerprint ON license_activations(fingerprint);
	CREATE INDEX IF NOT EXISTS idx_customer ON license_activations(customer);
	CREATE INDEX IF NOT EXISTS idx_is_active ON license_activations(is_active);
	`

	_, err = db.conn.Exec(indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	return nil
}

// runMigrations 运行数据库迁移
func (db *DB) runMigrations() error {
	// 创建迁移表
	migrationTableQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT NOT NULL UNIQUE,
		applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);
	`

	_, err := db.conn.Exec(migrationTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %v", err)
	}

	// 检查是否需要添加description字段
	var hasDescriptionColumn bool
	checkColumnQuery := `PRAGMA table_info(license_activations);`
	rows, err := db.conn.Query(checkColumnQuery)
	if err != nil {
		return fmt.Errorf("failed to check table columns: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return fmt.Errorf("failed to scan column info: %v", err)
		}

		if name == "description" {
			hasDescriptionColumn = true
			break
		}
	}

	// 如果没有description字段，则添加
	if !hasDescriptionColumn {
		migrationQuery := `ALTER TABLE license_activations ADD COLUMN description TEXT DEFAULT '';`
		_, err := db.conn.Exec(migrationQuery)
		if err != nil {
			return fmt.Errorf("failed to add description column: %v", err)
		}

		// 记录迁移
		insertMigrationQuery := `INSERT INTO schema_migrations (version) VALUES ('add_description_column');`
		_, err = db.conn.Exec(insertMigrationQuery)
		if err != nil {
			return fmt.Errorf("failed to record migration: %v", err)
		}

		log.Println("Database migration completed: Added description column to license_activations table")
	}

	return nil
}

// InsertLicenseActivation 插入许可证激活记录
func (db *DB) InsertLicenseActivation(activation *LicenseActivation) error {
	query := `
	INSERT INTO license_activations 
	(customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		activation.Customer,
		activation.Fingerprint,
		activation.License,
		activation.Description,
		activation.IssuedAt.Unix(),
		activation.ExpiresAt.Unix(),
		activation.ActivatedAt.Unix(),
		activation.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to insert license activation: %v", err)
	}

	return nil
}

// GetLicenseActivationByFingerprint 根据指纹获取许可证激活记录
func (db *DB) GetLicenseActivationByFingerprint(fingerprint string) (*LicenseActivation, error) {
	query := `
	SELECT id, customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active
	FROM license_activations
	WHERE fingerprint = ? AND is_active = 1
	ORDER BY activated_at DESC
	LIMIT 1
	`

	var activation LicenseActivation
	var issuedAt, expiresAt, activatedAt int64

	err := db.conn.QueryRow(query, fingerprint).Scan(
		&activation.ID,
		&activation.Customer,
		&activation.Fingerprint,
		&activation.License,
		&activation.Description,
		&issuedAt,
		&expiresAt,
		&activatedAt,
		&activation.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get license activation: %v", err)
	}

	activation.IssuedAt = time.Unix(issuedAt, 0)
	activation.ExpiresAt = time.Unix(expiresAt, 0)
	activation.ActivatedAt = time.Unix(activatedAt, 0)

	return &activation, nil
}

// GetActiveLicenseActivationByFingerprint 根据指纹获取有效的许可证激活记录
func (db *DB) GetActiveLicenseActivationByFingerprint(fingerprint string) (*LicenseActivation, error) {
	query := `
	SELECT id, customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active
	FROM license_activations
	WHERE fingerprint = ? AND is_active = 1 AND expires_at > ?
	ORDER BY activated_at DESC
	LIMIT 1
	`

	var activation LicenseActivation
	var issuedAt, expiresAt, activatedAt int64

	err := db.conn.QueryRow(query, fingerprint, time.Now().Unix()).Scan(
		&activation.ID,
		&activation.Customer,
		&activation.Fingerprint,
		&activation.License,
		&activation.Description,
		&issuedAt,
		&expiresAt,
		&activatedAt,
		&activation.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active license activation: %v", err)
	}

	activation.IssuedAt = time.Unix(issuedAt, 0)
	activation.ExpiresAt = time.Unix(expiresAt, 0)
	activation.ActivatedAt = time.Unix(activatedAt, 0)

	return &activation, nil
}

// GetLicenseActivationsWithPagination 分页获取许可证激活记录
func (db *DB) GetLicenseActivationsWithPagination(page, pageSize int) ([]LicenseActivation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询总数
	var total int64
	totalQuery := `SELECT COUNT(*) FROM license_activations`
	err := db.conn.QueryRow(totalQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count license activations: %v", err)
	}

	// 分页查询
	query := `
	SELECT id, customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active
	FROM license_activations
	ORDER BY activated_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := db.conn.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query license activations with pagination: %v", err)
	}
	defer rows.Close()

	var activations []LicenseActivation

	for rows.Next() {
		var activation LicenseActivation
		var issuedAt, expiresAt, activatedAt int64

		err := rows.Scan(
			&activation.ID,
			&activation.Customer,
			&activation.Fingerprint,
			&activation.License,
			&activation.Description,
			&issuedAt,
			&expiresAt,
			&activatedAt,
			&activation.IsActive,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan license activation: %v", err)
		}

		activation.IssuedAt = time.Unix(issuedAt, 0)
		activation.ExpiresAt = time.Unix(expiresAt, 0)
		activation.ActivatedAt = time.Unix(activatedAt, 0)

		activations = append(activations, activation)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating license activations: %v", err)
	}

	return activations, total, nil
}

// GetAllLicenseActivations 获取所有许可证激活记录（兼容旧版）
func (db *DB) GetAllLicenseActivations() ([]LicenseActivation, error) {
	query := `
	SELECT id, customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active
	FROM license_activations
	ORDER BY activated_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query license activations: %v", err)
	}
	defer rows.Close()

	var activations []LicenseActivation

	for rows.Next() {
		var activation LicenseActivation
		var issuedAt, expiresAt, activatedAt int64

		err := rows.Scan(
			&activation.ID,
			&activation.Customer,
			&activation.Fingerprint,
			&activation.License,
			&activation.Description,
			&issuedAt,
			&expiresAt,
			&activatedAt,
			&activation.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan license activation: %v", err)
		}

		activation.IssuedAt = time.Unix(issuedAt, 0)
		activation.ExpiresAt = time.Unix(expiresAt, 0)
		activation.ActivatedAt = time.Unix(activatedAt, 0)

		activations = append(activations, activation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating license activations: %v", err)
	}

	return activations, nil
}

// DeactivateLicense 将许可证标记为非活动状态
func (db *DB) DeactivateLicense(id int) error {
	query := `UPDATE license_activations SET is_active = 0 WHERE id = ?`

	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate license: %v", err)
	}

	return nil
}

// GetExpiredLicenses 获取已过期的许可证
func (db *DB) GetExpiredLicenses() ([]LicenseActivation, error) {
	query := `
	SELECT id, customer, fingerprint, license, description, issued_at, expires_at, activated_at, is_active
	FROM license_activations
	WHERE expires_at < ? AND is_active = 1
	`

	rows, err := db.conn.Query(query, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query expired licenses: %v", err)
	}
	defer rows.Close()

	var activations []LicenseActivation

	for rows.Next() {
		var activation LicenseActivation
		var issuedAt, expiresAt, activatedAt int64

		err := rows.Scan(
			&activation.ID,
			&activation.Customer,
			&activation.Fingerprint,
			&activation.License,
			&activation.Description,
			&issuedAt,
			&expiresAt,
			&activatedAt,
			&activation.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan license activation: %v", err)
		}

		activation.IssuedAt = time.Unix(issuedAt, 0)
		activation.ExpiresAt = time.Unix(expiresAt, 0)
		activation.ActivatedAt = time.Unix(activatedAt, 0)

		activations = append(activations, activation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired licenses: %v", err)
	}

	return activations, nil
}

// CleanupExpiredLicenses 将已过期的许可证标记为非活动状态
func (db *DB) CleanupExpiredLicenses() error {
	query := `UPDATE license_activations SET is_active = 0 WHERE expires_at < ? AND is_active = 1`

	result, err := db.conn.Exec(query, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired licenses: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected > 0 {
		log.Printf("Deactivated %d expired licenses", rowsAffected)
	}

	return nil
}

// DeleteLicenseActivation 删除许可证激活记录
func (db *DB) DeleteLicenseActivation(id int) error {
	query := `DELETE FROM license_activations WHERE id = ?`

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete license activation: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no license activation found with id %d", id)
	}

	return nil
}
