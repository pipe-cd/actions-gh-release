.PHONY: build
build:
	npm run build && npm run package

.PHONY: test
test:
	npm test

.PHONY: dep
dep:
	npm install
