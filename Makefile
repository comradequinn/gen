.PHONY: build test

NAME=gen
BIN=./bin
TIMESTAMP:=$(shell date -u +'%Y_%m_%d_%H_%M_%S')

.PHONY: build
build:
	@go build -ldflags="-X 'main.commit=dev-build-${TIMESTAMP}'" -o ${BIN}/${NAME}

.PHONY: test
test:
	@go test -count=1 ./...

.PHONY: install
install: build
	@sudo cp ${BIN}/${NAME} /usr/local/bin/${NAME}

#VERTEX_AUTH=
VERTEX_AUTH=--vertex-access-token "$$(gcloud auth application-default print-access-token)" --gcp-project "comradequinn" --gcs-bucket "comradequinn-default"

.PHONY: examples
examples: build
	@${BIN}/${NAME} ${VERTEX_AUTH} --delete-all 2> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n -v "in one sentence, what is the weather like in london tomorrow?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -v "in one sentence, what about the day after?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n --pro -v -f main.go "in one sentence, summarise this file" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -v --stats "is it well written?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n -v --schema="colour:string" "pick a colour of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n -v --schema="[]colour:string:a rainbow colour" "list all colours of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -n -v --exec "list all files in my current directory" 2>> ${BIN}/debug.log	
	@${BIN}/${NAME} ${VERTEX_AUTH} -v --exec "what do the files indicate may be the purpose of the directory?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} -v --exec "copy the last line from any of them into a new file name temp.txt" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} ${VERTEX_AUTH} --list
#	@${BIN}/${NAME} ${VERTEX_AUTH} --delete-all