up-db:
	docker-compose up -d account-database
	
migrate:
	psql --username=admin --host=localhost --port=8101 --file=migrate.sql

up:
	docker-compose up -d account-service