package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
)

func getData() []Zone {
	var zones []Zone

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	zones, age, err := load(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load data: %v\n", err)
		os.Exit(1)
	}

	if time.Now().After(age.Add(time.Hour * 3)) {
		updateCurrency()
		zones = scrapeAllofThem()

		save(zones, conn)
	}

	return zones
}

type DataEntry struct {
	Age   time.Time `json:"age"`
	Zones []Zone    `json:"zones"`
}

func save(z []Zone, conn *pgx.Conn) {

	j, err := json.Marshal(DataEntry{
		time.Now(),
		z,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal data: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(context.Background(), "DELETE FROM power")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to delete data: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Query(context.Background(), "INSERT INTO power (data) VALUES ($1)", j)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to insert data: %v\n", err)
		os.Exit(1)
	}
}

func load(conn *pgx.Conn) ([]Zone, time.Time, error) {
	var data []byte

	err := conn.QueryRow(context.Background(), "SELECT data FROM power LIMIT 1").Scan(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No data, go get new plz: %v\n", err)
		return nil, time.Time{}, nil
	}

	var entry DataEntry
	err = json.Unmarshal(data, &entry)
	if err != nil {
		fmt.Println("Error unmarshalling data: ", err)
		return nil, time.Time{}, err
	}

	return entry.Zones, entry.Age, nil
}
