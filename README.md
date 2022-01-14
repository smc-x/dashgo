dashgo
======

A basic wrapper for serial-port communication with Dashgo D1.

## Quick Start

If you happen to have a Logitech F710 gamepad, the quick start guide should work without
modifications. Otherwise, you would need to craft your own YAML files and mount them to the
containers.

**Create a Container Network:**

```bash
docker network create nats-broker
```

**Start an NATS Broker:**

```bash
docker run -d \
  --name nats-broker \
  --network nats-broker \
  --restart always \
  -p 4222:4222 \
  nats:2.6.6
```
**Start a Container to Listen to Input Events:**
```bash
docker run -d \
  --name events \
  --network nats-broker \
  --privileged \
  --restart always \
  ghcr.io/smc-x/dashgo:events
```

**Start a Container to Control Dashgo D1:**

```bash
docker run -d \
  --name basic-gamepad \
  --network nats-broker \
  --privileged \
  --restart always \
  ghcr.io/smc-x/dashgo:basic-gamepad
```
