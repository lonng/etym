package api

import (
	"etym/cmd/etymd/protocol"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
)

func RejectImprove(reject *protocol.ReqReject) (*protocol.ResultResponse, error) {
	if reject.TranslationId < 0 {
		return nil, errutil.ErrInvalidImprovement
	}

	database := db.Database()
	record := &model.DbHistory{
		Id: reject.TranslationId,
	}

	found, err := database.Get(record)
	if err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}

	if !found {
		return nil, errutil.ErrIllegalParameter
	}

	record.Status = model.TranslationStatusReject
	_, err = database.Where("id=?", reject.TranslationId).Cols("status").Update(record)
	if err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}

	if record.PrevHistoryId > 0 {
		prevRecord := &model.DbHistory{
			Id: record.PrevHistoryId,
		}
		found, err = database.Get(prevRecord)
		if err != nil {
			log.Warn(err)
		} else if prevRecord.ProposalCount > 0 {
			prevRecord.ProposalCount--
			_, err = database.Where("id=?", record.PrevHistoryId).
				Cols("proposal_count").Update(prevRecord)
			if err != nil {
				log.Warn(err)
			}
		}
	}

	return protocol.SuccessResponse, nil
}
