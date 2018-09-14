package api

import (
	"etym/cmd/etymd/protocol"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
)

func RandReview() (*protocol.ResRandReview, error) {
	record := &model.DbHistory{}
	database := db.Database()
	total, err := database.Where("status=?", model.TranslationStatusProposal).Count(record)
	if err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}

	if total < 1 {
		return &protocol.ResRandReview{Total: -1}, nil
	}

	found, err := database.Where("status=?", model.TranslationStatusProposal).OrderBy("RAND()").Limit(1).Get(record)
	if err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}

	if !found {
		// EMPTY implied none pending review record
		//return nil, errutil.ErrServerInternal
		return &protocol.ResRandReview{Total: -1}, nil
	}

	res := &protocol.ResRandReview{
		Total: total,
		HistoryInfo: &protocol.ResHistoryInfo{
			Word:          record.Word,
			TranslationId: record.Id,
			PrevTrans:     record.PrevTrans,
			NextTrans:     record.NextTrans,
			Original:      record.Original,
		},
	}
	return res, nil
}
