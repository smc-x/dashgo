FROM python:3.8-buster AS build
COPY ./gamepad/requirements.txt .
RUN pip install --user -r requirements.txt

FROM python:3.8-slim-buster
COPY --from=build /root/.local /root/.local
COPY ./gamepad/publisher.py /root/publisher.py
COPY ./gamepad/gamepad.py /root/gamepad.py
COPY ./gamepad/keys.yaml /root/keys.yaml
COPY ./config/config.yaml /root/config.yaml

WORKDIR /root
ENTRYPOINT [ "python", "./gamepad.py" ]
