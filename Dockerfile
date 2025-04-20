FROM golang:1.24-alpine

WORKDIR /gopapageno

# Install git and basic build tools along with a shell for better interactive experience
RUN apk add --no-cache git make gcc libc-dev bash vim curl python3

RUN go install golang.org/x/perf/cmd/benchstat@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Prepare test JSON files

RUN if [ -f examples/json/data/emojis.json ] && [ -f scripts/json_mult.py ]; then \
        python3 scripts/json_mult.py examples/json/data/emojis.json 1000 && \
        rm examples/json/data/emojis.json; \
    else \
        echo "Warning: examples/json/data/emojis.json or scripts/json_mult.py not found"; \
    fi

RUN if [ -f examples/json/data/citylots.zip ]; then \
        unzip -o examples/json/data/citylots.zip -d examples/json/data/ && \
        rm examples/json/data/citylots.zip; \
    else \
        echo "Warning: examples/json/data/citylots.zip not found"; \
    fi

RUN if [ -f examples/json/data/wikidata-lexemes.zip ]; then \
        unzip -o examples/json/data/wikidata-lexemes.zip -d examples/json/data/ && \
        rm examples/json/data/wikidata-lexemes.zip; \
    else \
        echo "Warning: examples/json/data/wikidata-lexemes.zip not found"; \
  fi

RUN 

# Build GoPAPAGENO
RUN cd cmd/gopapageno && go build -o /go/bin/gopapageno

CMD ["/bin/bash"]