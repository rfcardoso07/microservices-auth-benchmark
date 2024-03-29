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

migrate-for-debug:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-debug.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-debug.sql
	
migrate-for-auth-debug:
	$(MAKE) migrate-for-debug
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8091 --file=auth-service/migrate-debug.sql

migrate-for-release:
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8081 --file=customer-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8083 --file=account-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8085 --file=transaction-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8087 --file=notification-service/migrate-release.sql
	PGPASSWORD=admin psql --username=admin --host=localhost --port=8089 --file=balance-service/migrate-release.sql

migrate-for-auth-release:
	$(MAKE) migrate-for-release
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

kube-apply:
	kubectl apply -f kubernetes/customer-deployment.yaml
	kubectl apply -f kubernetes/customer-service.yaml
	kubectl apply -f kubernetes/account-deployment.yaml
	kubectl apply -f kubernetes/account-service.yaml
	kubectl apply -f kubernetes/transaction-deployment.yaml
	kubectl apply -f kubernetes/transaction-service.yaml
	kubectl apply -f kubernetes/notification-deployment.yaml
	kubectl apply -f kubernetes/notification-service.yaml
	kubectl apply -f kubernetes/balance-deployment.yaml
	kubectl apply -f kubernetes/balance-service.yaml
	kubectl apply -f kubernetes/auth-deployment.yaml
	kubectl apply -f kubernetes/auth-service.yaml
	kubectl apply -k kubernetes/gateway-deployment.yaml
	kubectl apply -k kubernetes/gateway-service.yaml

kube-delete:
	kubectl delete deployment customer-service
	kubectl delete service customer-service
	kubectl delete deployment account-service
	kubectl delete service account-service
	kubectl delete deployment transaction-service
	kubectl delete service transaction-service
	kubectl delete deployment notification-service
	kubectl delete service notification-service
	kubectl delete deployment balance-service
	kubectl delete service balance-service
	kubectl delete deployment auth-service
	kubectl delete service auth-service
	kubectl delete deployment gateway
	kubectl delete service gateway