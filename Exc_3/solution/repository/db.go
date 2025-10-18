package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"ordersystem/model"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseHandler struct {
	dbConn *gorm.DB
}

func NewDatabaseHandler() (*DatabaseHandler, error) {
	slog.Info("Connecting to database")
	// connect to db
	dsn, err := getDsn()
	if err != nil {
		return nil, err
	}
	dbConn, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// create tables and migrate
	err = dbConn.AutoMigrate(&model.Drink{}, &model.Order{})
	if err != nil {
		return nil, err
	}
	// add test data to db
	err = prepopulate(dbConn)
	if err != nil {
		return nil, err
	}
	return &DatabaseHandler{dbConn: dbConn}, nil
}

func getDsn() (string, error) {
	dbUser, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		return "", errors.New("environment variable 'POSTGRES_USER' is not set")
	}
	dbPw, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		return "", errors.New("environment variable 'POSTGRES_PASSWORD' is not set")
	}
	dbName, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		return "", errors.New("environment variable 'POSTGRES_DB' is not set")
	}
	dbPort, ok := os.LookupEnv("POSTGRES_TCP_PORT")
	if !ok {
		return "", errors.New("environment variable 'POSTGRES_TCP_PORT' is not set")
	}
	dbHost, ok := os.LookupEnv("DB_HOST")
	if !ok {
		return "", errors.New("environment variable 'DB_HOST' is not set")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPw, dbName, dbPort)
	return dsn, nil
}

func prepopulate(dbConn *gorm.DB) error {
	// check if prepopulate has already run once
	var exists bool
	err := dbConn.Model(&model.Drink{}).
		Select("count(*) > 0").
		Find(&exists).
		Error
	if err != nil {
		return err
	}
	if exists {
		// don't prepopulate if has already run
		return nil
	}

	// create drink menu
	drinks := []model.Drink{
		{
			Name:        "Beer",
			Price:       2.00,
			Description: "Refreshing alcoholic beverage",
		},
		{
			Name:        "Spritzer",
			Price:       1.40,
			Description: "Wine mixed with sparkling water",
		},
		{
			Name:        "Coffee",
			Price:       1.00,
			Description: "Hot caffeinated beverage",
		},
	}

	// Insert drinks into database
	for _, drink := range drinks {
		err = dbConn.Create(&drink).Error
		if err != nil {
			return fmt.Errorf("failed to create drink %s: %w", drink.Name, err)
		}
		slog.Info("Created drink", "name", drink.Name, "price", drink.Price)
	}

	// Optionally create some sample orders
	// Get the created drinks first
	var createdDrinks []model.Drink
	err = dbConn.Find(&createdDrinks).Error
	if err != nil {
		return fmt.Errorf("failed to retrieve created drinks: %w", err)
	}

	// Create some sample orders
	sampleOrders := []model.Order{
		{DrinkID: createdDrinks[0].ID, Amount: 2}, // 2 Beers
		{DrinkID: createdDrinks[1].ID, Amount: 1}, // 1 Spritzer
		{DrinkID: createdDrinks[2].ID, Amount: 3}, // 3 Coffees
	}

	for _, order := range sampleOrders {
		err = dbConn.Create(&order).Error
		if err != nil {
			return fmt.Errorf("failed to create sample order: %w", err)
		}
	}

	slog.Info("Database prepopulation completed successfully")
	return nil
}

func (db *DatabaseHandler) GetDrinks() (drinks []model.Drink, err error) {
	err = db.dbConn.Find(&drinks).Error
	if err != nil {
		return nil, err
	}
	return drinks, nil
}

func (db *DatabaseHandler) GetOrders() (orders []model.Order, err error) {
	err = db.dbConn.Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

const totalledStmt = `SELECT drink_id, SUM(amount) AS total_amount_ordered FROM orders WHERE deleted_at IS NULL GROUP BY drink_id ORDER BY drink_id;`

func (db *DatabaseHandler) GetTotalledOrders() (totals []model.DrinkOrderTotal, err error) {
	err = db.dbConn.Raw(totalledStmt).Scan(&totals).Error
	if err != nil {
		return nil, err
	}
	return totals, nil
}

func (db *DatabaseHandler) AddOrder(order *model.Order) error {
	err := db.dbConn.Create(order).Error
	if err != nil {
		return err
	}
	return nil
}
