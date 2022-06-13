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
	"github.com/stripe/stripe-go/v72/setupintent"
)

func calculateOrderAmountForCommerce(commerce model.NewBasketCommerce) (int64, error) {
	result := 0

	databaseCommerce, err := commerces.GetById(commerce.CommerceID)

	if err != nil {
		return 0, err
	}

	if databaseCommerce == nil {
		return 0, &commerces.CommerceErrorNotFound{}
	}

	for _, product := range commerce.Products {
		databaseProduct, err := products.GetById(product.ProductID)

		if err != nil {
			return 0, err
		}

		if databaseProduct == nil {
			return 0, &products.ProductNotFoundError{}
		}

		result = result + int(databaseProduct.Price*product.Quantity*100)
	}

	for _, panier := range commerce.Paniers {
		databasePanier, err := paniers.GetById(panier)

		if err != nil {
			return 0, err
		}

		if databasePanier == nil {
			return 0, &paniers.PanierNotFoundError{}
		}

		result = result + int(databasePanier.Price*100)
	}

	return int64(result), nil
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
		price, err := calculateOrderAmountForCommerce(*commerce)

		if err != nil {
			return err
		}

		databaseCommerce, err := commerces.GetById(commerce.CommerceID)

		if err != nil {
			return err
		}

		if databaseCommerce == nil {
			return &commerces.CommerceErrorNotFound{}
		}

		// La command
		databaseCommerceCommand, err := commands.CommerceCreate(model.NewCommerceCommand{
			CommerceID:    databaseCommerce.ID.Hex(),
			PickupDate:    *commerce.PickupDate,
			PaymentMethod: paymentMethod,
			Price:         int(price),
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

func HanldeCreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	// Set Stripe API key
	apiKey := config.Cfg.Stripe.Key
	stripe.Key = apiKey

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UseStripeSDK    *bool   `json:"useStripeSKD"`
		PaymentMethodID *string `json:"paymentMethodId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var stripeCustomer, _ = authentification(w, r)

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
