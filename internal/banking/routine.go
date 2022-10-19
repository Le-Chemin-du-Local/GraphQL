package banking

import (
	"fmt"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/adlio/trello"
)

func SendBalance() {
	trelloClient := trello.NewClient(config.Cfg.Trello.Key, config.Cfg.Trello.Token)

	dbCommerces, err := commerces.GetAll()

	now := time.Now()
	if err != nil {
		card := trello.Card{
			Name:     "ERREUR DE L'API",
			Desc:     "L'API n'a pas réussi à noter les transferts du jours. Vous devez soit relancer le script, soit le faire manuellement.",
			Due:      &now,
			IDList:   "62fe8cab722001008b9a8940",
			IDLabels: []string{"62fe8cab1818e60499aa7f5f"},
		}

		err := trelloClient.CreateCard(&card, trello.Defaults())

		if err != nil {
			fmt.Print("[DailyBankTransferts] ERREUR : Erreur dans l'erreur ; ")
			fmt.Println(err)
		}
	}

	for _, commerce := range dbCommerces {
		if commerce.Balance == 0 {
			continue
		}

		labelID := config.Cfg.Trello.NonBankRegisteredLabelID

		if len(commerce.Transferts) > 0 {
			labelID = config.Cfg.Trello.AlreadyBankRegisteredLabelID
		}

		card := trello.Card{
			Name:     fmt.Sprintf("%.2f€ - (%s) %s", commerce.Balance, commerce.Name, *commerce.IBANOwner),
			Desc:     fmt.Sprintf("IBAN : %s\nBIC : %s", *commerce.IBAN, *commerce.BIC),
			Due:      &now,
			IDList:   config.Cfg.Trello.ListID,
			IDLabels: []string{labelID},
		}

		if commerce.Transferts == nil {
			commerce.Transferts = []model.Transfert{}
		}

		commerce.Transferts = append(commerce.Transferts, model.Transfert{
			Value:     commerce.Balance,
			IbanOwner: *commerce.IBANOwner,
			Iban:      *commerce.IBAN,
			Bic:       *commerce.BIC,
		})
		commerce.Balance = 0

		err = commerces.Update(&commerce, nil, nil)

		if err != nil {
			errorCard := trello.Card{
				Name:   fmt.Sprintf("ERREUR COMMERCE %s", commerce.Name),
				Desc:   fmt.Sprintf("Le commerce %s n'a pas pu être mise à jour, et la carte n'a pas été créé en conséquent. Montant dû : %.2f", commerce.Name, commerce.Balance),
				Due:    &now,
				IDList: config.Cfg.Trello.ListID,
			}

			err = trelloClient.CreateCard(&errorCard, trello.Defaults())

			if err != nil {
				fmt.Print("[DailyBankTransferts] ERREUR : Impossible d'envoyer la carte d'erreur pour le commerce " + commerce.Name + " ; ")
				fmt.Println(err)
			}

			continue
		}

		err = trelloClient.CreateCard(&card, trello.Defaults())

		if err != nil {
			fmt.Print("[DailyBankTransferts] ERREUR : Impossible d'envoyer la carte pour le commerce " + commerce.Name + " ; ")
			fmt.Println(err)
		}
	}

	fmt.Println("Le scripte des balance des commerces a été executé")
}
