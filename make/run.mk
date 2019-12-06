############################################################
#
# Run the shell
#
############################################################

.PHONY: run-shell
## run the shell mode
run-shell:
    STOPWATCH_POSTGRES_PORT=5432 \
	STOPWATCH_POSTGRES_DATABASE=postgres \
	STOPWATCH_POSTGRES_USER=postgres \
	go run main.go shell

