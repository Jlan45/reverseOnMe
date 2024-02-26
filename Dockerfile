FROM golang:latest
COPY ./ /app
WORKDIR /app
RUN go build -o ReverseOnMe
EXPOSE 8081
ENTRYPOINT [ "/app/ReverseOnMe"]
