package notifications

import (
	"chemin-du-local.bzh/graphql/internal/config"
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
