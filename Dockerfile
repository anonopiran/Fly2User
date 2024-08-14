FROM golang:1.22.6-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn"  -o /fly2user

FROM alpine AS app
COPY --from=build /fly2user /fly2user
RUN mkdir /fly2user-data
ENV SUPERVISOR__USER_DB=/fly2user-data/user.sqlite
VOLUME [ "/fly2user-data" ]
ENTRYPOINT [ "/fly2user" ]