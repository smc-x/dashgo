import asyncio
import os
import queue as lib_queue

import nats


class Publisher(object):
    """Publisher provides an NATS helper for publishing gamepad events."""

    def __init__(self, key="gamepad", alive_msg="ok", alive_gap=1, queue_cap=1024):
        """alive_gap defines the maximum idle waiting time before sending an alive_msg."""
        self.__key = key
        self.__alive_msg = alive_msg.encode()
        self.__alive_gap = alive_gap
        self.__msg_queue = lib_queue.Queue(maxsize=queue_cap)

    def publish(self, code, value):
        """publish puts a code-value pair to publish."""
        self.__msg_queue.put("{}:{}".format(code, value))

    async def app(self, nats_url):
        """app includes the main logics of the publisher."""
        nc = await nats.connect(nats_url)
        while True:
            try:
                msg = self.__msg_queue.get(timeout=self.__alive_gap)
                to_send = msg.encode()
            except lib_queue.Empty:
                to_send = self.__alive_msg

            try:
                await nc.publish(self.__key, to_send)
                await nc.flush()
            except:
                # Tear down the process directly to trigger external restarting policies
                os._exit(1)

    def run(self, nats_url):
        """run starts running the publisher blockingly."""
        asyncio.run(self.app(nats_url))
