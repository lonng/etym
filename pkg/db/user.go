package db

import (
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"time"
)

func RegisterUser(u *model.DbUser) error {
	if u == nil {
		return errutil.ErrIllegalParameter
	}

	if u.Status == model.AgentStatusUnknown {
		u.Status = model.AgentStatusActivated
	}

	if u.Role == model.RoleUnknown {
		u.Role = model.RoleOrdinary
	}

	u.CreateAt = time.Now().Unix()
	_, err := database.Insert(u)
	if err != nil {
		log.Error(err)
	}
	return err
}

func IsAccountExist(account string) bool {
	u := &model.DbUser{Account: account}

	ok, err := database.Where("status<>?", model.AgentStatusDeleted).Get(u)
	if err != nil {
		log.Error(err)
		return false
	}

	return ok
}

func QueryUserByAccount(account string) (*model.DbUser, error) {
	u := &model.DbUser{}

	ok, err := database.Where("status<>? AND account = ?", model.AgentStatusDeleted, account).Get(u)

	if !ok {
		return nil, errutil.ErrUserNotFound
	}

	if err != nil {
		log.Error(err)
		return nil, err
	}
	return u, nil
}
