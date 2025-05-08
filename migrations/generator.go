package migrations

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func CreateMigrationFile(name string) error {
	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(name, " ", "_")
	filename := fmt.Sprintf("migrations/%s_%s.go", timestamp, safeName)

	content := fmt.Sprintf(`package migrations

import (
	"gorm.io/gorm"
)

type %s struct{}

func (m %s) Name() string {
	return "%s_%s"
}

func (m %s) Up(db *gorm.DB) error {
	// TODO: Implement migration logic
	// Exemple: return db.AutoMigrate(&models.User{})
	return nil
}

func (m %s) Down(db *gorm.DB) error {
	// TODO: Implement rollback logic
	// Exemple: return db.Migrator().DropTable(&models.User{})
	return nil
}

func init() {
	Register(%s{})
}
`, pascalCase(safeName), pascalCase(safeName), timestamp, safeName,
		pascalCase(safeName), pascalCase(safeName), pascalCase(safeName))

	return os.WriteFile(filename, []byte(content), 0644)
}

func pascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
