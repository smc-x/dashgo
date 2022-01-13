FROM python:3.8-buster AS build
COPY ./gamepad/requirements.txt .
RUN pip install --user -r requirements.txt

FROM python:3.8-slim-buster
COPY --from=build /root/.local /root/.local
COPY ./events/*.py /root/
COPY ./events/*.yaml /root/
COPY ./config/*.yaml /root/

WORKDIR /root
ENTRYPOINT [ "python", "./main.py" ]
