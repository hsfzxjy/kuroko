package db

import (
	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"path"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	db          *sqlx.DB
	dbInitMutex sync.Mutex
)

const CREATE_TABLES_STMT string = `
DROP TABLE IF EXISTS peers;
CREATE TABLE IF NOT EXISTS peers (
  peerid INTEGER UNIQUE,
  name TEXT
);
CREATE INDEX IF NOT EXISTS peerid_index ON peers (peerid);

DROP TABLE IF EXISTS bluetooth;
DROP TABLE IF EXISTS bridges;
DROP TABLE IF EXISTS peer_history;
CREATE TABLE IF NOT EXISTS bridges (
  bridge_id INTEGER PRIMARY KEY,
  peerid INTEGER,
  type INTEGER,
  address TEXT,
  peer_time INTEGER,
  bridge_time INTEGER,
  UNIQUE(peerid, type, address)
);
CREATE INDEX IF NOT EXISTS bridge_id_index ON bridges (bridge_id);
CREATE INDEX IF NOT EXISTS address_index ON bridges (address);
CREATE INDEX IF NOT EXISTS peerid_index ON bridges (peerid);
CREATE INDEX IF NOT EXISTS type_index ON bridges (type);
CREATE INDEX IF NOT EXISTS peerid_type_index ON bridges (peerid, type);
CREATE INDEX IF NOT EXISTS peer_time_index ON bridges (peer_time);
CREATE INDEX IF NOT EXISTS bridge_time_index ON bridges (bridge_time);
`

func createTables(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return exc.Wrap(err)
	}
	defer tx.Rollback()
	stmts := strings.Split(CREATE_TABLES_STMT, ";")
	for _, stmt := range stmts {
		stmt := strings.TrimSpace(stmt)
		if len(stmt) == 0 {
			continue
		}
		_, err := tx.Exec(stmt)
		if err != nil {
			return exc.Fields("stmt", stmt).Wrap(err)
		}
	}
	return exc.Wrap(tx.Commit())
}

func Db() (ret *sqlx.DB, err error) {
	dbInitMutex.Lock()
	defer dbInitMutex.Unlock()

	if db != nil {
		return db, nil
	}
	ret, err = sqlx.Open("sqlite", path.Join(internal.APP_DIR, "main.db"))
	if err != nil {
		err = exc.Wrap(err)
		return
	}
	if dbIsOutdated() {
		if err = createTables(ret); err != nil {
			err = exc.Wrap(err)
			return
		}
	}
	if err = dumpDbVersion(); err != nil {
		err = exc.Wrap(err)
		return
	}
	db = ret
	return
}

func Tx(tx *sqlx.Tx, action func(tx *sqlx.Tx)) {
	newTx := tx == nil
	if newTx {
		if db, err := Db(); err != nil {
			panic(exc.Wrap(err))
		} else {
			tx = db.MustBegin()
			defer tx.Rollback()
		}
	}
	action(tx)
	if newTx {
		if err := tx.Commit(); err != nil {
			panic(exc.Wrap(err))
		}
	}
}
