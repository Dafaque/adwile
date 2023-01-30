package storage

import (
	"context"
	"database/sql"
	"fmt"
	"healthcheck/internal/saver"
	"log"
	"os"
	"strings"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/init.sql
var queryInit string

//go:embed sql/insert_status.sql
var queryInsertStatus string

//go:embed sql/insert_fail_details.sql
var queryInsertFailDetails string

//go:embed sql/get_last_status.sql
var queryGetLastStatus string

//go:embed sql/get_top_statuses.sql
var queryGetTopStuses string

//go:embed sql/get_status_details.sql
var queryGetStatusDetails string

func NewStorageSQLite3(connstr string, opTimeout int, enableStdout bool) *StorageSQLite3 {
	var connstrOverride string = connstr
	if val, exists := os.LookupEnv("APP.DB_CONN_STR"); exists {
		connstrOverride = val
	}
	db, errOpen := sql.Open("sqlite3", connstrOverride)
	if errOpen != nil {
		log.Printf("NewStorageSQLite3 Open: %s", errOpen)
		return nil
	}
	_, err := db.Exec(queryInit)
	if err != nil {
		log.Printf("NewStorageSQLite3 Exec: %s", err)
		return nil
	}
	return &StorageSQLite3{
		db:           db,
		opTimeout:    opTimeout,
		enableStdout: enableStdout,
	}
}

type StorageSQLite3 struct {
	db           *sql.DB
	opTimeout    int
	enableStdout bool
}

func (s StorageSQLite3) Save(url string, passed bool, failed []string, errMsg string) error {
	if s.enableStdout {
		saver.StdoutSaver{}.Save(url, passed, failed, errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.opTimeout*int(time.Second)))
	defer cancel()
	tx, errMakeTx := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if errMakeTx != nil {
		return errMakeTx
	}

	// @note sqlite3 hasn't bools
	var argPassed int = 0
	if passed {
		argPassed = 1
	}

	var argErrMsg *string = nil
	if len(errMsg) > 0 {
		argErrMsg = &errMsg
	}

	resultStatus, errExecStatus := tx.ExecContext(
		ctx,
		queryInsertStatus,
		time.Now().UnixMicro(),
		url,
		argPassed,
		argErrMsg,
	)
	if errExecStatus != nil {
		_ = tx.Rollback()
		return errExecStatus
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	if len(failed) == 0 {
		return nil
	}

	// MARK: -Insert fail details

	tx, errMakeTx = s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if errMakeTx != nil {
		return errMakeTx
	}

	id, errGetLastID := resultStatus.LastInsertId()
	if errGetLastID != nil {
		_ = tx.Rollback()
		return errGetLastID
	}

	var values []string
	for _, label := range failed {
		values = append(values, fmt.Sprintf("(%d, \"%s\")", id, label))
	}

	_, errExecDetails := tx.ExecContext(
		ctx,
		fmt.Sprintf(queryInsertFailDetails, strings.Join(values, ", ")),
	)
	if errExecDetails != nil {
		_ = tx.Rollback()
		return errExecDetails
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

var (
	statusOk   bool = true
	statusFail bool = false
)

func (s StorageSQLite3) GetLastStatus(url string) (*bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.opTimeout*int(time.Second)))
	defer cancel()
	tx, errMakeTx := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if errMakeTx != nil {
		return nil, errMakeTx
	}

	result, errQuery := tx.QueryContext(ctx, queryGetLastStatus, url)
	if errQuery != nil {
		return nil, errQuery
	}
	defer result.Close()
	if !result.Next() {
		return nil, nil
	}
	var lastStatus int
	if err := result.Scan(&lastStatus); err != nil {
		return nil, err
	}

	if lastStatus == 0 {
		return &statusFail, nil
	}
	return &statusOk, nil
}

type Status struct {
	ID      int      `json:"id"`
	Ts      int      `json:"timestamp"`
	Url     string   `json:"url"`
	Status  bool     `json:"status"`
	Details []string `json:"details,omitempty"`
}

func (s StorageSQLite3) GetTopStatuses(offset int) ([]Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.opTimeout*int(time.Second)))
	defer cancel()
	tx, errMakeTx := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if errMakeTx != nil {
		return nil, errMakeTx
	}

	result, err := tx.QueryContext(ctx, queryGetTopStuses, offset)
	if err != nil {
		return nil, err
	}
	var statuses []Status = make([]Status, 0)
	for result.Next() {
		var status Status
		var passed int
		if err := result.Scan(&status.ID, &status.Ts, &status.Url, &passed); err != nil {
			return nil, err
		}
		if passed > 0 {
			status.Status = true
		} else {
			details, err := s.GetStatusDetails(status.ID)
			if err != nil {
				log.Printf("GetStatusDetails: %s", err)
			} else {
				status.Details = details
			}
		}
		statuses = append(statuses, status)
	}
	result.Close()
	return statuses, nil
}

func (s StorageSQLite3) GetStatusDetails(check_id int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.opTimeout*int(time.Second)))
	defer cancel()
	tx, errMakeTx := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if errMakeTx != nil {
		return nil, errMakeTx
	}

	result, err := tx.QueryContext(ctx, queryGetStatusDetails, check_id)
	if err != nil {
		return nil, err
	}
	var details []string = make([]string, 0)
	for result.Next() {
		var label string
		if err := result.Scan(&label); err != nil {
			return nil, err
		}
		details = append(details, label)
	}
	result.Close()
	return details, nil
}
