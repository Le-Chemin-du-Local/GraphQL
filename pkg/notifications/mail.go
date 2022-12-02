package notifications

import (
	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const baseTableHTML = `
<style type="text/css">
.tg {margin-left: auto;margin-right: auto;}
.tg  {border-collapse:collapse;border-color:#ccc;border-spacing:0;}
.tg td{background-color:#fff;border-color:#ccc;border-style:solid;border-width:1px;color:#333;
  font-family:Arial, sans-serif;font-size:14px;overflow:hidden;padding:10px 14px;word-break:normal;}
.tg th{background-color:#f0f0f0;border-color:#ccc;border-style:solid;border-width:1px;color:#333;
  font-family:Arial, sans-serif;font-size:14px;font-weight:normal;overflow:hidden;padding:10px 14px;word-break:normal;}
.tg .tg-0v2f{background-color:#ff8c60;border-color:#ff8c60;color:#ffffff;font-family:Verdana, Geneva, sans-serif !important;
  font-size:12px;font-weight:bold;text-align:center;vertical-align:middle}
.tg .tg-mseo{background-color:#ff8c60;border-color:#ff8c60;color:#ffffff;font-weight:bold;text-align:center;vertical-align:top}
.tg .tg-nbk7{background-color:#ff8c60;border-color:#ff8c60;color:#ffffff;font-family:Verdana, Geneva, sans-serif !important;
  font-size:12px;font-weight:bold;text-align:center;vertical-align:top}
.tg td {border-color:#f0f0f0;font-family:Verdana, Geneva, sans-serif !important;font-size:12px;text-align:left;
  vertical-align:top}
</style>
<table class="tg">
<thead>
  <tr>
    <th class="tg-0v2f">Service</th>
    <th class="tg-mseo">Type de tarif</th>
    <th class="tg-nbk7">Quantité</th>
    <th class="tg-nbk7">Prix HT</th>
    <th class="tg-nbk7">Taux TVA</th>
    <th class="tg-nbk7">Prix TTC</th>
  </tr>
</thead>
<tbody>
{{table_content}}
</tbody>
</table>
`

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
