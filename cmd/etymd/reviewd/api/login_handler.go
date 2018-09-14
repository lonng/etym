package api

import (
	"etym/cmd/etymd/protocol"
	"etym/pkg/algoutil"
	"etym/pkg/constant"
	"etym/pkg/db"
	"etym/pkg/errutil"
	"etym/pkg/token"
	"strings"
)

func LoginHandler(req *protocol.ReqLogin) (*protocol.ResLogin, error) {
	account := strings.TrimSpace(req.Username)
	cipher := strings.TrimSpace(req.Password)

	if account == "" || cipher == "" {
		return nil, errutil.ErrIllegalParameter
	}

	password, err := decryptPassword(cipher)
	if err != nil {
		return nil, err
	}
	u, err := db.QueryUserByAccount(account)
	if err != nil {
		return nil, err
	}

	if !algoutil.VerifyPassword(password, u.Salt, u.Password) {
		return nil, errutil.ErrWrongPassword
	}

	t, err := token.New(u.Account)
	if err != nil {
		return nil, err
	}

	return &protocol.ResLogin{Token: t, Review: account == constant.Superadmin}, nil
}
