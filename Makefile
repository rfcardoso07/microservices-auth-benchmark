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

add-users:
	psql --username=admin --host=localhost --port=8081 --file=add-users.sql
	psql --username=admin --host=localhost --port=8083 --file=add-users.sql
	psql --username=admin --host=localhost --port=8085 --file=add-users.sql
	psql --username=admin --host=localhost --port=8087 --file=add-users.sql
	psql --username=admin --host=localhost --port=8089 --file=add-users.sql
	psql --username=admin --host=localhost --port=8091 --file=add-users.sql

noauth-debug:
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d transaction-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d gateway

edge-debug:
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d customer-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d account-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d transaction-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d notification-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d balance-service
	MODE=debug PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d gateway

centralized-debug:
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE up -d transaction-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d auth-service
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d gateway

decentralized-debug:
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d transaction-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d auth-service
	MODE=debug PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d gateway

noauth-release:
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d transaction-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=FALSE docker-compose up -d gateway

edge-release:
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d customer-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d account-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d transaction-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d notification-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d balance-service
	MODE=release PATTERN=NO_AUTH EDGE_AUTH=TRUE docker-compose up -d gateway

centralized-release:
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d transaction-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d auth-service
	MODE=release PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d gateway

decentralized-release:
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d customer-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d account-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d transaction-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d notification-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d balance-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d auth-service
	MODE=release PATTERN=DECENTRALIZED EDGE_AUTH=FALSE docker-compose up -d gateway

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