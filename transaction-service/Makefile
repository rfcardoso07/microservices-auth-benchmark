up-db:
	docker-compose up -d transaction-database
	
migrate:
	psql --username=admin --host=localhost --port=8085 --file=migrate.sql

up:
	docker-compose up -d transaction-service