ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
CMD_BUILD_DIR_LIST = $(foreach c, $(shell ls $(ROOT_DIR)/cmd), $(ROOT_DIR)/cmd/$(c))

build:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		go build --ldflags="-X 'uranus/pkg/logger.BuildDir=$(ROOT_DIR)/'" -o uranus-$$(basename $$i);	\
	done

clean:
	@for i in $(CMD_BUILD_DIR_LIST);									\
	do													\
		cd  $$i;											\
		rm uranus-$$(basename $$i);									\
	done

.PHONY: build clean
