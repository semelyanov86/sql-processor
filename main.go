package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
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
		Output string `yaml:"output"`
		Error  string `yaml:"error"`
	} `yaml:"paths"`
}

var (
	config Config
	db     *sql.DB
)

func main() {
	// Load configuration
	loadConfig()

	// Initialize database connection
	initDB()
	defer db.Close()

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add path to watcher
	err = watcher.Add(config.Paths.Input)
	if err != nil {
		log.Fatal(err)
	}

	// Main loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				processFile(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func loadConfig() {
	configFile, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Error parsing config file:", err)
	}
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping error:", err)
	}
}

func processFile(filePath string) {
	// Read SQL file
	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		handleError(filePath, err)
		return
	}

	// Execute SQL
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		handleError(filePath, err)
		return
	}

	// Move file to output
	newPath := filepath.Join(config.Paths.Output, filepath.Base(filePath))
	err = os.Rename(filePath, newPath)
	if err != nil {
		log.Println("Error moving file:", err)
	}

	// Send success email
	sendEmail("SQL Script Success",
		fmt.Sprintf("Script %s executed successfully", filepath.Base(filePath)))
}

func handleError(filePath string, err error) {
	log.Println("Error:", err)

	// Move file to error directory
	newPath := filepath.Join(config.Paths.Error, filepath.Base(filePath))
	os.Rename(filePath, newPath)

	// Send error email
	sendEmail("SQL Script Error",
		fmt.Sprintf("Error executing script %s: %v", filepath.Base(filePath), err))
}

func sendEmail(subject, body string) {
	auth := smtp.PlainAuth("",
		config.SMTP.Username,
		config.SMTP.Password,
		config.SMTP.Server)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", config.SMTP.To, subject, body))

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", config.SMTP.Server, config.SMTP.Port),
		auth,
		config.SMTP.From,
		[]string{config.SMTP.To},
		msg,
	)

	if err != nil {
		log.Println("Email send error:", err)
	}
}
