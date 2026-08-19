package mysql

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func Open(dsn string) (*sql.DB, error) {
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour)
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if e = db.PingContext(ctx); e != nil {
		return nil, e
	}
	return db, nil
}
