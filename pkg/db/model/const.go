package model

const (
	TranslationStatusProposal   = 0
	TranslationStatusApproved   = 1
	TranslationStatusReject     = 2
	TranslationStatusFinal      = 3
	TranslationStatusDeprecated = 4
)

// role
const (
	RoleUnknown    = 0 //未定义
	RoleSuperAdmin = 1 //超级管理员
	RoleAdmin      = 2 //管理员
	RoleOrdinary   = 3 //普通人员
)

// user status
const (
	AgentStatusUnknown   = 0 //未定义
	AgentStatusActivated = 1 //激活
	AgentStatusPending   = 2 //等待审核
	AgentStatusDeleted   = 3 //删除
)
