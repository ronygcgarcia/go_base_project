# Makefile for managing the Go project

run:
	go run main.go

migrate:
	go run main.go migrate

rollback:
	go run main.go rollback

rollback-step:
	go run main.go rollback:step

migration:
	go run main.go make:migration $(name)

seed:
	go run main.go seed

seed-rollback:
	go run main.go seed:rollback

seed-rollback-step:
	go run main.go seed:rollback:step

seeder:
	go run main.go make:seeder $(name)

activate-auth:
	go run main.go make:activate-auth type=$(type)

