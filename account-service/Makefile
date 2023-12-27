up-db:
	docker-compose up -d account-database
	
migrate:
	psql --username=admin --host=localhost --port=8083 --file=migrate.sql

up:
	docker-compose up -d account-service