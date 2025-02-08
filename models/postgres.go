package models

import (
	"database/sql"
	"fmt"
)

func Open(config PostgresConfig) (*sql.DB ,error){
	db, err := sql.Open("pxg", config.String())
	if err != nil{
		return nil, fmt.Errorf("open: %w", err)
	}
	return db, nil
}

func DefaultPostgresConfig() PostgresConfig{
	return PostgresConfig{
		Host: "localhost",
		Port: "5232",
		User: "baloo",
		Password: "junglebook",
		Database: "lenslocked",
		SSLMode: "disable",
	}
}

type PostgresConfig struct{
	Host string
	Port string
	User string
	Password string
	Database string
	SSLMode string 
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port,cfg.User, cfg.Password,
	cfg.Database, cfg.SSLMode)
}