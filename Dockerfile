FROM --platform=linux/amd64 debian:bullseye-slim AS tts-builder

RUN apt-get update && \
    apt-get install -y software-properties-common && \
    echo "deb http://deb.debian.org/debian bullseye contrib non-free" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y libttspico-utils libttspico0 libpopt0

RUN apt-get update && \
    apt-get install -y software-properties-common && \
    apt-get update && \
    apt-get install -y libttspico-utils libttspico0 && \
    mkdir -p /copy-root/usr/bin \
             /copy-root/lib/x86_64-linux-gnu \
             /copy-root/lib64 \
             /copy-root/usr/lib/x86_64-linux-gnu && \
    #  ensure architecture-specific paths
    mkdir -p $(find /usr/lib -name "*-linux-gnu" -type d | sed 's|^|/copy-root|') && \
    # Copy the binary and its direct dependencies
    cp $(which pico2wave) /copy-root/usr/bin/ && \
    # Find and copy all ttspico libraries
    find /usr/lib -name "libttspico*.so*" -exec cp {} /copy-root/usr/lib/x86_64-linux-gnu/ \; && \
    # required system libraries
    cp /usr/lib/x86_64-linux-gnu/libc.so.6 /copy-root/usr/lib/x86_64-linux-gnu/ || \
    cp /lib/x86_64-linux-gnu/libc.so.6 /copy-root/lib/x86_64-linux-gnu/ && \
    cp /usr/lib/x86_64-linux-gnu/libm.so.6 /copy-root/usr/lib/x86_64-linux-gnu/ || \
    cp /lib/x86_64-linux-gnu/libm.so.6 /copy-root/lib/x86_64-linux-gnu/ && \
    # the dynamic linker
    cp /usr/lib64/ld-linux-x86-64.so.2 /copy-root/lib64/ || \
    cp /lib64/ld-linux-x86-64.so.2 /copy-root/lib64/

RUN ls -laR /copy-root/ > /copy-root-contents.txt && cat /copy-root-contents.txt

RUN cp /usr/lib/x86_64-linux-gnu/libpopt* /copy-root/usr/lib/x86_64-linux-gnu/

RUN mkdir -p /copy-root/usr/share/pico/lang && \
cp -r /usr/share/pico/lang/* /copy-root/usr/share/pico/lang/

# Stage 2 build app
FROM --platform=linux/amd64 golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o go_quote

# Stage 3: Create the final minimal image
FROM scratch

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=tts-builder /copy-root /
COPY --from=builder /app/go_quote .

# Copy the data file
COPY data/quotes.bytesz data/quotes.bytesz

# Set GOMAXPROCS environment variable
# This version will run under a single core.
ENV GOMAXPROCS=1

ENTRYPOINT ["./go_quote", "-FILENAME", "data/quotes.bytesz", "-STORAGE", "bytesz"]
