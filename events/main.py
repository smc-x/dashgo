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
publisher = Publisher(
    config["publisher_key"],
    config["alive_gap"],
    config["queue_cap"],
)
publish = publisher.publish


def capture_events(device, lookup):
    """capture_events captures and publishes target input events."""
    active = {}

    try:
        for event in device.read_loop():
            code = event.code
            value = event.value
            if code in lookup and value in lookup[code]:
                ts = int(time.time() * 1000) / 1000
                active[code] = (ts, value)
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
    # Find input target
    path = None
    target = config["input_target"]
    mapping = find_input_devices()
    for k, v in mapping.items():
        if v == target:
            path = k
            break
    if path is None:
        print("input target not found", flush=True)
        failures.posthook()
    print("input target found:", path, flush=True)

    # Get input model
    lookup = keys_parse.parse(keys, pick=config["input_model"], print_out=False)

    # Capture gamepad events
    device = evdev.InputDevice(path)
    t = threading.Thread(target=capture_events, args=(device, lookup))
    t.start()

    # Start publisher
    publisher.run(config["nats_url"])
