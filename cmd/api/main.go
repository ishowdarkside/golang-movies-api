package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/data"
	"github.com/ishowdarkside/go-movies-app/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer *mailer.Mailer
}

func main() {

	godotenv.Load()
	var cfg config

	os.Setenv("MOVIE_API_DB_DSN", os.Getenv("DB_LINK"))

	// Parsing flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Evironment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("MOVIE_API_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// SMTP configuration
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 25
	}
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", port, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sener", os.Getenv("SMTP_SENDER"), "SMTP sender")

	flag.Parse()

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(&cfg)
	if err != nil {

		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	mailer, err := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	if err != nil {

		logger.Error(err.Error())
		os.Exit(1)
		return
	}

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {

		logger.Error(err.Error())
		os.Exit(1)
	}

}

func openDB(cfg *config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {

		db.Close()
		return nil, err

	}

	return db, nil

}
