DB_URL=postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable

migrate:
	migrate -path ./migrations -database "$(DB_URL)" up
