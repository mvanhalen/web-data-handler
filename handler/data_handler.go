package handler

import (
	"PostgresDataHandler/entries"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// PostgresDataHandler is a struct that implements the StateSyncerDataHandler interface. It is used by the
// consumer to insert/delete entries into the postgres database.
type PostgresDataHandler struct {
	// A Postgres DB used for the storage of chain data.
	DB *bun.DB
}

// HandleEntryBatch performs a bulk operation for a batch of entries, based on the encoder type.
func (postgresDataHandler *PostgresDataHandler) HandleEntryBatch(batchedEntries []*lib.StateChangeEntry) error {
	if len(batchedEntries) == 0 {
		return fmt.Errorf("PostgresDataHandler.HandleEntryBatch: No entries currently batched.")
	}

	// All entries in a batch should have the same encoder type.
	encoderType := batchedEntries[0].EncoderType

	var err error

	switch encoderType {
	case lib.EncoderTypePostEntry:
		err = entries.PostBatchOperation(batchedEntries, postgresDataHandler.DB)
	}

	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.CallBatchOperationForEncoderType")
	}
	return nil
}

func (postgresDataHandler *PostgresDataHandler) HandleSyncEvent(syncEvent consumer.SyncEvent) error {
	switch syncEvent {
	case consumer.SyncEventStart:
		RunMigrations(postgresDataHandler.DB, true, MigrationTypeInitial)
	case consumer.SyncEventHypersyncStart:
		fmt.Println("Starting hypersync")
	case consumer.SyncEventHypersyncComplete:
		fmt.Println("Hypersync complete")
		// After hypersync, we don't need to maintain so many idle open connections.
		postgresDataHandler.DB.SetMaxIdleConns(4)
		// TODO: Once more encoder types are written out, do a comprehensive comparison between creating indexes
		// immediately and applying indexes after the chain has been hypersynced.
		//postgresDataHandler.DB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
		//RunMigrations(postgresDataHandler.DB, false, MigrationTypePostHypersync)
	case consumer.SyncEventComplete:
		fmt.Printf("\n***** Sync complete *****\n\n")
	}

	return nil
}
