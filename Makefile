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

VERTEX_AUTH= --access-token "$$(gcloud auth application-default print-access-token)" --gcp-project "comradequinn" --gcs-bucket "comradequinn-default"

.PHONY: examples
examples: build
	${BIN}/${NAME} --delete-all 2> ${BIN}/debug.log
	${BIN}/${NAME} --verbose "in one sentence, what is the weather like in london tomorrow?" 2>> ${BIN}/debug.log
	${BIN}/${NAME} --continue -v "what about the day after?" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -v -f main.go --pro "in one sentence, summarise this file" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -c -v --stats "is it well written?" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -v --schema="colour:string" "pick a colour of the rainbow" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -v -s="[]colour:string:a rainbow colour" "list all colours of the rainbow" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -v -x --approve "list all files in my current directory" 2>> ${BIN}/debug.log	
	${BIN}/${NAME} -c -x "what do the files indicate may be the purpose of the directory?" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -x "generate a short story of around 50 words and write it to a file named my_story.log" 2>> ${BIN}/debug.log
	${BIN}/${NAME} -a "$$GEMINI_API_TOKEN" -v -x "summarise the targets in my Makefile" 2>> ${BIN}/debug.log	
	${BIN}/${NAME} ${VERTEX_AUTH} -v -x --quiet "list all .go files" 2>> ${BIN}/debug.log
	${BIN}/${NAME} ${VERTEX_AUTH} -c -v -x -q "count them" 2>> ${BIN}/debug.log
	${BIN}/${NAME} ${VERTEX_AUTH} -v -f main.go "how many lines in this file?" 2>> ${BIN}/debug.log
	${BIN}/${NAME} --list
	${BIN}/${NAME} --delete-all
	${BIN}/${NAME} -l
	@rm *.log