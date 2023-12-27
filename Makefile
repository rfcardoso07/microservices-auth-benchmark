up-dbs:
	docker-compose up -d customer-database
	docker-compose up -d account-database
	
migrate-dbs:
	psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate.sql
	psql --username=admin --host=localhost --port=8083 --file=account-service/migrate.sql

up:
	docker-compose up -d customer-service
	docker-compose up -d account-service