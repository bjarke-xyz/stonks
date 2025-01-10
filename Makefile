.PHONY: build generate

BINARY_NAME=stonks

npm-ci:
	npm ci

# npm-build-prod:
# 	cp node_modules/htmx.org/dist/htmx.min.js internal/web/static/js/vendor && \
# 	cp node_modules/htmx-ext-sse/sse.js internal/web/static/js/vendor && \
# 	cp node_modules/chart.js/dist/chart.umd.js internal/web/static/js/vendor

# npm-build-dev: npm-build-prod
# 	cp node_modules/chart.js/dist/chart.umd.js.map internal/web/static/js/vendor

build: npm-ci generate
	go mod tidy && \
	go build -ldflags="-w -s" -o ${BINARY_NAME} cmd/web/main.go

dev: generate
	templ generate --watch --cmd="go run cmd/web/main.go"

generate:
	sqlc generate
	templ generate
	npx tailwindcss build -i internal/web/static/css/style.css -o internal/web/static/css/tailwind.css -m

# rm -f internal/web/static/js/vendor/*.js
# rm -f internal/web/static/js/vendor/*.js.map
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f internal/web/static/css/tailwind.css
	rm -rf cache/*
	touch cache/.gitkeep

