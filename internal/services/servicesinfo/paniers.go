package servicesinfo

import "chemin-du-local.bzh/graphql/graph/model"

func Paniers() model.ServiceInfo {
	return model.ServiceInfo{
		ID:                                  "PANIERS",
		Name:                                "Paniers",
		ShortDescription:                    "Le service de Paniers. Mais la description courte est mieux",
		LongDescription:                     "Je ne vais pas me faire trop chier sur la description pour l'instant, mais je vous aime quand même <3",
		MonthPrice:                          30,
		MonthMinimumAllowedCa:               500,
		MonthRangePercentage:                15,
		MonthAugmentationPerRangePercentage: 17,
		MonthAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
			"Avantage 3",
			"Avantage 4",
		},
		TransactionPercentage: 2,
		TransactionAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
			"Avantage 3",
			"Avantage 4",
			"Avantage 5",
		},
	}
}
