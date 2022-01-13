import json
import threading
import time

import evdev


import keys_parse
keys = keys_parse.load([
    "./keys.yaml",
])
config = keys_parse.load([
    "./config.yaml",
    "../config/config.yaml",
])


import failures
failures.prehook()


from publisher import Publisher


def capture_events(device, lookup, publish):
    """capture_events captures and publishes target input events."""
    active = {}

    try:
        for event in device.read_loop():
            vi = (event.code, event.value)
            if vi in lookup:
                ts = int(time.time() * 1000) / 1000
                vo = lookup[vi]
                active[vo[0]] = (ts, vo[1])
                publish(json.dumps(active))
    except:
        # Tear down the process directly to trigger external restarting policies
        failures.posthook()


def find_input_devices():
    """find_input_devices builds a mapping from device path to id_vendor:id_product."""
    devices = [ evdev.InputDevice(path) for path in evdev.list_devices() ]
    mapping = {}
    for device in devices:
        id_vendor = device.info.vendor
        id_product = device.info.product
        if id_vendor > 0 and id_product > 0:
            mapping[device.path] = "%04x:%04x" % (id_vendor, id_product)
    return mapping


if __name__ == "__main__":
    # Find events target
    path, dev = None, None
    targets = config["events_targets"]
    mapping = find_input_devices()
    for id_, dev_ in targets:
        for k, v in mapping.items():
            if v == id_:
                path = k
                dev = dev_
                break
        if path is not None:
            break
    if path is None:
        print("events target not found", flush=True)
        failures.posthook()
    print("events target found: ({}, {})".format(path, dev), flush=True)

    # Get events model
    lookup = keys_parse.parse(keys, pick=dev)

    # Start publisher
    publisher = Publisher(
        keys[dev]["model"],
        config["alive_gap"],
        config["queue_cap"],
    )

    # Capture gamepad events
    device = evdev.InputDevice(path)
    t = threading.Thread(target=capture_events, args=(device, lookup, publisher.publish))
    t.start()

    publisher.run(config["nats_url"])
