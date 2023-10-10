package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDatabase() (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig("host=localhost port=5433 user=postgres password=Amine-1963 database=tbibi_app")
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

	}

	for _, query := range sqlQueries {
		_, err = conn.Exec(context.Background(), query)
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %v", err)
		}
	}

	return conn, nil
}

