package repository

import (
	"ordersystem/model"
	"time"
)

type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// todo
func NewDatabaseHandler() *DatabaseHandler {
	// Init the drinks slice with some test data
	drinks := []model.Drink{
		{ID: 1, NAME: "Beer", PRICE: 2.00},
		{ID: 2, NAME: "Spritzer", PRICE: 1.40},
		{ID: 3, NAME: "Coffee", PRICE: 1.00},
	}

	// Init orders slice with some test data
	orders := []model.Order{
		{DrinkID: 1, Amount: 1, CreatedAt: time.Now().UTC()},
	}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

// todo
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {
	// calculate total orders
	totalledOrders := make(map[uint64]uint64)
	for _, order := range db.orders {
		totalledOrders[order.DrinkID] += uint64(order.Amount)
	}
	// key = DrinkID, value = Amount of orders
	// totalledOrders map[uint64]uint64
	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// todo
	// add order to db.orders slice
	db.orders = append(db.orders, *order)
}
