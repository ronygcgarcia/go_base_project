package seeders

import (
	"gorm.io/gorm"

	"github.com/ronygcgarcia/go_base_project/models"
)

type InitUsers struct{}

func (s InitUsers) Name() string {
	return "20250508_111822_init_users"
}

func (s InitUsers) Up(db *gorm.DB) error {
	// Example insert for User
	return db.Create([]models.User{
		{Name: "Alice Smith", Email: "alice@example.com", Phone: "123456789"},
		{Name: "Bob Johnson", Email: "bob@example.com", Phone: "987654321"},
	}).Error
}

func (s InitUsers) Down(db *gorm.DB) error {
	// Example delete for User
	return db.Where("email IN ?", []string{"alice@example.com", "bob@example.com"}).Delete(&models.User{}).Error
}

func init() {
	Register(InitUsers{})
}
