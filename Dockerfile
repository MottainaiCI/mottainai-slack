FROM golang

COPY . /src/

WORKDIR /src

RUN go build

ENTRYPOINT ["/src/mottainai-slack"]
