up-db:
	docker-compose up -d customer-database
	
migrate:
	psql --username=admin --host=localhost --port=5432 --file=migrate.sql

up:
	docker-compose up -d customer-service