FROM golang:1.17.6-buster AS build
COPY . /root/src
RUN cd /root/src/gamepad/test \
    && go build -o /root/gamepad-ctrl

FROM debian:buster-slim
COPY --from=build /root/gamepad-ctrl /root/
COPY ./config/*.yaml /root/
COPY ./gamepad/*.yaml /root/

WORKDIR /root
ENTRYPOINT [ "./gamepad-ctrl" ]
