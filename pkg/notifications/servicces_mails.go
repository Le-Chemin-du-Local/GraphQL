package notifications

import "github.com/sendgrid/sendgrid-go/helpers/mail"

func SendMailServicesSubscription(
	receiverName *string,
	receiverEmail string,
	services string,
	billingDate string,
) error {
	templateID := "d-a7dc15b0ef2f4253b80d4eada833d04b"
	p := mail.NewPersonalization()

	receiverNameString := ""
	if receiverName == nil {
		receiverName = &receiverNameString
	}

	p.SetDynamicTemplateData("firstname", *receiverName)
	p.SetDynamicTemplateData("services", services)
	p.SetDynamicTemplateData("billingDate", billingDate)

	return sendMail(
		*receiverName,
		receiverEmail,
		templateID,
		p,
	)
}
