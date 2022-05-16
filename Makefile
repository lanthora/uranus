ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
CMD_DIR = $(ROOT_DIR)/cmd
CMD_LIST = $(shell ls $(CMD_DIR))
CMD_BUILD_DIR_LIST = $(foreach c, $(CMD_LIST), $(CMD_DIR)/$(c)) 
CMD_BIN_LIST = $(foreach c, $(CMD_LIST), $(CMD_DIR)/$(c)/$(c))

.PHONY: build
build:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		go build --ldflags="-X 'uranus/pkg/logger.BuildDir=$(ROOT_DIR)/'" -o uranus-$$(basename $$i);	\
	done	
	

.PHONY: clean
clean:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		rm uranus-$$(basename $$i);									\
	done	
