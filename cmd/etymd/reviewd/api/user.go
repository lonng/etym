package api

import (
	"etym/cmd/etymd/protocol"
	"etym/pkg/algoutil"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"github.com/lonnng/nex"
	"strings"
)

func AddUser(query *nex.Form) (*protocol.ResultResponse, error) {
	var username = strings.TrimSpace(query.Get("username"))
	var password = strings.TrimSpace(query.Get("password"))

	if username == "" || password == "" {
		return nil, errutil.ErrIllegalParameter
	}

	if db.IsAccountExist(username) {
		return nil, errutil.ErrUserExists
	}

	log.Info("Create user, username=%s", username)

	hash, salt := algoutil.PasswordHash(password)
	u := &model.DbUser{
		Name:          username,
		Account:       username,
		Password:      hash,
		Salt:          salt,
		Role:          model.RoleSuperAdmin,
		Status:        model.AgentStatusActivated,
		CreateAccount: username,
		Extra:         "CREATE AUTOMATICALLY",
	}

	if err := db.RegisterUser(u); err != nil {
		log.Error(err)
		return nil, errutil.ErrDatabaseError
	}

	return protocol.SuccessResponse, nil
}
