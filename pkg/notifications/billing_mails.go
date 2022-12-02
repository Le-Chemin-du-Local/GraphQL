package notifications

import (
	"strings"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendMailBillStoreServices(
	receiverName *string,
	receiverEmail string,
	month string,
	year string,
	billingDate string,
	commerceID string,
	commerceSIREN string,
	commerceAddress string,
	totalHT string,
	totalTVA string,
	totalTTC string,
	talbeHTML string,
) error {
	templateID := "d-be11aa69969e421e9a62927c24d5270d"
	p := mail.NewPersonalization()

	receiverNameString := ""
	if receiverName == nil {
		receiverName = &receiverNameString
	}

	p.SetDynamicTemplateData("month", month)
	p.SetDynamicTemplateData("year", year)
	p.SetDynamicTemplateData("commerce_name", *receiverName)
	p.SetDynamicTemplateData("account_id", commerceID)
	p.SetDynamicTemplateData("billing_date", billingDate)
	p.SetDynamicTemplateData("commerce_siren", commerceSIREN)
	p.SetDynamicTemplateData("billing_address", commerceAddress)
	p.SetDynamicTemplateData("table_html", strings.ReplaceAll(baseTableHTML, "{{table_content}}", talbeHTML))
	p.SetDynamicTemplateData("total_ht", totalHT)
	p.SetDynamicTemplateData("total_tva", totalTVA)
	p.SetDynamicTemplateData("total_ttc", totalTTC)

	return sendMail(
		*receiverName,
		receiverEmail,
		templateID,
		p,
	)
}
