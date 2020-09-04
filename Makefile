
BASH_RUN := bash $(DEMON_ENV)

default: binary

.PHONY: all
all:
	$(BASH_RUN) build/make.sh

.PHONY: bianry
binary:
	$(BASH_RUN) build/make.sh binary

.PHONY: rpm
rpm:
	$(BASH_RUN) build/make.sh build-rpm

.PHONY: test-unit
test-unit:
	$(BASH_RUN) build/make.sh test-unit

.PHONY: test-unit-cover
test-unit-cover:
	$(BASH_RUN) build/make.sh test-unit cover

.PHONY: validate
validate:
	$(BASH_RUN) build/make.sh validate-lint validate-gofmt validate-test validate-vet

.PHONY: format
format:
	$(BASH_RUN) build/make.sh format

.PHONY: clean
clean:
	-rm -rf autogen
	-rm -rf bundles
	-rm -f go-lint golint
	-rm -rf docs
