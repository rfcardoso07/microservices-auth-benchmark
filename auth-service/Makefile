up-db:
	docker-compose up -d auth-database
	
migrate:
	psql --username=admin --host=localhost --port=8091 --file=migrate.sql

up:
	docker-compose up -d auth-service