############################################################
#
# Tests
#
############################################################

.PHONY: test
## run all tests excluding fixtures and vendored packages
test: start-database-test 
	STOPWATCH_LOG_LEVEL=info \
	STOPWATCH_ENABLE_DB_LOGS=false \
	STOPWATCH_CLEAN_TEST_DATA=true \
	STOPWATCH_POSTGRES_USER=test \
	STOPWATCH_POSTGRES_DATABASE=test \
	STOPWATCH_POSTGRES_PORT=5433 \
	go test -p 1 -v ./...
