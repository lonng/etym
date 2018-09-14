package etym

import (
	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/services/etym/memdb"
	"etym/pkg/errutil"
	"github.com/lonnng/nex"
	"strings"
)

func Etym(query *nex.Form) (*protocol.ResEtym, error) {
	word := strings.TrimSpace(query.Get("word"))
	if word == "" {
		return nil, errutil.ErrWordCantEmpty
	}

	return memdb.Etym(word), nil
}
