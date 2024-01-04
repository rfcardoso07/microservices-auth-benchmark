up-dbs:
	docker-compose up -d customer-database
	docker-compose up -d account-database
	docker-compose up -d transaction-database
	docker-compose up -d notification-database
	docker-compose up -d balance-database
	docker-compose up -d auth-database
	
migrate-dbs:
	psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate.sql
	psql --username=admin --host=localhost --port=8083 --file=account-service/migrate.sql
	psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate.sql
	psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate.sql
	psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate.sql
	psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate.sql

up:
	docker-compose up -d customer-service
	docker-compose up -d account-service
	docker-compose up -d transaction-service
	docker-compose up -d notification-service
	docker-compose up -d balance-service
	docker-compose up -d auth-service
	docker-compose up -d gateway