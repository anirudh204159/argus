package argus

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
)

const (
	hardcodedSourceID  = 3
	checkpointInterval = 5 * time.Second
)

var (
	currentPositionMu sync.RWMutex
	currentPosition   mysql.Position

	engineCancel context.CancelFunc
)

func updateCurrentPosition(pos mysql.Position) {
	currentPositionMu.Lock()
	currentPosition = pos
	currentPositionMu.Unlock()
}

func getCurrentTrackedPosition() mysql.Position {
	currentPositionMu.RLock()
	defer currentPositionMu.RUnlock()
	return currentPosition
}

// RunEngine starts the binlog reader and event pipeline.
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

	if err := initMetadataDB(); err != nil {
		return fmt.Errorf("metadata db init: %w", err)
	}
	defer metadataDB.Close()
	fmt.Println("Engine connected to metadata DB")

	if err := initRedis(); err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer rdb.Close()
	fmt.Println("Connected to Redis")

	ctx, cancel := context.WithCancel(context.Background())
	engineCancel = cancel
	defer cancel()

	pos, err := determineStartPosition(ctx)
	if err != nil {
		return fmt.Errorf("determine start position: %w", err)
	}
	fmt.Printf("Starting from %s:%d\n", pos.Name, pos.Pos)
	updateCurrentPosition(pos)

	// Start background checkpoint saver
	go checkpointLoop(ctx)

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
		ev, err := streamer.GetEvent(ctx)
		if err != nil {
			// Context cancelled = graceful shutdown, not an error
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("get event: %w", err)
		}

		updateCurrentPosition(mysql.Position{
			Name: getCurrentTrackedPosition().Name,
			Pos:  ev.Header.LogPos,
		})

		if rotateEvent, ok := ev.Event.(*replication.RotateEvent); ok {
			updateCurrentPosition(mysql.Position{
				Name: string(rotateEvent.NextLogName),
				Pos:  uint32(rotateEvent.Position),
			})
			continue
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

// determineStartPosition decides where to begin reading the binlog.
func determineStartPosition(ctx context.Context) (mysql.Position, error) {
	file, posVal, err := loadCheckpoint(ctx, hardcodedSourceID)
	if err != nil {
		return mysql.Position{}, err
	}

	if file != "" {
		fmt.Printf("Resuming from checkpoint: %s:%d\n", file, posVal)
		return mysql.Position{Name: file, Pos: posVal}, nil
	}

	fmt.Println("No checkpoint found, starting from current position")
	return getCurrentPosition()
}

// checkpointLoop periodically saves the current binlog position.
// Runs as a goroutine for the lifetime of the engine.
func checkpointLoop(ctx context.Context) {
	ticker := time.NewTicker(checkpointInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final checkpoint write on shutdown
			pos := getCurrentTrackedPosition()
			if pos.Name != "" {
				if err := saveCheckpoint(context.Background(), hardcodedSourceID, pos.Name, pos.Pos); err != nil {
					fmt.Printf("final checkpoint error: %v\n", err)
				} else {
					fmt.Printf("Final checkpoint saved: %s:%d\n", pos.Name, pos.Pos)
				}
			}
			return

		case <-ticker.C:
			pos := getCurrentTrackedPosition()
			if pos.Name == "" {
				continue
			}
			if err := saveCheckpoint(ctx, hardcodedSourceID, pos.Name, pos.Pos); err != nil {
				fmt.Printf("checkpoint error: %v\n", err)
			}
		}
	}
}

// ShutdownEngine triggers graceful shutdown of the engine.
// Causes the main loop to exit and the checkpoint loop to save a final position.
func ShutdownEngine() {
	if engineCancel != nil {
		engineCancel()
	}
}
