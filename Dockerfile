# Build Geth in a stock Go builder container
FROM golang:1.22-bookworm 

ENV PATH=$PATH:/root/.cargo/bin
RUN apt-get update && apt-get -y install curl gcc xvfb libvulkan1 bash
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y  && cargo install twgpu-tools

ENV export DISPLAY=:0
RUN Xvfb :0 -screen 0 1024x768x16 &

WORKDIR /app

COPY . .

RUN go build -o /bin/app

CMD ["/bin/app"]
