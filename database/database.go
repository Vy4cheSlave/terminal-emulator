package database

import (
	// std
	"context"
	"fmt"
	"log"
	"time"
	"path/filepath"
	// local
	"github.com/Vy4cheSlave/test-task-postgres/models"
	// web
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:generate mockgen -source=database.go -destination=mock/mock.go

type DBWorker interface {
	Ping(context.Context) error
	Close()
	CreateNewCommandsQuery([]models.CommandsWithoutID, context.Context) error
	GettingListCommandsQuery(context.Context) (*[]models.Commands, error)
	GettingSingleCommandQuery(uint, context.Context) (*models.Commands, error)
}

type DB struct {
	pool *pgxpool.Pool
}

func ConnectToDB(databaseUrl string, numerAttemptToConnect uint) (DBWorker, error) {
	var err error
	var pool *pgxpool.Pool
	for range numerAttemptToConnect {
		pool, err = pgxpool.New(context.Background(), databaseUrl)
		if err != nil {
			log.Printf("Unable to create connection pool: %v\n", err)
		}
		if err = pool.Ping(context.Background()); err != nil {
			log.Printf("Unable to ping database: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// migrations
		sqlDB := stdlib.OpenDBFromPool(pool)
		if err := goose.Up(sqlDB, filepath.Join("database/", "migrations/")); err != nil {
			return nil, fmt.Errorf("Couldn't up migrations")
		}

		return DB{pool: pool}, nil
	}
	return nil, fmt.Errorf("Unable to create connection pool %v", err)
}

func (db DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db DB) Close() {
	db.pool.Close()
}

func (db DB) CreateNewCommandsQuery(commands []models.CommandsWithoutID, ctx context.Context) error {
	query := "insert into commands (command, is_error, log) values ($1, $2, $3);"

	// if _, err := db.pool.Exec(ctx, query, command.Command, command.Log); err != nil {
	// 	return fmt.Errorf("unable to insert row: %w", err)
	// }

	batch := &pgx.Batch{}
	for _, command := range commands {
		batch.Queue(query, command.Command, command.IsError, command.Log)
	}

	results := db.pool.SendBatch(ctx, batch)
  	return results.Close()

	// rows, err := db.pool.SendBatch(ctx, batch).Query()
	// if err != nil {
	// 	return fmt.Errorf("unable to query: %w", err)
	// }
	// defer rows.Close()

	// indexCommands := 0
	// for rows.Next() {
	// 	err := rows.Scan(
	// 		&(*commands)[indexCommands].Id, 
	// 		&(*commands)[indexCommands].Command, 
	// 		&(*commands)[indexCommands].IsError,
	// 		&(*commands)[indexCommands].Log,
	// 	)
	// 	if err != nil {
	// 		return fmt.Errorf("unable to values row: %w", err)
	// 	}
	// 	indexCommands++
	// }

	// return nil
}

func (db DB) GettingListCommandsQuery(ctx context.Context) (*[]models.Commands, error) {
	query := "select id, command, is_error, log from commands;"
	
	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}
	defer rows.Close()

	commands := []models.Commands{}
	for rows.Next() {
		command := models.Commands{}
		err := rows.Scan(&command.Id, &command.Command, &command.IsError, &command.Log)
		if err != nil {
		return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		commands = append(commands, command)
	}
	
	return &commands, err
}

func (db DB) GettingSingleCommandQuery(requestId uint, ctx context.Context) (*models.Commands, error) {
	query := "select command, log from commands where id = $1;"

	command := models.Commands{Id: requestId}
	err := db.pool.QueryRow(ctx, query, command.Id).Scan(&command.Command, &command.Log)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}

	return &command, nil
}