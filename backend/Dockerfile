FROM golang:1.24-bookworm

ARG PG_VERSION=17

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

RUN curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-archive-keyring.gpg

RUN echo "deb [signed-by=/usr/share/keyrings/postgresql-archive-keyring.gpg] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list

RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client-${PG_VERSION} \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app && mkdir /etc/backups
RUN chmod 0755 /etc/backups

ENV TZ=Europe/Warsaw
ENV GIN_MODE=release

COPY . /app

WORKDIR /app

EXPOSE 8080

RUN go build -o app

CMD [ "/app/app" ]