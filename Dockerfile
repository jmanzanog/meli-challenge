FROM golang:1.16.7 AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /out/app .
FROM postgres:13-alpine
ENV POSTGRES_PASSWORD=postgres
WORKDIR /
COPY --from=builder /out/app /
COPY create_database.sql /docker-entrypoint-initdb.d/init.sql
RUN apk add --no-cache ca-certificates supervisor coreutils && rm -rf /var/cache/apk/*
COPY docker-entrypoint.sh /usr/local/bin/
RUN ["chmod", "755", "/usr/local/bin/docker-entrypoint.sh"]
RUN ["chmod", "755", "/docker-entrypoint-initdb.d/init.sql"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
ADD supervisor.conf /etc/supervisor.conf
EXPOSE 4500 5432
CMD ["supervisord", "-c", "/etc/supervisor.conf"]