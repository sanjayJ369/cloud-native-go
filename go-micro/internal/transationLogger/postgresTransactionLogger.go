package transactionLogger

import (
	"database/sql"
	"fmt"
	"sync/atomic"

	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events      chan<- Event
	errors      <-chan error
	db          *sql.DB
	lastEventId uint64
}

type PostgresDBParams struct {
	DBName   string
	User     string
	Host     string
	Password string
}

func NewPostgresTransactionLogger(config PostgresDBParams) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		config.Host, config.DBName, config.User, config.Password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %s", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pining db: %s", err)
	}
	logger := &PostgresTransactionLogger{db: db}
	exists, err := logger.verifyTableExists("transactions")
	if err != nil {
		return nil, err
	}

	if !exists {
		err = logger.createTable()
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}

// type TransactionLogger interface {
// 	WritePut(string, string)
// 	WriteDel(string)

// 	Err() <-chan error
// 	Run()
// 	ReadEvents() (<-chan Event, <-chan error) // stream the logged event in file
// 	GetLastEventId() uint64                   // retuns the number of events written to the file
// }

func (p *PostgresTransactionLogger) WritePut(key, value string) {
	p.events <- Event{Key: key, Value: value}
}

func (p *PostgresTransactionLogger) WriteDel(key string) {
	p.events <- Event{Key: key}
}

func (p *PostgresTransactionLogger) Err() <-chan error {
	return p.errors
}

func (p *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	p.events = events

	errors := make(chan error, 1)
	p.errors = errors

	go func() {
		query := `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3)`

		for event := range events {
			res, err := p.db.Exec(query, event.EventType, event.Key, event.Value)
			if err != nil {
				errors <- err
			}
			num, err := res.LastInsertId()
			if err != nil {
				errors <- err
			}
			atomic.StoreUint64(&p.lastEventId, uint64(num))
		}
	}()
}

func (p *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := `select sequence, event_type, key, value 
					from transactions 
					order by sequence`

		rows, err := p.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("error querying db: %s", err)
			return
		}
		defer rows.Close()

		e := Event{}

		for rows.Next() {
			err := rows.Scan(&e.Id, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("error scaning: %s", err)
				return
			}

			outEvent <- e
		}

		if err := rows.Err(); err != nil {
			outError <- fmt.Errorf("error scanning rows: %s", err)
		}
	}()

	return outEvent, outError
}

func (p *PostgresTransactionLogger) GetLastEventId() uint64 {
	return atomic.LoadUint64(&p.lastEventId)
}

func (p *PostgresTransactionLogger) createTable() error {
	query := `CREATE TABLE transactions (
				sequence BIGSERIAL PRIMARY KEY,
				event_type INT NOT NULL,
				key TEXT NOT NULL,
				value TEXT
			);`

	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table: %s", err)
	}

	return nil
}
func (p *PostgresTransactionLogger) verifyTableExists(name string) (bool, error) {
	query := `SELECT tablename FROM pg_catalog.pg_tables
		WHERE schemaname != 'pg_catalog' AND 
    	schemaname != 'information_schema';`

	rows, err := p.db.Query(query)
	if err != nil {
		return false, fmt.Errorf("error executing tables query: %s", err)
	}

	var tbname string
	for rows.Next() {
		err := rows.Scan(tbname)
		if err != nil {
			return false, fmt.Errorf("error scanning rows: %s", err)
		}
		if tbname == name {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("error scanning rows: %s", err)
	}

	return false, nil
}
