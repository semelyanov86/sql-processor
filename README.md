# SQL File Processor

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/sql-processor)](https://goreportcard.com/report/github.com/yourusername/sql-processor)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A robust Go-based service that monitors a directory for SQL files and executes them against a MySQL database. Notifications are sent via email for both successful executions and errors.

## Features

- 🔄 **Real-time Monitoring**: Uses filesystem watcher for immediate detection of new SQL files
- 📧 **Email Notifications**: Configurable SMTP integration for success/error alerts
- 🔒 **Secure Configuration**: All sensitive parameters stored in YAML config file
- ✅ **Error Handling**: Automatic file relocation and detailed error logging
- 🚀 **Lightweight**: Low memory footprint and efficient resource usage

## Prerequisites

- Go 1.21+ (compatible with 1.24)
- MySQL 5.7+
- SMTP server credentials

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/sql-processor.git
cd sql-processor
```

2. Install dependencies:
```bash
go get github.com/fsnotify/fsnotify \
       gopkg.in/yaml.v3 \
       github.com/go-sql-driver/mysql
```

3. Build the binary:
```bash
go build -o sql-processor
```

## Configuration

Create `config.yml` in the project root with following structure:

```yaml
database:
  host: localhost
  port: 3306
  user: db_user
  password: db_password
  name: database_name

smtp:
  server: smtp.yourprovider.com
  port: 587
  username: smtp_user
  password: smtp_password
  from: alerts@yourdomain.com
  to: admin@yourdomain.com

paths:
  input: ./sql/input
  output: ./sql/processed
  error: ./sql/errors
```

## Usage

```bash
./sql-processor
```

Directory structure will be automatically created if not exists:
```
.
├── sql
│   ├── input      # Place new SQL files here
│   ├── processed  # Successfully executed files
│   └── errors     # Files with execution errors
```

## Security Considerations

1. Keep `config.yml` with proper file permissions:
```bash
chmod 600 config.yml
```

2. Use environment variables for sensitive data (optional):
```go
// Modify in code to read from environment:
config.Database.Password = os.Getenv("DB_PASSWORD")
```

3. Regularly rotate SMTP credentials

## Logging

All operations are logged to stdout with timestamps:
```
2024-03-12T14:35:22+03:00 INFO: Started watching directory ./sql/input
2024-03-12T14:35:45+03:00 INFO: Processing file: init_db.sql
2024-03-12T14:35:46+03:00 INFO: Successfully executed init_db.sql
```

## License

MIT License - See [LICENSE](LICENSE) for details
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -am 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Support

For issues and feature requests, please [open an issue](https://github.com/yourusername/sql-processor/issues).
</think>

```markdown
# SQL File Processor

A Go-based service that monitors a directory for SQL files, executes them against a MySQL database, and sends email notifications.

## Features

- Real-time file monitoring
- Email notifications (success/error)
- Configurable via YAML
- Secure credential handling
- Automatic file management

## Prerequisites

- Go 1.21+
- MySQL 5.7+
- SMTP server access

## Installation

1. Clone repo:
```bash
git clone https://github.com/yourusername/sql-processor.git
cd sql-processor
```

2. Install dependencies:
```bash
go get github.com/fsnotify/fsnotify gopkg.in/yaml.v3 github.com/go-sql-driver/mysql
```

3. Build:
```bash
go build -o sql-processor
```

## Configuration

Create `config.yml`:
```yaml
database:
  host: localhost
  port: 3306
  user: db_user
  password: db_pass
  name: db_name

smtp:
  server: smtp.example.com
  port: 587
  username: smtp_user
  password: smtp_pass
  from: alerts@example.com
  to: admin@example.com

paths:
  input: ./sql/input
  output: ./sql/processed
  error: ./sql/errors
```

## Usage

```bash
./sql-processor
```

Directory structure:
```
sql/
├── input      # New SQL files
├── processed  # Successful executions
└── errors     # Failed executions
```

## Security

- Set strict permissions:
```bash
chmod 600 config.yml
```

- Use environment variables for production

## License

MIT - See [LICENSE](LICENSE)