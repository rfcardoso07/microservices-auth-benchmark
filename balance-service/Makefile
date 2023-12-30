up-db:
	docker-compose up -d balance-database
	
migrate:
	psql --username=admin --host=localhost --port=8089 --file=migrate.sql

up:
	docker-compose up -d balance-service