package protocol

type (
	Reference struct {
		Word    string         `json:"word"`
		Dict    *ResDictWord   `json:"dict"`
		Related []*ResDictWord `json:"related"`
	}

	ResEtymInfo struct {
		Name    string   `json:"name"`
		Etym    string   `json:"etym"`
		EtymCN  string   `json:"etymCN,omitempty"`
		Ref     []string `json:"ref"`
		Foreign []string `json:"foreign"`
		Final   bool     `json:"final"`
	}

	ResEtym struct {
		Word    string         `json:"word"`
		Etym    []*ResEtymInfo `json:"etym"`
		Trans   *ResDictWord   `json:"trans"`
		Related []*ResDictWord `json:"related"`
		Ref     []*Reference   `json:"ref"`
	}

	ResDictWord struct {
		Word        string `json:"word"`
		Phonetic    string `json:"phonetic"`
		Translation string `json:"translation"`
	}

	ResultResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	ReqSubmitImprove struct {
		Word        string `json:"word"`
		Translation string `json:"translation"`
	}

	WordListItem struct {
		Value int    `json:"value"`
		Desc  string `json:"desc"`
	}

	ResWordList struct {
		Items []*WordListItem `json:"items"`
	}

	ResHistoryInfo struct {
		Word          string       `json:"word"`
		TranslationId int64        `json:"translation_id"`
		PrevTrans     string       `json:"prev,omitempty"`
		NextTrans     string       `json:"next"`
		Original      string       `json:"original"`
		Translation   *ResDictWord `json:"translation,omitempty"`
	}

	ResRandReview struct {
		Total       int64           `json:"total"`
		HistoryInfo *ResHistoryInfo `json:"history_info"`
	}

	ReqReject struct {
		TranslationId int64 `json:"translation_id"`
	}
)

var SuccessResponse = &ResultResponse{Msg: "success"}
