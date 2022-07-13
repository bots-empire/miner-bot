package model

type Top struct {
	Top       int   `json:"top,omitempty"`
	UserID    int64 `json:"user_id,omitempty"`
	TimeOnTop int   `json:"time_on_top,omitempty"`
	Balance   int   `json:"balance,omitempty"`
}
