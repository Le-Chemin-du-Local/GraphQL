package servicesinfo

import "chemin-du-local.bzh/graphql/graph/model"

func ClickAndCollect() model.ServiceInfo {
	return model.ServiceInfo{
		Name:             "Click&Collect",
		ShortDescription: "Le service de Click&Collect",
		LongDescription:  "Je ne vais pas me faire trop chier sur la description pour l'instant, mais je vous aime quand même <3",
		MonthPrice:       60,
		MonthAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
		},
		MonthConditions: []string{
			"Condition 1",
			"Condition 2",
			"Condition 3",
			"Condition 4",
		},
		TransactionPercent: 5,
		TransactionAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
		},
		TransactionConditions: []string{
			"Condition 1",
			"Condition 2",
			"Condition 3",
			"Condition 4",
		},
	}
}
