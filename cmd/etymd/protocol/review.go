package protocol

type (
	ReqLogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	UserInfo struct {
		Id            int64  `json:"id"`
		Role          int    `json:"role"`
		Name          string `json:"name"`
		Account       string `json:"account"`
		Status        string `json:"status"`
		CreateAt      int64  `json:"create_at"`
		CreateAccount string `json:"create_account"`
		Extra         string `json:"extra"`
	}

	ResLogin struct {
		Code   int    `json:"code"`
		Token  string `json:"token"`
		Review bool   `json:"review"`
	}
)
