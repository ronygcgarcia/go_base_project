package migrations

import (
	"gorm.io/gorm"
	"github.com/ronygcgarcia/go_base_project/models"
)

type Role struct{}

func (m Role) Name() string {
	return "20250507_175657_user"
}

func (m Role) Up(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

func (m Role) Down(db *gorm.DB) error {
	// TODO: Implement rollback logic
	return db.Migrator().DropTable(&models.User{})
}

func init() {
	Register(Role{})
}
