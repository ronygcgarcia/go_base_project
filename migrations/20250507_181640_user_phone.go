package migrations

import (
    "gorm.io/gorm"
    "github.com/ronygcgarcia/go_base_project/models"
)

type AddPhoneToUser struct{}

func (m AddPhoneToUser) Name() string {
    return "20250507_181640_user_phone"
}

func (m AddPhoneToUser) Up(db *gorm.DB) error {
    if !db.Migrator().HasColumn(&models.User{}, "Phone") {
        return db.Migrator().AddColumn(&models.User{}, "Phone")
    }
    return nil
}

func (m AddPhoneToUser) Down(db *gorm.DB) error {
    if db.Migrator().HasColumn(&models.User{}, "Phone") {
        return db.Migrator().DropColumn(&models.User{}, "Phone")
    }
    return nil
}

func init() {
    Register(AddPhoneToUser{})
}
