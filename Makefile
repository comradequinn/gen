.PHONY: build test

NAME=gen
BIN=./bin

.PHONY: build
build:
	@go build -o ${BIN}/${NAME}

.PHONY: test
test:
	@go test -count=1 ./...

VERTEX_AUTH=
#VERTEX_AUTH=--vertex-access-token "$$(gcloud auth application-default print-access-token)" --gcp-project "comradequinn" --gcs-bucket "comradequinn-default"

.PHONY: examples
examples: build
	@${BIN}/${NAME} ${VERTEX_AUTH} --delete-all
	@${BIN}/${NAME} ${VERTEX_AUTH} -n --debug "in one sentence, what is the weather like in london tomorrow?" 2> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} --flash --debug "in one sentence, what about the day after?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n --debug -f main.go "in one sentence, summarise this file" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} --debug --stats "is it well written?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n --debug --schema="colour:string" "pick a colour of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n --debug --schema="[]colour:string" "list all colours of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} --list
	@${BIN}/${NAME} ${VERTEX_AUTH} --delete-all