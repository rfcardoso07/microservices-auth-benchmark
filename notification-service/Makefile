up-db:
	docker-compose up -d notification-database
	
migrate:
	psql --username=admin --host=localhost --port=8087 --file=migrate.sql

up:
	docker-compose up -d notification-service