package postgres

import (
	"context"
	"fileservice/internal/service/config"
	fileSystem "fileservice/internal/transport/grpc/fileservice"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"os"
)

type Postgres struct {
	log zerolog.Logger
	db  *pgx.Conn
}

func InitDatabase(log zerolog.Logger, cfg config.Config) (*Postgres, error) {
	connConfig := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)

	conn, err := pgx.Connect(context.Background(), connConfig)
	if err != nil {
		//log.Error(`Error connecting to the database: %v`, err)
		return nil, err
	}

	//log.Info("Successfully connected to the database")

	return &Postgres{log: log, db: conn}, nil
}

func (p *Postgres) FileSaver(ctx context.Context, fileName string) (id int, err error) {
	query := "INSERT INTO files.files (filename, created_at, modified_at) VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"

	var insertedID int
	err = p.db.QueryRow(ctx, query, fileName).Scan(&insertedID)
	if err != nil {
		p.log.Error().Msgf("Error executing INSERT query: %v", err)
		return 0, err
	}

	p.log.Info().Msgf("File saved successfully. Inserted ID: %d", insertedID)
	return insertedID, nil
}

func (p *Postgres) GetNameFiles(ctx context.Context) (files []fileSystem.BrowseElements, err error) {
	query := "SELECT id, filename, created_at, modified_at FROM files.files ORDER BY id DESC"
	rows, err := p.db.Query(ctx, query)
	if err != nil {
		p.log.Error().Msgf("Error executing SELECT query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var filesArray []fileSystem.BrowseElements
	for rows.Next() {
		var file fileSystem.BrowseElements
		if err := rows.Scan(&file.Id, &file.Filename, &file.Created_at, &file.Modified_at); err != nil {
			p.log.Error().Msgf("Error scanning row: %v", err)
			return nil, err
		}
		filesArray = append(filesArray, file)
	}

	if err := rows.Err(); err != nil {
		p.log.Error().Msgf("Error iterating over rows: %v", err)
		return nil, err
	}

	return filesArray, nil
}

func (p *Postgres) GetFile(ctx context.Context, fileId int64) (file []byte, err error) {
	query := "SELECT filename FROM files.files WHERE id = $1"
	var filename string
	err = p.db.QueryRow(ctx, query, fileId).Scan(&filename)
	if err != nil {
		p.log.Error().Msgf("Ошибка в выборе файла: %v", err)
		return []byte{}, err
	}
	currentDir, err := os.Getwd()
	if err != nil {
		p.log.Error().Msgf("Ошибка в получении текущей директории")
		return []byte{}, err
	}
	filePath := fmt.Sprintf("%s/SaveFiles/%s", currentDir, filename)
	fileR, err := os.ReadFile(filePath)
	if err != nil {
		p.log.Error().Msgf("Ошибка в чтении файла")
		return []byte{}, err
	}
	return fileR, nil
}

func (p *Postgres) Close(ctx context.Context) error {
	p.log.Info().Msg("Stopping postgres")
	err := p.db.Close(ctx)
	if err != nil {
		p.log.Error().Msgf("Error closing the database connection: %v", err)
	}
	return err
}
