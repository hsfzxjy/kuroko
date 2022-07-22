package models

import (
	"encoding/gob"
	"fmt"
	"kmux"
	"kuroko-linux/internal"
	"kuroko-linux/internal/db"
	"kuroko-linux/internal/exc"

	"github.com/jmoiron/sqlx"
)

type Bridge struct {
	Id int64

	Type kmux.TransportType
	Addr string
	Name string

	Peer       *Peer
	PeerTime   *int64
	BridgeTime *int64
}

func init() {
	gob.Register(new(Bridge))
}

const QUERY_TMPL string = `
SELECT
	peers.peerid as peerid,
	peers.name as name,
	bridges.bridge_id as bridge_id,
	bridges.peer_time as peer_time,
	bridges.bridge_time as bridge_time,
	bridges.type as type,
	bridges.address as address
FROM
	peers
LEFT OUTER JOIN
	bridges
ON
	bridges.peerid = peers.peerid %s
`

func query[R any](querier func(string, ...any) R, cond string, args ...any) R {
	stmt := fmt.Sprintf(QUERY_TMPL, cond)
	return querier(stmt, args...)
}

func queryError[R any](querier func(string, ...any) (R, error), cond string, args ...any) (R, error) {
	stmt := fmt.Sprintf(QUERY_TMPL, cond)
	return querier(stmt, args...)
}

func NewBridge(bi internal.BridgeInfo) *Bridge {
	return &Bridge{
		Type: bi.GetType(),
		Addr: bi.GetAddr(),
	}
}

func GetBridgeList() (ret []*Bridge, err error) {
	db, err := db.Db()
	if err != nil {
		return
	}
	rows, err := queryError(db.Query, "ORDER BY peer_time DESC, bridge_time DESC")
	if err != nil {
		return
	}
	for rows.Next() {
		bi := new(Bridge)
		peer := new(Peer)
		err = rows.Scan(&peer.Id, &peer.Name, &bi.Id, &bi.BridgeTime, &bi.PeerTime, &bi.Type, &bi.Addr)
		if err != nil {
			ret = nil
			return
		}
		bi.Peer = peer
		ret = append(ret, bi)
	}
	return
}

func (b *Bridge) GetAddr() string             { return b.Addr }
func (b *Bridge) GetType() kmux.TransportType { return b.Type }

func (bi *Bridge) IsAnonymous() bool { return bi.Peer == nil }

func (bi *Bridge) UpdateSelf(tx *sqlx.Tx, peerId *uint64) {
	var cond string
	var args []any
	if peerId == nil {
		cond = "WHERE type = ? AND address = ?"
		args = []any{bi.Type, bi.Addr}
	} else {
		cond = "WHERE type = ? AND address = ? AND bridges.peerid = ?"
		args = []any{bi.Type, bi.Addr, *peerId}
	}
	db.Tx(tx, func(tx *sqlx.Tx) {
		row := query(tx.QueryRow, cond, args...)
		bi.Peer = &Peer{}
		err := row.Scan(
			&bi.Peer.Id, &bi.Peer.Name,
			&bi.Id, &bi.PeerTime, &bi.BridgeTime, &bi.Type, &bi.Addr,
		)
		if err != nil {
			panic(exc.Wrap(err))
		}
	})
}

func (bi *Bridge) BindPeer(tx *sqlx.Tx, peer *Peer) {
	if !bi.IsAnonymous() {
		panic(exc.New("Cannot bind peer to non-anonymous bridge"))
	}
	db.Tx(tx, func(tx *sqlx.Tx) {
		peer.EnsureExist(tx)
		tx.MustExec(`
			INSERT INTO peers (peerid, name)
			VALUES (?, ?)
			ON CONFLICT (peerid) DO
			UPDATE SET name=excluded.name`,
			peer.Id, peer.Name,
		)
		res := tx.MustExec(`
			INSERT OR IGNORE INTO bridges (peerid, type, address)
			VALUES (?, ?, ?)`,
			peer.Id, bi.Type, bi.Addr,
		)
		if nrows, _ := res.RowsAffected(); nrows > 0 {
			bi.Id, _ = res.LastInsertId()
		}
	})
}
