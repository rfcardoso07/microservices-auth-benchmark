TAG:=1.0

build-tag-push:
	docker build --no-cache -t customer-service:${TAG} ./customer-service
	docker build --no-cache -t account-service:${TAG} ./account-service
	docker build --no-cache -t transaction-service:${TAG} ./transaction-service
	docker build --no-cache -t notification-service:${TAG} ./notification-service
	docker build --no-cache -t balance-service:${TAG} ./balance-service
	docker build --no-cache -t auth-service:${TAG} ./auth-service
	docker build --no-cache -t gateway:${TAG} ./gateway
	docker tag customer-service:${TAG} rfcardoso07/customer-service:${TAG}
	docker tag account-service:${TAG} rfcardoso07/account-service:${TAG}
	docker tag transaction-service:${TAG} rfcardoso07/transaction-service:${TAG}
	docker tag notification-service:${TAG} rfcardoso07/notification-service:${TAG}
	docker tag balance-service:${TAG} rfcardoso07/balance-service:${TAG}
	docker tag auth-service:${TAG} rfcardoso07/auth-service:${TAG}
	docker tag gateway:${TAG} rfcardoso07/gateway:${TAG}
	docker push rfcardoso07/customer-service:${TAG}
	docker push rfcardoso07/account-service:${TAG}
	docker push rfcardoso07/transaction-service:${TAG}
	docker push rfcardoso07/notification-service:${TAG}
	docker push rfcardoso07/balance-service:${TAG}
	docker push rfcardoso07/auth-service:${TAG}
	docker push rfcardoso07/gateway:${TAG}

up-dbs:
	docker-compose up -d customer-database
	docker-compose up -d account-database
	docker-compose up -d transaction-database
	docker-compose up -d notification-database
	docker-compose up -d balance-database

up-all-dbs:
	$(MAKE) up-dbs
	docker-compose up -d auth-database

clean-dbs:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=customer-service/clean.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=account-service/clean.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=transaction-service/clean.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=notification-service/clean.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=balance-service/clean.sql

migrate-debug:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-debug.sql
	
migrate-auth-debug:
	$(MAKE) migrate-debug
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate-debug.sql

migrate-release:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-release.sql

migrate-auth-release:
	$(MAKE) migrate-release
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate-release.sql

add-users:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=scripts/add-users.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=scripts/add-users.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=scripts/add-users.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=scripts/add-users.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=scripts/add-users.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8091 --file=scripts/add-users.sql

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
	MODE=debug PATTERN=CENTRALIZED EDGE_AUTH=FALSE docker-compose up -d transaction-service
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