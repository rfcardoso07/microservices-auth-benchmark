up-dbs:
	docker-compose up -d customer-database
	docker-compose up -d account-database
	docker-compose up -d transaction-database
	docker-compose up -d notification-database
	docker-compose up -d balance-database

up-all-dbs:
	$(MAKE) up-dbs
	docker-compose up -d auth-database

migrate-debug:
	psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-debug.sql
	psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-debug.sql
	psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-debug.sql
	psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-debug.sql
	psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-debug.sql
	
migrate-auth-debug:
	$(MAKE) migrate-debug
	psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate-debug.sql

migrate-release:
	psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-release.sql
	psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-release.sql
	psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-release.sql
	psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-release.sql
	psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-release.sql

migrate-auth-release:
	$(MAKE) migrate-release
	psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate-release.sql

noauth-debug:
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH customer-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH account-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH transaction-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH notification-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH balance-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

edge-debug:
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH customer-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH account-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH transaction-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH notification-service
	docker-compose up -d -e MODE=debug -e PATTERN=NO_AUTH balance-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=TRUE gateway

centralized-debug:
	docker-compose up -d -e MODE=debug -e PATTERN=CENTRALIZED customer-service
	docker-compose up -d -e MODE=debug -e PATTERN=CENTRALIZED account-service
	docker-compose up -d -e MODE=debug -e PATTERN=CENTRALIZED transaction-service
	docker-compose up -d -e MODE=debug -e PATTERN=CENTRALIZED notification-service
	docker-compose up -d -e MODE=debug -e PATTERN=CENTRALIZED balance-service
	docker-compose up -d -e MODE=debug auth-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

decentralized-debug:
	docker-compose up -d -e MODE=debug -e PATTERN=DECENTRALIZED customer-service
	docker-compose up -d -e MODE=debug -e PATTERN=DECENTRALIZED account-service
	docker-compose up -d -e MODE=debug -e PATTERN=DECENTRALIZED transaction-service
	docker-compose up -d -e MODE=debug -e PATTERN=DECENTRALIZED notification-service
	docker-compose up -d -e MODE=debug -e PATTERN=DECENTRALIZED balance-service
	docker-compose up -d -e MODE=debug auth-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

noauth-release:
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH customer-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH account-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH transaction-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH notification-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH balance-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

edge-release:
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH customer-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH account-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH transaction-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH notification-service
	docker-compose up -d -e MODE=release -e PATTERN=NO_AUTH balance-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=TRUE gateway

centralized-release:
	docker-compose up -d -e MODE=release -e PATTERN=CENTRALIZED customer-service
	docker-compose up -d -e MODE=release -e PATTERN=CENTRALIZED account-service
	docker-compose up -d -e MODE=release -e PATTERN=CENTRALIZED transaction-service
	docker-compose up -d -e MODE=release -e PATTERN=CENTRALIZED notification-service
	docker-compose up -d -e MODE=release -e PATTERN=CENTRALIZED balance-service
	docker-compose up -d -e MODE=release auth-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

decentralized-release:
	docker-compose up -d -e MODE=release -e PATTERN=DECENTRALIZED customer-service
	docker-compose up -d -e MODE=release -e PATTERN=DECENTRALIZED account-service
	docker-compose up -d -e MODE=release -e PATTERN=DECENTRALIZED transaction-service
	docker-compose up -d -e MODE=release -e PATTERN=DECENTRALIZED notification-service
	docker-compose up -d -e MODE=release -e PATTERN=DECENTRALIZED balance-service
	docker-compose up -d -e MODE=release auth-service
	docker-compose up -d -e MODE=debug -e EDGE_AUTH=FALSE gateway

up-noauth:
	$(MAKE) up-dbs
	$(MAKE) migrate-release
	$(MAKE) noauth-release

up-edge:
	$(MAKE) up-all-dbs
	$(MAKE) migrate-auth-release
	$(MAKE) edge-release

up-centralized:
	$(MAKE) up-all-dbs
	$(MAKE) migrate-auth-release
	$(MAKE) centralized-release

up-decentralized:
	$(MAKE) up-all-dbs
	$(MAKE) migrate-auth-release
	$(MAKE) decentralized-release