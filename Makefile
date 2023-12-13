BIN_DIR=bin
DIST_DIR=dist
CMD_DIR=cmd
CSV_DIR=csv
EXECUTABLE=gocsv

.DEFAULT_GOAL := bin
.PHONY: cleanall bin

test:
	cd $(CMD_DIR) && go test -cover
	cd $(CSV_DIR) && go test -cover

bin:
	bash scripts/build-bin.sh

cleanall:
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)

personal: cleanall bin
	cp bin/gocsv ~/bin/gocsv
