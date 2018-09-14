package etym

import (
	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/services/etym/memdb"
	"etym/pkg/errutil"
	"github.com/lonnng/nex"
	"strings"
)

func Search(query *nex.Form) ([]*protocol.ResDictWord, error) {
	word := strings.TrimSpace(query.Get("word"))
	if len(word) < 1 || word == "" {
		return nil, errutil.ErrWordCantEmpty
	}

	word = strings.ToLower(word)
	return memdb.Search(word), nil
}
