package stripehandler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/clickandcollect"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/users"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/setupintent"
)

func calculateOrderAmountForCommerce(commerce model.NewBasketCommerce) (int64, float64, float64, error) {
	result := 0
	resultPaniers := 0.0
	resultClickAndCollect := 0.0

	databaseCommerce, err := commerces.GetById(commerce.CommerceID)

	if err != nil {
		return 0, 0, 0, err
	}

	if databaseCommerce == nil {
		return 0, 0, 0, &commerces.CommerceErrorNotFound{}
	}

	for _, product := range commerce.Products {
		databaseProduct, err := products.GetById(product.ProductID)

		if err != nil {
			return 0, 0, 0, err
		}

		if databaseProduct == nil {
			return 0, 0, 0, &products.ProductNotFoundError{}
		}

		result = result + int(databaseProduct.Price*product.Quantity*100)
		resultClickAndCollect = resultClickAndCollect + databaseProduct.Price*product.Quantity
	}

	for _, panier := range commerce.Paniers {
		databasePanier, err := paniers.GetById(panier)

		if err != nil {
			return 0, 0, 0, err
		}

		if databasePanier == nil {
			return 0, 0, 0, &paniers.PanierNotFoundError{}
		}

		result = result + int(databasePanier.Price*100)
		resultPaniers = resultPaniers + databasePanier.Price
	}

	return int64(result), resultClickAndCollect, resultPaniers, nil
}

func order(user users.User, paymentMethod string, basket model.NewBasket) error {
	databaseCommand, err := commands.Create(model.NewCommand{
		CreationDate: time.Now(),
		User:         user.ID.Hex(),
	})

	if err != nil {
		return err
	}

	for _, commerce := range basket.Commerces {
		price, priceClickAndCollect, pricePaniers, err := calculateOrderAmountForCommerce(*commerce)

		if err != nil {
			return err
		}

		// La command
		databaseCommerceCommand, err := commands.CommerceCreate(model.NewCommerceCommand{
			CommerceID:           commerce.CommerceID,
			PickupDate:           *commerce.PickupDate,
			PaymentMethod:        paymentMethod,
			Price:                int(price),
			PriceClickAndCollect: priceClickAndCollect,
			PricePaniers:         pricePaniers,
		}, databaseCommand.ID)

		if err != nil {
			return err
		}

		// Click & collect
		commandProducts := []*model.NewCCProcuct{}

		for _, product := range commerce.Products {
			commandProducts = append(commandProducts, &model.NewCCProcuct{
				Quantity:  int(product.Quantity),
				ProductID: product.ProductID,
			})
		}

		command := model.NewCCCommand{
			PickupDate: *commerce.PickupDate,
			ProductsID: commandProducts,
		}

		_, err = clickandcollect.Create(databaseCommerceCommand.ID, command)

		if err != nil {
			return err
		}

		// Paniers
		for _, panier := range commerce.Paniers {
			panierCommand := model.NewPanierCommand{
				PanierID:   panier,
				PickupDate: command.PickupDate,
			}

			_, err = paniers.CreateCommand(databaseCommerceCommand.ID, panierCommand)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func authentification(w http.ResponseWriter, r *http.Request) (*string, *users.User) {
	// On doit créer le consumer Stripe si nécesaire
	user := auth.ForContext(r.Context())

	if user == nil {
		http.Error(w, "access denied", http.StatusForbidden)
		return nil, nil
	}

	var stripeCustomer *string

	if user.StripID == nil {
		userFullName := *user.FirstName + " " + *user.LastName
		params := &stripe.CustomerParams{
			Email: &user.Email,
			Name:  &userFullName,
		}

		createdStripeCustomer, err := customer.New(params)
		stripeCustomer = &createdStripeCustomer.ID

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		user.StripID = stripeCustomer
		err = users.Update(user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		stripeCustomer = user.StripID
	}

	return stripeCustomer, user
}

func authentificationForCommerce(w http.ResponseWriter, r *http.Request) (*string, *commerces.Commerce) {
	// On doit créer le consumer Stripe si nécesaire
	user := auth.ForContext(r.Context())

	if user == nil {
		http.Error(w, "access denied", http.StatusForbidden)
		return nil, nil
	}

	commerce, err := commerces.GetForUser(user.ID.Hex())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, nil
	}

	if commerce == nil {
		http.Error(w, "l'utilisateur connecté ne possède pas de commerce", http.StatusBadRequest)
		return nil, nil
	}

	var stripeCustomer *string

	if commerce.StripID == nil {
		commerceStripeName := "(Commerce) " + commerce.Name
		params := &stripe.CustomerParams{
			Email: &user.Email,
			Name:  &commerceStripeName,
		}

		createdStripeCustomer, err := customer.New(params)
		stripeCustomer = &createdStripeCustomer.ID

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		commerce.StripID = stripeCustomer
		err = commerces.Update(commerce, nil, nil)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		stripeCustomer = commerce.StripID
	}

	return stripeCustomer, commerce
}

func HanldeCreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	// Set Stripe API key
	apiKey := config.Cfg.Stripe.Key
	stripe.Key = apiKey

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IsForCommerce   *bool   `json:"isForCommerce"`
		UseStripeSDK    *bool   `json:"useStripeSKD"`
		PaymentMethodID *string `json:"paymentMethodId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var stripeCustomer *string
	var commerce *commerces.Commerce

	if req.IsForCommerce != nil && *req.IsForCommerce {
		stripeCustomer, commerce = authentificationForCommerce(w, r)

		// Il faut mettre à jour le commerce
		if commerce.LastBilledDate == nil {
			now := time.Now().Local()
			commerce.LastBilledDate = &now

			err := commerces.Update(commerce, nil, nil)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		stripeCustomer, _ = authentification(w, r)
	}

	if req.PaymentMethodID != nil {
		params := &stripe.SetupIntentParams{
			Customer:      stripe.String(*stripeCustomer),
			PaymentMethod: req.PaymentMethodID,
		}

		si, err := setupintent.New(params)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("pi.New: %v", err)
			return
		}

		if req.IsForCommerce != nil && *req.IsForCommerce {
			if commerce.DefaultPaymentMethodID != nil {
				paymentmethod.Detach(*commerce.DefaultPaymentMethodID, nil)
			}

			commerce.DefaultPaymentMethodID = req.PaymentMethodID

			err := commerces.Update(commerce, nil, nil)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		writeJSON(w, struct {
			ClientSecret   string `json:"clientSecret"`
			RequiresAction bool   `json:"requiresAction"`
		}{
			ClientSecret:   si.ClientSecret,
			RequiresAction: si.Status == stripe.SetupIntentStatusRequiresAction,
		})
	} else {
		params := &stripe.SetupIntentParams{
			Customer: stripe.String(*stripeCustomer),
		}

		si, err := setupintent.New(params)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("pi.New: %v", err)
			return
		}

		writeJSON(w, struct {
			ClientSecret   string `json:"clientSecret"`
			RequiresAction bool   `json:"requiresAction"`
		}{
			ClientSecret:   si.ClientSecret,
			RequiresAction: si.Status == stripe.SetupIntentStatusRequiresAction,
		})
	}

}

func HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PaymentMethodID *string         `json:"paymentMethodId"`
		Basket          model.NewBasket `json:"basket"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var _, user = authentification(w, r)

	err := order(*user, *req.PaymentMethodID, req.Basket)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, struct {
		PaymentMethod string `json:"paymentMethod"`
		Success       bool   `json:"success"`
	}{
		PaymentMethod: *req.PaymentMethodID,
		Success:       true,
	})
}

func HandleCompleteOrder(w http.ResponseWriter, r *http.Request) {
	// Set Stripe API key
	apiKey := config.Cfg.Stripe.Key
	stripe.Key = apiKey

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CommerceCommandID *string `json:"commerceCommandId"`
		PaymentIntentID   *string `json:"paymentIntentId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var stripeCustomer, _ = authentification(w, r)
	databaseCommerceCommand, err := commands.CommerceGetById(*req.CommerceCommandID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if databaseCommerceCommand == nil {
		http.Error(w, "la commande n'a pas été trouvée", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	databaseCommerceCommand.Status = commands.COMMERCE_COMMAND_STATUS_DONE

	if req.PaymentIntentID != nil {
		err := commands.CommerceUpdate(databaseCommerceCommand)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = commerces.UpdateBalancesForOrder(databaseCommerceCommand.CommerceID.Hex(), databaseCommerceCommand.Price, databaseCommerceCommand.PriceClickAndCollect, databaseCommerceCommand.PricePaniers)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := &stripe.PaymentIntentConfirmParams{
			PaymentMethod: &databaseCommerceCommand.PaymentMethod,
		}

		pi, err := paymentintent.Confirm(*req.PaymentIntentID, params)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, struct {
			ClientSecret string `json:"clientSecret"`
		}{
			ClientSecret: pi.ClientSecret,
		})
	} else {
		// Create a PaymentIntent with amount and currency
		confirm := true
		confirmationMethode := "manual"

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := &stripe.PaymentIntentParams{
			Amount:             stripe.Int64(int64(databaseCommerceCommand.Price)),
			Currency:           stripe.String(string(stripe.CurrencyEUR)),
			PaymentMethod:      &databaseCommerceCommand.PaymentMethod,
			Confirm:            &confirm,
			ConfirmationMethod: &confirmationMethode,
			Customer:           stripeCustomer,
		}

		pi, err := paymentintent.New(params)
		log.Printf("pi.New: %v", pi.ClientSecret)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("pi.New: %v", err)
			return
		}

		if pi.Status == stripe.PaymentIntentStatusSucceeded {
			err := commands.CommerceUpdate(databaseCommerceCommand)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = commerces.UpdateBalancesForOrder(databaseCommerceCommand.CommerceID.Hex(), databaseCommerceCommand.Price, databaseCommerceCommand.PriceClickAndCollect, databaseCommerceCommand.PricePaniers)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		writeJSON(w, struct {
			ClientSecret   string `json:"clientSecret"`
			RequiresAction bool   `json:"requiresAction"`
		}{
			ClientSecret:   pi.ClientSecret,
			RequiresAction: pi.Status == stripe.PaymentIntentStatusRequiresAction,
		})
	}

}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
