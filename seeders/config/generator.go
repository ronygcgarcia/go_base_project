package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var knownModels = map[string]string{
	"user":        "User",
	"users":       "User",
	"role":        "Role",
	"roles":       "Role",
	"permission":  "Permission",
	"permissions": "Permission",
}

func CreateSeederFile(name string) error {
	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	filename := fmt.Sprintf("seeders/%s_%s.go", timestamp, safeName)

	structName := pascalCase(safeName)
	model := detectModel(safeName)
	importModels := model != ""

	content := fmt.Sprintf(`package seeders

import (
	"gorm.io/gorm"%s
)

type %s struct{}

func (s %s) Name() string {
	return "%s_%s"
}

func (s %s) Up(db *gorm.DB) error {
%s
	return nil
}

func (s %s) Down(db *gorm.DB) error {
%s
	return nil
}

func init() {
	Register(%s{})
}
`,
		optionalImport(importModels),
		structName, structName, timestamp, safeName,
		structName, generateUpExample(model),
		structName, generateDownExample(model),
		structName,
	)

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

func detectModel(safeName string) string {
	for k, v := range knownModels {
		if strings.Contains(safeName, k) {
			return v
		}
	}
	return ""
}

func optionalImport(add bool) string {
	if add {
		return `

	"github.com/ronygcgarcia/go_base_project/models"`
	}
	return ""
}

func generateUpExample(model string) string {
	if model == "" {
		return `	// TODO: Insert seed data`
	}
	return fmt.Sprintf(`	// Example insert for %s
	// return db.Create([]models.%s{
	// 	{Name: "Example 1"},
	// 	{Name: "Example 2"},
	// }).Error`, model, model)
}

func generateDownExample(model string) string {
	if model == "" {
		return `	// TODO: Revert seed data`
	}
	return fmt.Sprintf(`	// Example delete for %s
	// return db.Where("name IN ?", []string{"Example 1", "Example 2"}).Delete(&models.%s{}).Error`, model, model)
}
