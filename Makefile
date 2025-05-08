run:
	go run main.go

migrate:
	go run main.go migrate

rollback:
	go run main.go rollback

make-migration:
	go run main.go make:migration $(name)
