package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type ConnectionStringer interface {
	ConnectionString() string
}

var connections map[string]*sql.DB = make(map[string]*sql.DB)
var lock sync.RWMutex

func Open(connStringer ConnectionStringer) (*sql.DB, error) {
	lock.Lock()
	defer lock.Unlock()
	existingDb, ok := connections[connStringer.ConnectionString()]
	if ok {
		return existingDb, nil
	}
	db, err := sql.Open("libsql", connStringer.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	connections[connStringer.ConnectionString()] = db
	return db, nil
}

// Repo owns the SQL. Callers only ever see the model types in models.go.
type Repo struct {
	db *sql.DB
}

func OpenRepo(connStringer ConnectionStringer) (*Repo, error) {
	conn, err := Open(connStringer)
	if err != nil {
		return nil, err
	}
	return &Repo{db: conn}, nil
}
