package notifications

import (
	"strconv"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Commerçants

func SendMailNewCommand(
	receiverName *string,
	receiverEmail string,
) error {
	templateID := "d-bd3c098dab99485cb6cc8434608eb5d0"
	p := mail.NewPersonalization()

	receiverNameString := ""
	if receiverName == nil {
		receiverName = &receiverNameString
	}

	p.SetDynamicTemplateData("firstname", *receiverName)

	return sendMail(
		*receiverName,
		receiverEmail,
		templateID,
		p,
	)
}

// Clients

func SendMailCommandConfirmation(
	receiverName *string,
	receiverEmail string,
	basket model.NewBasket,
	commandNumber string,
	price int,
	commercesService commerces.CommercesService,
	productsService products.ProductsService,
	paniersService paniers.PaniersService,
) error {
	templateID := "d-e6343e0de4e442bfac4ba732578015a1"
	p := mail.NewPersonalization()

	receiverNameString := ""
	if receiverName == nil {
		receiverName = &receiverNameString
	}

	// On doit faire la lite des crénaux
	schedulesListString := ""
	basketRecapString := ""

	for _, commerce := range basket.Commerces {
		databaseCommerce, err := commercesService.GetById(commerce.CommerceID)

		if err != nil {
			schedulesListString += "<strong>Commerce Inconnu</strong><br/>"
		} else {
			schedulesListString += "<strong>" + databaseCommerce.Name + "</strong> - "
			schedulesListString += *databaseCommerce.Address.Number + " " + *databaseCommerce.Address.Route + ", "
			if databaseCommerce.Address.OptionalRoute != nil {
				schedulesListString += *databaseCommerce.Address.OptionalRoute + ", "
			}
			schedulesListString += *databaseCommerce.Address.City + " " + *databaseCommerce.Address.PostalCode + "<br/>"

		}

		schedulesListString += "<i>À retirer le" + commerce.PickupDate.Format("02/01/2006") + " entre " + commerce.PickupDate.Format("15:04") + " et " + commerce.PickupDate.Add(time.Minute*time.Duration(30)).Format("15:04") + "</i>"
		schedulesListString += "<br/><hr/>"

		// On liste les produits
		basketRecapString += "<strong>" + databaseCommerce.Name + "</strong><br/><ul>"
		for _, product := range commerce.Products {
			databaseProduct, err := productsService.GetById(product.ProductID)

			if err != nil {
				basketRecapString += "<li><strong>(x" + strconv.FormatInt(int64(product.Quantity), 10) + ")</strong> Inconnue</li>"
			} else {
				basketRecapString += "<li><strong>(x" + strconv.FormatInt(int64(product.Quantity), 10) + ")</strong> " +
					databaseProduct.Name + "</li>"
			}
		}

		for _, panier := range commerce.Paniers {
			databasePanier, err := paniersService.GetById(panier)

			if err != nil {
				basketRecapString += "<li><strong>(x1)</strong> Panier Inconnue</li>"
			} else {
				basketRecapString += "<li><strong>(x1)</strong>" + databasePanier.Name + "</li>"
			}
		}
		basketRecapString += "</ul><hr/>"
	}

	p.SetDynamicTemplateData("numcommande", commandNumber)
	p.SetDynamicTemplateData("firstname", *receiverName)
	p.SetDynamicTemplateData("listcreneaux", schedulesListString)
	p.SetDynamicTemplateData("recapbasket", basketRecapString)
	p.SetDynamicTemplateData("totalprice", price/100)

	return sendMail(
		*receiverName,
		receiverEmail,
		templateID,
		p,
	)
}
