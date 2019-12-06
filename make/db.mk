.PHONY: start-database
start-database:
	- mkdir -p ./tmp/db/postgres
	docker run \
	    --detach \
		--name stopwatch_db \
		--env POSTGRES_ADMIN_PASSWORD=mysecretpassword \
		--mount type=bind,src=${CURDIR}/tmp/db/postgres,dst=/var/lib/postgresql/data \
		--publish 5432:5432 \
		postgres:9.6 


.PHONY: start-database-test
start-database-test:
	-docker rm -f stopwatch_db_test 
	docker run \
		--detach \
		--name stopwatch_db_test \
		--env POSTGRES_ADMIN_PASSWORD=mysecretpassword \
		--env POSTGRES_USER=test \
		--mount type=bind,src=${CURDIR}/pkg/model,dst=/tmp \
		--publish 5433:5432 \
		postgres:9.6 
	docker exec stopwatch_db_test sh -c 'while ! psql -U test -c "select 1" > /dev/null 2>&1; do echo "Echo waiting for postgres to come up..."; sleep 1; done'
	docker exec -it stopwatch_db_test /bin/bash -c "psql -U test --file /tmp/db.sql"
	
	
