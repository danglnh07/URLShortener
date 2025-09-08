postgres:
	sudo docker run --name url_postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres:17.5-alpine3.22

createdb: 
	sudo docker exec -it url_postgres createdb --username=root --owner=root url_shortener

dropdb: 
	sudo docker exec -it url_postgres dropdb url_shortener

initschema:
	sudo docker exec -i url_postgres psql -U root -d url_shortener < ./db/schema/schema.sql

destroyschema:
	sudo docker exec -i	 url_postgres psql -U root -d url_shortener < ./db/schema/destroy.sql

psql: 
	sudo docker exec -it url_postgres psql -U root -d url_shortener

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

run:
	go run main.go

.PHONY: postgres createdb dropdb initschema destroyschema psql sqlc test run 