#FROM rust:alpine
#FROM rust
FROM ubuntu

WORKDIR /
COPY . /.

# Update and upgrade repo
RUN DEBIAN_FRONTEND=noninteractive apt-get -y update
# Install tools we might need
#RUN DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y -q curl build-essential ca-certificates git 

# Download IntelMKL
#RUN DEBIAN_FRONTEND=noninteractive apt -y install libmkl-dev

# Download Go 1.2.2 and install it to /usr/local/go

#RUN curl -s https://go.dev/dl/go1.21.5.linux-amd64.tar.gz| tar -v -C /usr/local -xz

RUN apt-get install -y wget
#RUN wget -P https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
RUN wget -P /tmp "https://dl.google.com/go/go1.21.5.linux-amd64.tar.gz"

RUN rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go1.21.5.linux-amd64.tar.gz
RUN rm "/tmp/go1.21.5.linux-amd64.tar.gz"

# Let's people find our Go binaries
ENV PATH $PATH:/usr/local/go/bin

RUN go build
CMD ["./wfem"]

#RUN cargo build
