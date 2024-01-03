package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

func InitDatabase() (*pgxpool.Pool, error) {

	if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	database_name := os.Getenv("DATABASE_NAME")

	config, err := pgxpool.ParseConfig(" host=" + host + " port=" + port + " user=" + user + " password=" + password + " database=" + database_name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Create tables
	sqlQueries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		
		

		`CREATE TABLE IF NOT EXISTS doctor_info (
			doctor_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) NOT NULL,
			first_name VARCHAR(50) NOT NULL,	
			last_name VARCHAR(50) NOT NULL,
			age INTEGER NOT NULL,
			sex VARCHAR(50) NOT NULL,
			hashed_password VARCHAR(50) NOT NULL,
			salt VARCHAR(50) NOT NULL,
			specialty VARCHAR(50) NOT NULL,
			experience VARCHAR(50) NOT NULL,
			rating_score NUMERIC NOT NULL,
			rating_count INTEGER NOT NULL,
			create_at TIMESTAMP NOT NULL DEFAULT NOW(),
			update_at TIMESTAMP NOT NULL DEFAULT NOW(),
			medical_license VARCHAR(50) NOT NULL,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			doctor_bio VARCHAR(50) NOT NULL,
			email VARCHAR(50) NOT NULL,
			phone_number VARCHAR(50) NOT NULL,
			street_address VARCHAR(50) NOT NULL,
			city_name VARCHAR(50) NOT NULL,
			state_name VARCHAR(50) NOT NULL,
			zip_code VARCHAR(50) NOT NULL,
			country_name VARCHAR(50) NOT NULL,
			birth_date DATE NOT NULL,
			location VARCHAR(50) NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS patient_info (
			patient_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) NOT NULL,
			first_name VARCHAR(50) NOT NULL,	
			last_name VARCHAR(50) NOT NULL,
			age INTEGER NOT NULL,
			sex VARCHAR(50) NOT NULL,
			hashed_password VARCHAR(50) NOT NULL,
			salt VARCHAR(50) NOT NULL,
			create_at TIMESTAMP NOT NULL DEFAULT NOW(),
			update_at TIMESTAMP NOT NULL DEFAULT NOW(),
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			patient_bio VARCHAR(50) NOT NULL,
			email VARCHAR(50) NOT NULL,
			phone_number VARCHAR(50) NOT NULL,
			street_address VARCHAR(50) NOT NULL,
			city_name VARCHAR(50) NOT NULL,
			state_name VARCHAR(50) NOT NULL,
			zip_code VARCHAR(50) NOT NULL,
			country_address VARCHAR(50) NOT NULL,
			birth_date DATE NOT NULL,
			location VARCHAR(50) NOT NULL
		)`,


		`CREATE TABLE IF NOT EXISTS availabilities (
			availability_id SERIAL PRIMARY KEY,
			availability_start TIMESTAMP NOT NULL,
			availability_end TIMESTAMP NOT NULL,
			doctor_id uuid REFERENCES doctor_info(doctor_id)
		)`,


		`CREATE TABLE IF NOT EXISTS appointments (
			appointment_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			appointment_start TIMESTAMP NOT NULL,
			appointment_end TIMESTAMP NOT NULL,
			title VARCHAR(50) NOT NULL,
			doctor_id uuid REFERENCES doctor_info(doctor_id),
			patient_id uuid REFERENCES patient_info(patient_id)
		)`,
	

		`CREATE TABLE IF NOT EXISTS folder_file_info (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			type VARCHAR(50) NOT NULL,
			size INTEGER NOT NULL,
			extension VARCHAR(50) NOT NULL,
			path VARCHAR(50) NOT NULL,
			user_id uuid NOT NULL,
			user_type VARCHAR(50) NOT NULL,
			parent_id uuid REFERENCES folder_file_info(folder_id)
		)`,

		`CREATE TABLE IF NOT EXISTS shared_items (
			id SERIAL PRIMARY KEY,
			shared_by_id VARCHAR(255) NOT NULL, 
			shared_with_id VARCHAR(255) NOT NULL, 
			shared_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			item_id uuid REFERENCES folder_file_info(id)
		)`,

		`CREATE TABLE IF NOT EXISTS chats (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS participants (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			chat_id uuid NOT NULL,
			user_id uuid NOT NULL,
			joined_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS messages (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			chat_id uuid NOT NULL,
			sender_id uuid NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		)`,


	}

	for _, query := range sqlQueries {
		_, err = conn.Exec(context.Background(), query)
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %v", err)
		}
	}

	return conn, nil
}

