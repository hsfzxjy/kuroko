package kmux

import "github.com/hsfzxjy/smux"

var SessionNewExtraFunc func(Session) any

var SmuxClientConfig *smux.Config
