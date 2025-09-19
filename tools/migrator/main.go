package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// 数据库连接信息 - 后续可通过配置文件或环境变量动态获取
	DB_HOST     = "localhost"
	DB_PORT     = "3406"
	DB_USER     = "root"
	DB_PASSWORD = "root123"
	DB_NAME     = "go_template_db"

	// 迁移文件目录
	MIGRATIONS_DIR = "../../db/migrations"
)

// 获取数据库连接字符串
func getDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
}

// 连接数据库
func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接不可用: %v", err)
	}

	return db, nil
}

// 创建迁移文件
func createMigration(name string) error {
	// 确保迁移目录存在
	if err := os.MkdirAll(MIGRATIONS_DIR, 0755); err != nil {
		return fmt.Errorf("创建迁移目录失败: %v", err)
	}

	// 生成时间戳作为版本号
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s", timestamp, name)

	// 创建 up 文件
	upFile := filepath.Join(MIGRATIONS_DIR, filename+".up.sql")
	upContent := fmt.Sprintf("-- Migration: %s (UP)\n-- Version: %s\n-- Created at: %s\n\n", name, timestamp, time.Now().Format("2006-01-02 15:04:05"))
	if err := createMigrationFile(upFile, upContent); err != nil {
		return err
	}

	// 创建 down 文件
	downFile := filepath.Join(MIGRATIONS_DIR, filename+".down.sql")
	downContent := fmt.Sprintf("-- Migration: %s (DOWN)\n-- Version: %s\n-- Created at: %s\n\n", name, timestamp, time.Now().Format("2006-01-02 15:04:05"))
	if err := createMigrationFile(downFile, downContent); err != nil {
		return err
	}

	fmt.Printf("迁移文件创建成功:\n")
	fmt.Printf("  版本号: %s\n", timestamp)
	fmt.Printf("  UP:     %s\n", upFile)
	fmt.Printf("  DOWN:   %s\n", downFile)
	fmt.Printf("\n使用版本号 %s 进行版本跳转:\n", timestamp)
	fmt.Printf("  task migrator:goto -- %s\n", timestamp)

	return nil
}

// 创建迁移文件
func createMigrationFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %v", filename, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭文件失败 %s: %v", filename, err)
		}
	}()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("写入文件失败 %s: %v", filename, err)
	}

	return nil
}

// 执行指定步数的 UP 迁移
func migrateUpSteps(steps int) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	// 创建迁移记录表（如果不存在）
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// 获取所有 up 迁移文件
	files, err := getUpMigrationFiles()
	if err != nil {
		return err
	}

	// 按版本号排序
	sort.Strings(files)

	// 获取已执行的迁移
	executed, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	// 执行未执行的迁移
	count := 0
	for _, file := range files {
		version := extractVersionFromFilename(file)
		if !executed[version] {
			if err := executeMigrationFile(db, file, version, "up"); err != nil {
				return err
			}
			count++
			fmt.Printf("执行迁移: %s (版本: %s)\n", file, version)

			// 如果指定了步数，检查是否已达到
			if steps > 0 && count >= steps {
				break
			}
		}
	}

	if count == 0 {
		fmt.Println("没有待执行的迁移文件")
	} else {
		fmt.Printf("成功执行 %d 个迁移文件\n", count)
	}

	return nil
}

// 执行指定步数的 DOWN 迁移
func migrateDownSteps(steps int) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	// 获取已执行的迁移列表（按版本降序）
	versions, err := getExecutedMigrationVersions(db)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		fmt.Println("没有可回滚的迁移")
		return nil
	}

	// 限制回滚步数
	if steps > len(versions) {
		steps = len(versions)
	}

	count := 0
	for i := 0; i < steps; i++ {
		version := versions[i]

		// 构建 down 文件路径
		var downFile string
		files, err := filepath.Glob(filepath.Join(MIGRATIONS_DIR, version+"_*.down.sql"))
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("找不到版本 %s 的回滚文件", version)
		}

		downFile = files[0]
		downFileName := filepath.Base(downFile)

		// 执行 down 迁移
		if err := executeMigrationFile(db, downFileName, version, "down"); err != nil {
			return err
		}

		// 从迁移记录中删除
		if err := removeMigrationRecord(db, version); err != nil {
			return err
		}

		count++
		fmt.Printf("成功回滚迁移: %s (版本: %s)\n", downFileName, version)
	}

	fmt.Printf("成功回滚 %d 个迁移\n", count)
	return nil
}

// 显示当前数据库版本
func showVersion() error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	// 创建迁移记录表（如果不存在）
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// 获取当前版本
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	if currentVersion == "" {
		fmt.Println("当前数据库版本: 无迁移记录")
	} else {
		fmt.Printf("当前数据库版本: %s\n", currentVersion)
	}

	// 显示迁移历史
	versions, err := getExecutedMigrationVersions(db)
	if err != nil {
		return err
	}

	if len(versions) > 0 {
		fmt.Println("\n已执行的迁移:")
		for i := len(versions) - 1; i >= 0; i-- {
			marker := "  "
			if i == len(versions)-1 {
				marker = "* " // 标记当前版本
			}

			// 查找对应的文件名
			fileName := findMigrationFileName(versions[i])
			if fileName != "" {
				fmt.Printf("%s%s (%s)\n", marker, versions[i], fileName)
			} else {
				fmt.Printf("%s%s\n", marker, versions[i])
			}
		}
	}

	// 显示可用的迁移版本
	fmt.Println("\n可用的迁移版本:")
	allVersions, err := getAllMigrationVersions()
	if err != nil {
		return err
	}

	executed, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	for _, version := range allVersions {
		status := "未执行"
		if executed[version] {
			status = "已执行"
		}
		fileName := findMigrationFileName(version)
		fmt.Printf("  %s - %s (%s)\n", version, status, fileName)
	}

	return nil
}

// 跳转到指定版本
func gotoVersion(targetVersion string) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	// 创建迁移记录表（如果不存在）
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// 获取当前版本
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	// 获取所有可用的迁移文件
	allMigrations, err := getAllMigrationVersions()
	if err != nil {
		return err
	}

	// 检查目标版本是否存在
	targetExists := false
	for _, v := range allMigrations {
		if v == targetVersion {
			targetExists = true
			break
		}
	}

	if !targetExists {
		return fmt.Errorf("目标版本 %s 不存在", targetVersion)
	}

	fmt.Printf("从版本 %s 迁移到版本 %s\n", currentVersion, targetVersion)

	// 比较版本号决定是向上还是向下迁移
	if compareVersions(targetVersion, currentVersion) > 0 {
		// 向上迁移
		return migrateToVersionUp(db, currentVersion, targetVersion)
	} else if compareVersions(targetVersion, currentVersion) < 0 {
		// 向下迁移
		return migrateToVersionDown(db, currentVersion, targetVersion)
	} else {
		fmt.Println("已经在目标版本")
		return nil
	}
}

// 向上迁移到指定版本
func migrateToVersionUp(db *sql.DB, currentVersion, targetVersion string) error {
	// 获取所有 up 迁移文件
	files, err := getUpMigrationFiles()
	if err != nil {
		return err
	}

	sort.Strings(files)

	// 获取已执行的迁移
	executed, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	count := 0
	for _, file := range files {
		version := extractVersionFromFilename(file)

		// 跳过已执行的和超过目标版本的
		if executed[version] || compareVersions(version, targetVersion) > 0 {
			continue
		}

		if err := executeMigrationFile(db, file, version, "up"); err != nil {
			return err
		}

		count++
		fmt.Printf("执行迁移: %s (版本: %s)\n", file, version)

		// 如果达到目标版本就停止
		if version == targetVersion {
			break
		}
	}

	fmt.Printf("成功执行 %d 个迁移，当前版本: %s\n", count, targetVersion)
	return nil
}

// 向下迁移到指定版本
func migrateToVersionDown(db *sql.DB, currentVersion, targetVersion string) error {
	// 获取已执行的迁移列表（按版本降序）
	versions, err := getExecutedMigrationVersions(db)
	if err != nil {
		return err
	}

	count := 0
	for _, version := range versions {
		// 如果当前版本已经达到或低于目标版本，停止
		if compareVersions(version, targetVersion) <= 0 {
			break
		}

		// 构建 down 文件路径
		files, err := filepath.Glob(filepath.Join(MIGRATIONS_DIR, version+"_*.down.sql"))
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("找不到版本 %s 的回滚文件", version)
		}

		downFileName := filepath.Base(files[0])

		// 执行 down 迁移
		if err := executeMigrationFile(db, downFileName, version, "down"); err != nil {
			return err
		}

		// 从迁移记录中删除
		if err := removeMigrationRecord(db, version); err != nil {
			return err
		}

		count++
		fmt.Printf("回滚迁移: %s (版本: %s)\n", downFileName, version)
	}

	fmt.Printf("成功回滚 %d 个迁移，当前版本: %s\n", count, targetVersion)
	return nil
}

// 导入数据
func importData(filePath string) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	ext := filepath.Ext(filePath)
	switch strings.ToLower(ext) {
	case ".sql":
		return importSQL(db, filePath)
	case ".csv":
		return importCSV(db, filePath)
	default:
		return fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// 导出数据
func exportData(format, outputPath string) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	switch strings.ToLower(format) {
	case "sql":
		return exportSQL(db, outputPath)
	case "csv":
		return exportCSV(db, outputPath)
	default:
		return fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// 创建迁移记录表
func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (version)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	_, err := db.Exec(query)
	return err
}

// 获取 UP 迁移文件列表
func getUpMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(MIGRATIONS_DIR, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".up.sql") {
			files = append(files, info.Name())
		}
		return nil
	})

	return files, err
}

// 获取已执行的迁移
func getExecutedMigrations(db *sql.DB) (map[string]bool, error) {
	executed := make(map[string]bool)

	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return executed, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("关闭查询结果集失败: %v", err)
		}
	}()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return executed, err
		}
		executed[version] = true
	}

	return executed, nil
}

// 从文件名提取版本号
func extractVersionFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// 执行迁移文件
func executeMigrationFile(db *sql.DB, filename, version, direction string) error {
	filePath := filepath.Join(MIGRATIONS_DIR, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败 %s: %v", filePath, err)
	}

	// 执行 SQL
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("执行迁移失败 %s: %v", filename, err)
	}

	// 记录迁移（仅对 up 迁移）
	if direction == "up" {
		return recordMigration(db, version)
	}

	return nil
}

// 记录迁移
func recordMigration(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
	return err
}

// 删除迁移记录
func removeMigrationRecord(db *sql.DB, version string) error {
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	return err
}

// 获取最后一个迁移版本
func getLastMigrationVersion(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return version, err
}

// 获取当前版本（与getLastMigrationVersion相同，但语义更清晰）
func getCurrentVersion(db *sql.DB) (string, error) {
	return getLastMigrationVersion(db)
}

// 获取已执行的迁移版本列表（按版本降序）
func getExecutedMigrationVersions(db *sql.DB) ([]string, error) {
	var versions []string

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version DESC")
	if err != nil {
		return versions, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("关闭查询结果集失败: %v", err)
		}
	}()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return versions, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// 获取所有可用的迁移版本
func getAllMigrationVersions() ([]string, error) {
	var versions []string
	versionSet := make(map[string]bool)

	// 获取所有 up 文件的版本
	upFiles, err := getUpMigrationFiles()
	if err != nil {
		return versions, err
	}

	for _, file := range upFiles {
		version := extractVersionFromFilename(file)
		if !versionSet[version] {
			versions = append(versions, version)
			versionSet[version] = true
		}
	}

	// 获取所有 down 文件的版本
	downFiles, err := getDownMigrationFiles()
	if err != nil {
		return versions, err
	}

	for _, file := range downFiles {
		version := extractVersionFromFilename(file)
		if !versionSet[version] {
			versions = append(versions, version)
			versionSet[version] = true
		}
	}

	sort.Strings(versions)
	return versions, nil
}

// 获取 DOWN 迁移文件列表
func getDownMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(MIGRATIONS_DIR, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".down.sql") {
			files = append(files, info.Name())
		}
		return nil
	})

	return files, err
}

// 比较两个版本号的大小
func compareVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	// 如果其中一个为空，空版本视为最小
	if v1 == "" {
		return -1
	}
	if v2 == "" {
		return 1
	}

	// 直接字符串比较（因为我们使用时间戳格式）
	if v1 < v2 {
		return -1
	}
	return 1
}

// 根据版本号查找对应的迁移文件名
func findMigrationFileName(version string) string {
	// 查找 up 文件
	upFiles, err := filepath.Glob(filepath.Join(MIGRATIONS_DIR, version+"_*.up.sql"))
	if err == nil && len(upFiles) > 0 {
		fileName := filepath.Base(upFiles[0])
		// 移除 .up.sql 后缀，只返回描述部分
		name := strings.TrimSuffix(fileName, ".up.sql")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) > 1 {
			return parts[1] // 返回描述部分
		}
	}
	return ""
}

// 导入 SQL 文件
func importSQL(db *sql.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %v", err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("执行 SQL 失败: %v", err)
	}

	fmt.Printf("成功导入 SQL 文件: %s\n", filePath)
	return nil
}

// 导入 CSV 文件（需要指定表名）
func importCSV(db *sql.DB, filePath string) error {
	// 从文件名推断表名
	filename := filepath.Base(filePath)
	tableName := strings.TrimSuffix(filename, filepath.Ext(filename))

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开 CSV 文件失败: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭CSV文件失败: %v", err)
		}
	}()

	reader := csv.NewReader(file)

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取 CSV 表头失败: %v", err)
	}

	// 构建插入语句
	placeholders := strings.Repeat("?,", len(headers))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后的逗号

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(headers, ","), placeholders)

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("关闭预处理语句失败: %v", err)
		}
	}()

	// 读取并插入数据行
	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 CSV 数据失败: %v", err)
		}

		// 转换为 interface{} 切片
		values := make([]interface{}, len(record))
		for i, v := range record {
			values[i] = v
		}

		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("插入数据失败: %v", err)
		}
		count++
	}

	fmt.Printf("成功导入 %d 行数据到表 %s\n", count, tableName)
	return nil
}

// 导出 SQL
func exportSQL(db *sql.DB, outputPath string) error {
	// 获取所有表名
	tables, err := getAllTables(db)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建导出文件失败: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭导出文件失败: %v", err)
		}
	}()

	writer := bufio.NewWriter(file)
	defer func() {
		if err := writer.Flush(); err != nil {
			log.Printf("刷新写入缓冲区失败: %v", err)
		}
	}()

	// 写入文件头
	if _, err := writer.WriteString("-- MySQL dump\n"); err != nil {
		return fmt.Errorf("写入文件头失败: %v", err)
	}
	if _, err := fmt.Fprintf(writer, "-- Generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("写入生成时间失败: %v", err)
	}

	for _, table := range tables {
		if err := exportTableSQL(db, writer, table); err != nil {
			return err
		}
	}

	fmt.Printf("成功导出数据到: %s\n", outputPath)
	return nil
}

// 导出 CSV
func exportCSV(db *sql.DB, outputDir string) error {
	// 获取所有表名
	tables, err := getAllTables(db)
	if err != nil {
		return err
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建导出目录失败: %v", err)
	}

	for _, table := range tables {
		if err := exportTableCSV(db, outputDir, table); err != nil {
			return err
		}
	}

	fmt.Printf("成功导出数据到目录: %s\n", outputDir)
	return nil
}

// 获取所有表名
func getAllTables(db *sql.DB) ([]string, error) {
	var tables []string

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("关闭查询结果集失败: %v", err)
		}
	}()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		// 跳过迁移记录表
		if table != "schema_migrations" {
			tables = append(tables, table)
		}
	}

	return tables, nil
}

// 导出单个表的 SQL
func exportTableSQL(db *sql.DB, writer *bufio.Writer, table string) error {
	// 获取表结构
	createTable, err := getCreateTableStatement(db, table)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "-- Table: %s\n", table); err != nil {
		return fmt.Errorf("写入表注释失败: %v", err)
	}
	if _, err := fmt.Fprintf(writer, "DROP TABLE IF EXISTS `%s`;\n", table); err != nil {
		return fmt.Errorf("写入DROP语句失败: %v", err)
	}
	if _, err := writer.WriteString(createTable + ";\n\n"); err != nil {
		return fmt.Errorf("写入建表语句失败: %v", err)
	}

	// 导出数据
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("关闭查询结果集失败: %v", err)
		}
	}()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 读取数据
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	count := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		if count == 0 {
			if _, err := fmt.Fprintf(writer, "INSERT INTO `%s` (`%s`) VALUES\n",
				table, strings.Join(columns, "`, `")); err != nil {
				return fmt.Errorf("写入INSERT语句失败: %v", err)
			}
		}

		if count > 0 {
			if _, err := writer.WriteString(",\n"); err != nil {
				return fmt.Errorf("写入逗号失败: %v", err)
			}
		}

		if _, err := writer.WriteString("("); err != nil {
			return fmt.Errorf("写入左括号失败: %v", err)
		}
		for i, val := range values {
			if i > 0 {
				if _, err := writer.WriteString(", "); err != nil {
					return fmt.Errorf("写入逗号失败: %v", err)
				}
			}

			if val == nil {
				if _, err := writer.WriteString("NULL"); err != nil {
					return fmt.Errorf("写入NULL失败: %v", err)
				}
			} else {
				switch v := val.(type) {
				case string:
					if _, err := fmt.Fprintf(writer, "'%s'", strings.ReplaceAll(v, "'", "''")); err != nil {
						return fmt.Errorf("写入字符串值失败: %v", err)
					}
				case []byte:
					if _, err := fmt.Fprintf(writer, "'%s'", strings.ReplaceAll(string(v), "'", "''")); err != nil {
						return fmt.Errorf("写入字节值失败: %v", err)
					}
				default:
					if _, err := fmt.Fprintf(writer, "%v", v); err != nil {
						return fmt.Errorf("写入默认值失败: %v", err)
					}
				}
			}
		}
		if _, err := writer.WriteString(")"); err != nil {
			return fmt.Errorf("写入右括号失败: %v", err)
		}
		count++
	}

	if count > 0 {
		if _, err := writer.WriteString(";\n\n"); err != nil {
			return fmt.Errorf("写入结束符失败: %v", err)
		}
	}

	return nil
}

// 导出单个表的 CSV
func exportTableCSV(db *sql.DB, outputDir, table string) error {
	filename := filepath.Join(outputDir, table+".csv")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建 CSV 文件失败: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭CSV文件失败: %v", err)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("关闭查询结果集失败: %v", err)
		}
	}()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 写入表头
	if err := writer.Write(columns); err != nil {
		return err
	}

	// 读取数据
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	count := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		record := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
		count++
	}

	fmt.Printf("导出表 %s: %d 行数据到 %s\n", table, count, filename)
	return nil
}

// 获取建表语句
func getCreateTableStatement(db *sql.DB, table string) (string, error) {
	var createTable string
	err := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table)).Scan(&table, &createTable)
	return createTable, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  go run main.go create <migration_name>  - 创建新的迁移文件")
		fmt.Println("  go run main.go up [steps]               - 执行迁移 (可选步数)")
		fmt.Println("  go run main.go down [steps]             - 回滚迁移 (可选步数)")
		fmt.Println("  go run main.go version                  - 显示当前数据库版本")
		fmt.Println("  go run main.go goto <version>           - 跳转到指定版本")
		fmt.Println("  go run main.go import <file_path>       - 导入 SQL 或 CSV 文件")
		fmt.Println("  go run main.go export sql <output_path> - 导出数据为 SQL 文件")
		fmt.Println("  go run main.go export csv <output_dir>  - 导出数据为 CSV 文件")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("请提供迁移名称")
		}
		if err := createMigration(os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case "up":
		steps := 0
		if len(os.Args) > 2 {
			var err error
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatal("步数必须是数字")
			}
		}
		if err := migrateUpSteps(steps); err != nil {
			log.Fatal(err)
		}

	case "down":
		steps := 1 // 默认回滚 1 步
		if len(os.Args) > 2 {
			var err error
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatal("步数必须是数字")
			}
		}
		if err := migrateDownSteps(steps); err != nil {
			log.Fatal(err)
		}

	case "version":
		if err := showVersion(); err != nil {
			log.Fatal(err)
		}

	case "goto":
		if len(os.Args) < 3 {
			log.Fatal("请提供目标版本号")
		}
		if err := gotoVersion(os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case "import":
		if len(os.Args) < 3 {
			log.Fatal("请提供要导入的文件路径")
		}
		if err := importData(os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case "export":
		if len(os.Args) < 4 {
			log.Fatal("请提供导出格式和输出路径")
		}
		format := os.Args[2]
		outputPath := os.Args[3]
		if err := exportData(format, outputPath); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf("未知命令: %s\n", command)
		os.Exit(1)
	}
}
