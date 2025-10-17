package model

type Drink struct {
	ID    uint64  `json:"id"`
	NAME  string  `json:"name"`
	PRICE float64 `json:"price"`
	// todo Add fields: Name, Price, Description with suitable types
	// todo json attributes need to be snakecase, i.e. name, created_at, my_variable, ..
}
