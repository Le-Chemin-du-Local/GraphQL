package notifications

import (
	"strconv"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func sendMail(
	receiverName string,
	receiverEmail string,
	templateID string,
	personalization *mail.Personalization,
) error {
	m := mail.NewV3Mail()
	from := mail.NewEmail("Le Chemin du Local", "contact@chemin-du-local.bzh")
	to := mail.NewEmail(receiverName, receiverEmail)

	personalization.AddTos(to)
	m.SetFrom(from)
	m.SetTemplateID(templateID)
	m.AddPersonalizations(personalization)

	// On envoie l'email ici
	request := sendgrid.GetRequest(config.Cfg.SendGrid.Key, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"

	var Body = mail.GetRequestBody(m)
	request.Body = Body

	_, err := sendgrid.API(request)
	return err
}

func SendMailWelcome(
	receiverName *string,
	receiverEmail string,
) error {
	templateID := "d-860a27a33e9b46ee925a0e46b678f16c"
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

func SendMailWelcomeStoreKeeper(
	receiverName *string,
	receiverEmail string,
) error {
	templateID := "d-352769f792944888952cc0c073c8b482"
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

func SendMailNewCommand(
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

		schedulesListString += "<i>" + commerce.PickupDate.Format("02/01/2006 15:04") + " - " + commerce.PickupDate.Add(time.Minute*time.Duration(30)).Format("15:04") + "</i>"
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
