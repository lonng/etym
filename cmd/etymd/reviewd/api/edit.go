package api

import (
	"context"
	"etym/cmd/etymd/protocol"
	"etym/pkg/db"
	"etym/pkg/db/model"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"net/http"
	"strings"
)

func RandEdit() (*protocol.ResHistoryInfo, error) {
	record := &model.DbHistory{}
	database := db.Database()
	found, err := database.
		Where("status=? AND proposal_count=0", model.TranslationStatusApproved).
		OrderBy("CHAR_LENGTH(original)").
		Limit(1).
		Get(record)
	if err != nil {
		log.Warn(err)
		return nil, errutil.ErrDatabaseError
	}

	if !found {
		return nil, errutil.ErrServerInternal
	}
	res := &protocol.ResHistoryInfo{
		Word:          record.Word,
		TranslationId: record.Id,
		Original:      record.Original,
		NextTrans:     record.NextTrans,
	}
	return res, nil
}

func SubmitEdit(ctx context.Context, request *http.Request, improvement *protocol.ReqSubmitImprove) (*protocol.ResultResponse, error) {
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

	if improvement.Word != record.Word {
		log.Warnf("收到无效改进请求, Word=%s, ReceiveWord=%s, RemoteAddr=%s",
			record.Word, improvement.Word, request.RemoteAddr)
		return nil, errutil.ErrInvalidImprovement
	}

	if improvement.Translation == record.NextTrans {
		return protocol.SuccessResponse, nil
	}

	newRecored := &model.DbHistory{
		Word:          record.Word,
		LowerWord:     record.LowerWord,
		Score:         record.Score + 1,
		Original:      record.Original,
		PrevHistoryId: record.Id,
		PrevTrans:     record.NextTrans,
		NextTrans:     improvement.Translation,
		CreateAccount: ctx.Value("account").(string),
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

	/*resp, err := http.Get(fmt.Sprintf("%s/api/v1/review/approve?id=%d&final=%d",
		RemoteUrl, newRecored.Id, improvement.Final))
	if err != nil {
		log.Warnf("提交到Etymd失败, err=%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("提交到Etymd失败, Status=%v", resp.Status)
	}*/

	return protocol.SuccessResponse, nil
}
