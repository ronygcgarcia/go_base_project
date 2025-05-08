package seeders

import (
	"gorm.io/gorm"

	"github.com/ronygcgarcia/go_base_project/models"
)

type InitUsers2 struct{}

func (s InitUsers2) Name() string {
	return "20250508_112055_init_users_2"
}

func (s InitUsers2) Up(db *gorm.DB) error {
	// Example insert for User
	return db.Create([]models.User{
		{Name: "Alice Smithc", Email: "alicec@example.com", Phone: "123456789"},
		{Name: "Bob Johnsonc", Email: "bobc@example.com", Phone: "987654321"},
	}).Error
}

func (s InitUsers2) Down(db *gorm.DB) error {
	// Example delete for User
	return db.Where("email IN ?", []string{"alicec@example.com", "bobc@example.com"}).Delete(&models.User{}).Error
}

func init() {
	Register(InitUsers2{})
}
