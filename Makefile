ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
CMD_BUILD_DIR_LIST = $(foreach c, $(shell ls $(ROOT_DIR)/cmd), $(ROOT_DIR)/cmd/$(c))

build:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		printf "[%s][build] %s/%s\n" $$(date +"%H:%M:%S") $$(pwd) uranus-$$(basename $$i);		\
		go build -o uranus-$$(basename $$i) -ldflags="-X 'uranus/pkg/logger.BuildDir=$(ROOT_DIR)/'";	\
		printf "[%s][strip] %s/%s\n" $$(date +"%H:%M:%S") $$(pwd) uranus-$$(basename $$i);		\
		strip uranus-$$(basename $$i);									\
	done

clean:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		rm uranus-$$(basename $$i);									\
	done

init:
	go mod tidy

.PHONY: init build clean
