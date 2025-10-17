package rest

import (
	"net/http"
	"ordersystem/model"
	"ordersystem/repository"

	"encoding/json"
	"time"

	"github.com/go-chi/render"
)

// GetMenu 			godoc
// @tags 			Menu
// @Description 	Returns the menu of all drinks
// @Produce  		json
// @Success 		200 {array} model.Drink
// @Router 			/api/menu [get]
func GetMenu(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo
		// get slice from db
		// render.Status(r, http.StatusOK)
		// render.JSON(w, r, <your-slice>)
		menu := db.GetDrinks()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, menu)
	}
}

// GetOrders		godoc
// @tags 			Order
// @Description 	Returns all orders
// @Produce  		json
// @Success 		200 {array} model.Order
// @Router 			/api/order/all [get]
func GetOrders(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders := db.GetOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, orders)
	}
}

// GetOrdersTotal 	godoc
// @tags 			Order
// @Description 	Returns a tally of ordered drinks (drink_id -> total amount)
// @Produce  		json
// @Success 		200 {object} map[uint64]uint64
// @Router 			/api/order/total [get]
func GetOrdersTotal(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total := db.GetTotalledOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, total)
	}
}

// PostOrder 		godoc
// @tags 			Order
// @Description 	Adds an order to the db
// @Accept 			json
// @Param 			b body model.Order true "Order"
// @Produce  		json
// @Success 		200
// @Failure     	400
// @Router 			/api/order [post]
func PostOrder(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo
		// declare empty order struct
		var o model.Order
		// err := json.NewDecoder(r.Body).Decode(&<your-order-struct>)
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			// handle error and render Status 400
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid JSON"})
			return
		}
		// basic validation
		if o.DrinkID == 0 || o.Amount <= 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "drink_id and amount are required; amount must be > 0"})
			return
		}

		// handle error and render Status 400
		// add to db
		if o.CreatedAt.IsZero() {
			o.CreatedAt = time.Now().UTC()
		}

		// add to db
		db.AddOrder(&o)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, "ok")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, "ok")
	}
}
