FROM golang:1.17.6-buster AS build
COPY . /root/src
RUN cd /root/src/control/basic-gamepad \
    && go build -o /root/main

FROM debian:buster-slim
COPY --from=build /root/main /root/
COPY ./config/*.yaml /root/

WORKDIR /root
ENTRYPOINT [ "./main" ]
