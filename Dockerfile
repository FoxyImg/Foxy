###
# Base image containing vips and required dependencies
###
FROM debian:bookworm-slim as foxyvips
RUN apt update \
    && apt install -y  \
    fontconfig \
    libde265-0 \
    libheif1 \
    libvips


###
# Builds the go app that is using govips
###
FROM foxyvips as builder

RUN apt update
RUN apt install -y \
	software-properties-common \
	build-essential \
    pkg-config \
    libvips-dev \
	unzip \
	wget

WORKDIR /tmp

RUN wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz \
    && tar -C /usr/local/ -xzf go1.22.4.linux-amd64.tar.gz

ENV OG_PATH="${PATH}"
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=1
ENV CGO_CFLAGS_ALLOW="-Xpreprocessor"
ENV CGO_LDFLAGS="-L/usr/local/lib/vips/lib"
ENV GOOS=linux
ENV GO111MODULE=on

COPY . .
RUN go build .

###
# Final app image
###
FROM foxyvips

COPY --from=builder /app/foxy /foxy

WORKDIR /

EXPOSE 8080
CMD ["/foxy"]
