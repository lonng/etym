package review

import (
	"etym/cmd/etymd/protocol"
	"etym/cmd/etymd/services/etym/memdb"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"github.com/lonnng/nex"
	"html"
	"net/http"
	"strings"
)

func ImproveInfo(query *nex.Form) (*protocol.ResHistoryInfo, error) {
	word := html.UnescapeString(query.Get("word"))
	word = strings.TrimSpace(word)
	if word == "" {
		return nil, errutil.ErrEmptyQueryWord
	}

	record := model.DbHistory{
		Word: word,
	}

	database := db.Database()
	found, err := database.Where("status=?", model.TranslationStatusApproved).Desc("id").Get(&record)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errutil.ErrWordNotFound
	}

	res := &protocol.ResHistoryInfo{
		Word:          record.Word,
		TranslationId: record.Id,
		Original:      record.Original,
		NextTrans:     record.NextTrans,
		Translation:   memdb.FindDictItem(model.CleanWord(word)),
	}

	return res, nil
}

func SubmitImprove(request *http.Request, improvement *protocol.ReqSubmitImprove) (*protocol.ResultResponse, error) {
	improvement.Word = strings.TrimSpace(improvement.Word)
	if improvement.Word < "" {
		return nil, errutil.ErrInvalidImprovement
	}

	record := model.DbHistory{
		Word: improvement.Word,
	}
	found, err := db.Database().Where("status=?", model.TranslationStatusApproved).Desc("id").Get(&record)
	if err != nil {
		return nil, err
	}

	if !found {
		log.Warnf("收到无效改进请求, Word=%d, RemoteAddr=%s",
			improvement.Word, request.RemoteAddr)
		return nil, errutil.ErrInvalidImprovement
	}

	if record.Status == model.TranslationStatusFinal {
		return protocol.SuccessResponse, nil
	}

	if improvement.Word != record.Word {
		log.Warnf("收到无效改进请求, Word=%s, ReceiveWord=%s, RemoteAddr=%s",
			record.Word, improvement.Word, request.RemoteAddr)
		return nil, errutil.ErrInvalidImprovement
	}

	newRecored := model.DbHistory{
		Word:          record.Word,
		LowerWord:     record.LowerWord,
		Original:      record.Original,
		Score:         record.Score + 1,
		PrevHistoryId: record.Id,
		PrevTrans:     record.NextTrans,
		NextTrans:     improvement.Translation,
	}

	if _, err := db.Database().Insert(newRecored); err != nil {
		log.Errorf("数据库操作错误: %v", err)
		return nil, errutil.ErrDatabaseError
	}

	record.ProposalCount++
	if _, err := db.Database().
		Where("id=?", record.Id).
		Cols("proposal_count").
		Update(record); err != nil {
		log.Warnf("更新Proposal失败, TranslationId=%d, Error=%v", record.Id, err)
	}

	log.Infof("提交修改成功, Word:%s, RemoteAddr=%s", improvement.Word, request.RemoteAddr)
	return protocol.SuccessResponse, nil
}
