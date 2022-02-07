package model

type User struct {
	ID             int64  `json:"id"`
	Balance        int    `json:"balance"`
	Completed      int    `json:"completed"`
	CompletedToday int    `json:"completed_today"`
	LastVoice      int64  `json:"last_voice"`
	ReferralCount  int    `json:"referral_count"`
	TakeBonus      bool   `json:"take_bonus"`
	Language       string `json:"language"`
}
