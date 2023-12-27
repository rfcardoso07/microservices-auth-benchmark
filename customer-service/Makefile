up-db:
	docker-compose up -d customer-database
	
migrate:
	psql --username=admin --host=localhost --port=8100 --file=migrate.sql

up:
	docker-compose up -d customer-service