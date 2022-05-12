CMD_DIR = ./cmd
CMD_LIST = $(shell ls $(CMD_DIR))
CMD_BUILD_DIR_LIST = $(foreach c, $(CMD_LIST), $(CMD_DIR)/$(c))  
CMD_BIN_LIST = $(foreach c, $(CMD_LIST), $(CMD_DIR)/$(c)/$(c))  

.PHONY: build
build:
	@for i in $(CMD_BUILD_DIR_LIST);	\
	do					\
		echo go build -o $$i $$i;	\
		go build -o $$i $$i;		\
	done	
	

.PHONY: clean
clean:
	@for i in $(CMD_BIN_LIST);		\
	do					\
		echo rm -f $$i;			\
		rm -f $$i;			\
	done 
