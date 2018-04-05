.PHONY: build test clean deploy

build:
	GOOS=linux vgo build -o scribe -i ./src

clean:
	rm -f scribe