package model

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//goland:noinspection ALL
var (
	// income
	TotalIncome = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_income_users",
			Help: "Total count of income users",
		},
		[]string{"bot_link", "bot_name"},
	)
	IncomeBySource = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "type_income_source",
			Help: "Source where the user came from",
		},
		[]string{"bot_link", "bot_name", "source"},
	)

	// updates
	HandleUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "count_of_handle_updates",
			Help: "Total count of handle updates",
		},
		[]string{"bot_link", "bot_name"},
	)

	// clicks
	MoreMoneyButtonClick = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "more_money_button_click",
			Help: "Total click on more money button",
		},
		[]string{"bot_link", "bot_name"},
	)
	CheckSubscribe = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_check_subscribe",
			Help: "Total check subscribe",
		},
		[]string{"bot_link", "bot_name", "advert_link", "source"},
	)

	// mailing
	MailToUser = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_mailing_users",
			Help: "Total check subscribe",
		},
		[]string{"bot_name"},
	)
	BlockUser = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_block_users",
			Help: "Total blocked users",
		},
		[]string{"bot_name"},
	)
)
