package models

import (
	"encoding/gob"
	"kuroko-linux/internal/db"

	"github.com/jmoiron/sqlx"
)

type Peer struct {
	Name string
	Id   uint64
}

func (p *Peer) EnsureExist(tx *sqlx.Tx) {
	db.Tx(tx, func(tx *sqlx.Tx) {
		tx.MustExec(`
			INSERT INTO peers (peerid, name) VALUES (?, ?)
			ON CONFLICT (peerid) DO
				UPDATE SET name=excluded.name`,
			p.Id, p.Name,
		)
	})
}

func init() {
	gob.Register(&Peer{})
}
