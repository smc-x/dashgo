import asyncio
import queue as lib_queue
import time

import nats

import failures


class Publisher(object):
    """Publisher provides an NATS helper for publishing input events."""

    def __init__(self, key, alive_gap=0.1, queue_cap=1024):
        """alive_gap defines the maximum idle waiting time before sending an cached msg."""
        self.__key = key
        self.__alive_gap = alive_gap
        self.__msg_queue = lib_queue.Queue(maxsize=queue_cap)
        self.__cached = None

    def publish(self, msg):
        """publish puts a msg to publish."""
        self.__msg_queue.put(msg)

    async def app(self, nats_url):
        """app includes the main logics of the publisher."""
        try:
            nc = await nats.connect(nats_url)
            while True:
                try:
                    self.__cached = self.__msg_queue.get(timeout=self.__alive_gap)
                except lib_queue.Empty:
                    pass
                to_send = self.__cached

                if to_send is not None:
                    to_send = "{\"ts\":%d,\"pl\":%s}" % (int(1000 * time.time()), to_send)
                    await nc.publish(self.__key, to_send.encode())
                    await nc.flush()
        except:
            # Tear down the process directly to trigger external restarting policies
            failures.posthook()

    def run(self, nats_url):
        """run starts running the publisher blockingly."""
        loop = asyncio.get_event_loop()
        try:
            loop.run_until_complete(self.app(nats_url))
        finally:
            loop.close()
