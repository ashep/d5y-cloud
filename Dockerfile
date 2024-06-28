FROM golang:alpine
COPY . /app
WORKDIR /app
RUN go build
#ENTRYPOINT ["/app/d5y"]
ENTRYPOINT ["sh", "-c", "env && /app/app"]
