package review

import (
	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/syncer"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"github.com/lonnng/nex"
)

func Approve(query *nex.Form) (*protocol.ResultResponse, error) {
	translationId := query.Int64OrDefault("id", -1)
	if translationId < 0 {
		return nil, errutil.ErrInvalidImprovement
	}

	record := model.DbHistory{
		Id: translationId,
	}
	found, err := db.Database().Get(&record)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errutil.ErrInvalidImprovement
	}

	session := db.Database().NewSession()
	if err := session.Begin(); err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}
	defer session.Close()

	// 开始事务
	final := query.Int("final") > 0
	if final {
		record.Status = model.TranslationStatusFinal
	} else {
		record.Status = model.TranslationStatusApproved
	}

	// 更新当前记录状态
	_, err = session.Where("id=?", translationId).Cols("status").Update(record)
	if err != nil {
		log.Error(err)
		if err := session.Rollback(); err != nil {
			log.Warn(err)
		}
		return nil, errutil.ErrDatabaseError
	}

	// 非当前记录设为弃用
	_, err = session.Where("`word` = ? AND `id` <> ?", record.Word, translationId).
		Cols("status", "proposal_count").Update(&model.DbHistory{Status: model.TranslationStatusDeprecated, ProposalCount: 0})
	if err != nil {
		log.Error(err)
		if err := session.Rollback(); err != nil {
			log.Warn(err)
		}
		return nil, errutil.ErrDatabaseError
	}

	if err := session.Commit(); err != nil {
		log.Error(err)
		return nil, errutil.ErrDatabaseError
	}
	// 事务结束

	syncer.Sync(syncer.Record{Word: record.Word, Trans: record.NextTrans, Final: final})
	return protocol.SuccessResponse, nil
}
