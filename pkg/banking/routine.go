package banking

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/pkg/htmltopdf"
	"chemin-du-local.bzh/graphql/pkg/notifications"
	"chemin-du-local.bzh/graphql/pkg/stripehandler"
	"github.com/adlio/trello"
)

func sendBalance() {
	// On doit créer les services
	commercesService := commerces.NewCommercesService()

	trelloClient := trello.NewClient(config.Cfg.Trello.Key, config.Cfg.Trello.Token)

	dbCommerces, err := commercesService.GetAll()

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

		err = commercesService.Update(&commerce, nil, nil)

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

func getServicesPayment() {
	// On doit créer les services
	commercesService := commerces.NewCommercesService()

	dbCommerces, err := commercesService.GetAll()

	if err != nil {
		fmt.Println(err)
		fmt.Println("Impossible d'accéder à la base de donnée")
	}

	for _, commerce := range dbCommerces {
		if len(commerce.Services) == 0 {
			continue
		}

		y1, m1, d1 := commerce.LastBilledDate.AddDate(0, 0, 30).Date()
		y2, m2, d2 := time.Now().Date()

		if y1 != y2 || m1 != m2 || d1 != d2 {
			continue
		}

		billedServicesString := ""
		totalToBill := 0.0
		// On doit faire chaque service individuellement
		// Click and Collect à la consommation
		if commerce.DueBalanceClickAndCollectC > 0 {
			roundedPrice := math.Round(commerce.DueBalanceClickAndCollectC*100) / 100
			totalToBill += roundedPrice

			billedServicesString += fmt.Sprintf("<tr><td>Click&Collect</td><td>Consommation</td><td>1</td><td>%.2f€</td><td>20%%</td><td>%.2f€</td></tr>",
				roundedPrice, roundedPrice*1.2)
			commerce.DueBalanceClickAndCollectC = 0
		}

		// Click and Collect au mois
		if commerce.DueBalanceClickAndCollectM > 0 {
			roundedPrice := math.Round(commerce.DueBalanceClickAndCollectM*100) / 100
			totalToBill += roundedPrice

			billedServicesString += fmt.Sprintf("<tr><td>Click&Collect</td><td>Mensuelle</td><td>1</td><td>%.2f€</td><td>20%%</td><td>%.2f€</td></tr>",
				roundedPrice, roundedPrice*1.2)
			commerce.DueBalanceClickAndCollectM = 0
		}

		// Paniers à la consommation
		if commerce.DueBalancePaniersC > 0 {
			roundedPrice := math.Round(commerce.DueBalancePaniersC*100) / 100
			totalToBill += roundedPrice

			billedServicesString += fmt.Sprintf("<tr><td>Paniers</td><td>Consommation</td><td>1</td><td>%.2f€</td><td>20%%</td><td>%.2f€</td></tr>",
				roundedPrice, roundedPrice*1.2)
			commerce.DueBalancePaniersC = 0
		}

		// Paniers au mois
		if commerce.DueBalancePaniersM > 0 {
			roundedPrice := math.Round(commerce.DueBalancePaniersM*100) / 100
			totalToBill += roundedPrice

			billedServicesString += fmt.Sprintf("<tr><td>Paniers</td><td>Mensuelle</td><td>1</td><td>%.2f€</td><td>20%%</td><td>%.2f€</td></tr>",
				roundedPrice, roundedPrice*1.2)
			commerce.DueBalancePaniersM = 0
		}

		if totalToBill > 0 {
			// On doit avant tout vérifier que le paiement passe
			success, err := stripehandler.BillStoresServices(int(math.Round(totalToBill*120)), *commerce.DefaultPaymentMethodID, *commerce.StripID)

			if err != nil || !success {
				fmt.Println("Impossible de débiter le commerce " + commerce.Name)
				continue
			} else {
				fmt.Println("Paiement succes !")
			}

			nowRounded := time.Date(
				time.Now().Year(),
				time.Now().Month(),
				time.Now().Day(),
				1, 0, 0, 0, time.Local,
			)
			commerce.LastBilledDate = &nowRounded

			err = commercesService.Update(&commerce, nil, nil)

			if err != nil {
				fmt.Println("Impossible de mettre à jour le commerce " + commerce.Name)
			}

			// Ensuite on génère la facture et le mail
			invoiceData := htmltopdf.InvoiceData{
				Day:             strconv.Itoa(time.Now().Day()),
				Month:           strconv.Itoa(int(time.Now().Month())),
				Year:            strconv.Itoa(time.Now().Year()),
				InvoiceNumber:   "INV20221100001",
				StoreName:       commerce.Name,
				StoreSIREN:      commerce.Siret[0:9],
				StoreAddress1:   fmt.Sprintf("%s, %s", *commerce.Address.Number, *commerce.Address.Route),
				StoreAddress2:   fmt.Sprintf("%s %s", *commerce.Address.PostalCode, *commerce.Address.City),
				StoreTVA:        "FRXXXXXXXXXXX",
				StoreEmail:      commerce.Email,
				ServicesContent: template.HTML(billedServicesString),
				BilledDate: fmt.Sprintf("%d/%d/%d",
					time.Now().Day(),
					time.Now().Month(),
					time.Now().Year(),
				),
				Card:     "VISA ****-4242",
				TotalHT:  fmt.Sprintf("%.2f", math.Round(totalToBill*100)/100),
				TotalTVA: fmt.Sprintf("%.2f", math.Round(totalToBill*20)/100),
				TotalTTC: fmt.Sprintf("%.2f", math.Round(totalToBill*120)/100),
			}

			htmltopdf.InvoiceToPDF(commerce.ID.Hex(), invoiceData)

			notifications.SendMailBillStoreServices(
				&commerce.Name,
				commerce.Email,
				invoiceData.Month,
				invoiceData.Year,
				invoiceData.BilledDate,
				commerce.ID.Hex(),
				invoiceData.StoreSIREN,
				invoiceData.StoreAddress1+", "+invoiceData.StoreAddress2,
				invoiceData.TotalHT,
				invoiceData.TotalTVA,
				invoiceData.TotalTTC,
				billedServicesString,
			)
		}
	}

	fmt.Println("Le scripte des balance des commerces a été executé")
}
func ExecutreBankingRoutine() {
	sendBalance()
	getServicesPayment()
}
