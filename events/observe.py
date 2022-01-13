import asyncio
import signal
import sys

import nats


async def run(nats_url, subject):
    async def error_cb(e):
        print("Error:", e)

    async def closed_cb():
        # Wait for tasks to stop otherwise get a warning.
        await asyncio.sleep(0.2)
        loop.stop()

    async def reconnected_cb():
        print(f"Connected to NATS at {nc.connected_url.netloc}...")

    async def subscribe_handler(msg):
        subject = msg.subject
        data = msg.data.decode()
        print(
            "Got '{subject}': {data}".format(
                subject=subject, data=data,
            )
        )

    options = {
        "error_cb": error_cb,
        "closed_cb": closed_cb,
        "reconnected_cb": reconnected_cb,
        "servers": nats_url,
    }
    nc = await nats.connect(**options)

    def signal_handler():
        if nc.is_closed:
            return
        asyncio.create_task(nc.drain())

    for sig in ("SIGINT", "SIGTERM"):
        asyncio.get_running_loop().add_signal_handler(getattr(signal, sig), signal_handler)

    await nc.subscribe(subject, cb=subscribe_handler)


if __name__ == "__main__":
    args = sys.argv
    assert len(args) >= 3

    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(args[1], args[2]))
    try:
        loop.run_forever()
    finally:
        loop.close()
