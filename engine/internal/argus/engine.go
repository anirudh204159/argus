package argus

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
)

// RunEngine starts the binlog reader and event pipeline.
// Blocks forever — returns only on fatal error.
func RunEngine() error {
	cfg := replication.BinlogSyncerConfig{
		ServerID: 100,
		Flavor:   "mysql",
		Host:     "127.0.0.1",
		Port:     3307,
		User:     "root",
		Password: "rootpass",
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	var err error
	schemaDB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("schema connection: %w", err)
	}
	defer schemaDB.Close()

	if err := initRedis(); err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer rdb.Close()
	fmt.Println("Connected to Redis")

	pos, err := getCurrentPosition()
	if err != nil {
		return fmt.Errorf("get current position: %w", err)
	}
	fmt.Printf("Starting from %s:%d\n", pos.Name, pos.Pos)

	syncer := replication.NewBinlogSyncer(cfg)
	defer syncer.Close()

	streamer, err := syncer.StartSync(pos)
	if err != nil {
		return fmt.Errorf("start sync: %w", err)
	}

	eventChan := make(chan Event, 1000)
	go consumer(eventChan)

	fmt.Println("Listening for new events...")

	for {
		ev, err := streamer.GetEvent(context.Background())
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		switch e := ev.Event.(type) {
		case *replication.RowsEvent:
			events, err := parseRowsEvent(ev.Header.EventType, e)
			if err != nil {
				fmt.Println("parse error:", err)
				continue
			}
			for _, event := range events {
				eventChan <- event
			}
		}
	}
}
