PROG_NAME := tego
SRC_DIR := src
SRC_FILES := $(wildcard $(SRC_DIR)/*.go)
SRC_INSTALL := /usr/local/bin

$(PROG_NAME): $(SRC_FILES)
	go build -o $(PROG_NAME) ./src/

install: $(PROG_NAME)
	sudo install -Dm 755 $(PROG_NAME) $(SRC_INSTALL)/$(PROG_NAME)

clean: $(PROG_NAME)
	rm -f $(PROG_NAME)
	
uninstall: $(PROG_NAME)
	rm -f $(SRC_INSTALL)/$(PROG_NAME)


