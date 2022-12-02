package notifications

import "github.com/sendgrid/sendgrid-go/helpers/mail"

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
