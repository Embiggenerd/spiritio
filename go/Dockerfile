FROM golang:1.22.5


WORKDIR /spiritio
COPY . .


RUN CGO_ENABLED=1 go build -o spiritio ./cmd

# For documentation only
EXPOSE 8080 

ENTRYPOINT ["./spiritio"]
