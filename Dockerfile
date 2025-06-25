FROM ubuntu:24.04 as goios
RUN apt-get update && apt-get -y install unzip wget curl git
RUN curl -O https://dl.google.com/go/go1.23.0.linux-amd64.tar.gz \
    && tar -xvf go1.23.0.linux-amd64.tar.gz \
    && mv go /usr/local
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH
WORKDIR /app
RUN git clone https://github.com/danielpaulus/go-ios.git .
RUN go build -o goios
RUN chmod +x goios


FROM ubuntu:24.04
RUN apt-get update && apt-get install -y git build-essential pkg-config autoconf automake \
    libtool-bin python3 python-is-python3 libssl-dev libusb-1.0-0-dev  \
     socat checkinstall curl  libcurl4-openssl-dev
WORKDIR /opt/src
RUN git clone https://github.com/libimobiledevice/libplist
RUN git clone https://github.com/libimobiledevice/libtatsu.git
RUN git clone https://github.com/libimobiledevice/libusbmuxd
RUN git clone https://github.com/libimobiledevice/libimobiledevice
RUN git clone https://github.com/libimobiledevice/libimobiledevice-glue
RUN git clone https://github.com/libimobiledevice/usbmuxd.git
RUN cd libplist && ./autogen.sh && make && make install
RUN cd libtatsu && ./autogen.sh && make && make install
RUN cd libimobiledevice-glue && ./autogen.sh --enable-debug && make && make install
RUN cd libusbmuxd && ./autogen.sh && make && make install
RUN cd libimobiledevice && ./autogen.sh --enable-debug && make && make install
RUN cd usbmuxd && ./autogen.sh && make && make install
RUN ldconfig
RUN rm -rf /var/lib/apt/lists/* /opt/src
WORKDIR /app
ENV USBMUXD_SOCKET_ADDRESS=/tmp/socket.sock
COPY run.sh run.sh
COPY --from=goios /app/goios /goios
RUN chmod +x run.sh
ENTRYPOINT ["./run.sh"]
