package servicesinfo

import "chemin-du-local.bzh/graphql/graph/model"

func ClickAndCollect() model.ServiceInfo {
	return model.ServiceInfo{
		ID:                                  "CLICKANDCOLLECT",
		Name:                                "Click&Collect",
		ShortDescription:                    "Vendez en ligne comme dans votre commerce avec Le Chemin du Local !",
		LongDescription:                     "Vous souhaitez tester le service avant de vous engager sur un mois ? Le service à la consommation est là pour ça !",
		MonthPrice:                          70,
		MonthMinimumAllowedCa:               1500,
		MonthRangePercentage:                30,
		MonthAugmentationPerRangePercentage: 33,
		MonthAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
			"Avantage 3",
			"Avantage 4",
		},
		TransactionPercentage: 5,
		TransactionAdvantages: []string{
			"Avantage 1",
			"Avantage 2",
		},
	}
}
