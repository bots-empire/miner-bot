package model

type IncomeInfo struct {
	UserID int64  `json:"user_id,omitempty"`
	Source string `json:"source,omitempty"`
}
