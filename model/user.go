package model

type User struct {
	ID              int64   `json:"id"`
	Balance         int     `json:"balance"`
	BalanceHash     int     `json:"balance_hash"`
	BalanceBTC      float64 `json:"balance_btc"`
	MiningToday     int     `json:"mining_today"`
	LastClick       int64   `json:"last_click"`
	MinerLevel      int8    `json:"miner_level"`
	FatherID        int64   `json:"father_id"`
	AllReferrals    string  `json:"all_referrals"` // 10/20/30/40
	AdvertChannel   int     `json:"advert_channel"`
	TakeBonus       bool    `json:"take_bonus"`
	Language        string  `json:"language"`
	RegisterTime    int64   `json:"register_time"`
	MinWithdrawal   int     `json:"min_withdrawal"`
	FirstWithdrawal bool    `json:"first_withdrawal"`
	Status          string  `json:"status"`
}
