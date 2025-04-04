package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	_ "github.com/octoper/go-ray"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`

	SMTP struct {
		Server   string `yaml:"server"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
		To       string `yaml:"to"`
	} `yaml:"smtp"`

	Paths struct {
		Input  string `yaml:"input"`
		Done   string `yaml:"done"`
		Failed string `yaml:"failed"`
	} `yaml:"paths"`
}

var (
	config Config
	db     *sql.DB
)

func main() {
	loadConfig("config.yml")

	initDB()
	defer db.Close()

	for {
		processFiles()
		time.Sleep(5 * time.Second) // Проверка каждые 5 секунд
	}
}

func loadConfig(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
}

func processFiles() {
	files, err := os.ReadDir(config.Paths.Input)
	if err != nil {
		log.Printf("Error reading input directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(config.Paths.Input, file.Name())
		processFile(filePath)
	}
}

func processFile(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return
	}

	if !strings.Contains(string(data), "-- &SQL DONE& --") {
		log.Printf("File %s is incomplete (missing '-- &SQL DONE& --'). Skipping execution.", filePath)
		return
	}

	err = executeSQL(string(data))
	if err != nil {
		handleError(filePath, err)
	} else {
		handleSuccess(filePath)
	}
}

func executeSQL(query string) error {
	if strings.Contains(query, "\ufeff") {
		query = strings.ReplaceAll(query, "\ufeff", "")
	}

	_, err := db.Exec(query)
	return err
}

func handleError(filePath string, err error) {
	log.Printf("Error executing SQL: %v", err)
	sendEmail("SQL Execution Error", fmt.Sprintf("Error: %v\nFile: %s", err, filePath))
	moveFile(filePath, config.Paths.Failed)
}

func handleSuccess(filePath string) {
	log.Println("SQL executed successfully")
	sendEmail("SQL Execution Success", fmt.Sprintf("File: %s", filePath))
	moveFile(filePath, config.Paths.Done)
}

func moveFile(source, destDir string) {
	fileName := filepath.Base(source)
	destPath := filepath.Join(destDir, fileName)

	err := os.Rename(source, destPath)
	if err != nil {
		log.Printf("Error moving file: %v", err)
	}
}

func sendEmail(subject, body string) {
	tlsConfig := &tls.Config{
		ServerName:         config.SMTP.Server,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.SMTP.Server, config.SMTP.Port), tlsConfig)
	if err != nil {
		log.Printf("Error creating TLS connection: %v", err)
		return
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.SMTP.Server)
	if err != nil {
		log.Printf("Error creating SMTP client: %v", err)
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", config.SMTP.Username, config.SMTP.Password, config.SMTP.Server)
	if err := client.Auth(auth); err != nil {
		log.Printf("SMTP auth error: %v", err)
		return
	}

	if err := client.Mail(config.SMTP.From); err != nil {
		log.Printf("Mail command error: %v", err)
		return
	}
	if err := client.Rcpt(config.SMTP.To); err != nil {
		log.Printf("Rcpt command error: %v", err)
		return
	}

	wc, err := client.Data()
	if err != nil {
		log.Printf("Data command error: %v", err)
		return
	}
	defer wc.Close()

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.SMTP.From,
		config.SMTP.To,
		subject,
		body,
	)

	if _, err = fmt.Fprint(wc, msg); err != nil {
		log.Printf("Error writing message: %v", err)
		return
	}
}
