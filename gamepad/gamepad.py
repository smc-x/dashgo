import os
import threading

import evdev
from ruamel.yaml import YAML
yaml = YAML(typ="safe")
with open("./keys.yaml") as fp:
    keys = yaml.load(fp)
with open("../config/config.yaml") as fp:
    config = yaml.load(fp)

from publisher import Publisher
publisher = Publisher(
    config["publisher_key"],
    config["alive_msg"],
    config["alive_gap"],
    config["queue_cap"],
)
publish = publisher.publish

ABS_X = keys["abs_x"]
ABS_X_VALUES = set(keys["abs_x_values"])

ABS_Y = keys["abs_y"]
ABS_Y_VALUES = set(keys["abs_y_values"])

KEY_CODES = set(keys["key_codes"])
KEY_VALUES = set(keys["key_values"])


def capture_events(device):
    """capture_events captures and publishes target gamepad events."""
    types = { evdev.ecodes.EV_KEY, evdev.ecodes.EV_ABS }

    try:
        for event in device.read_loop():
            if event.type not in types:
                continue

            if (
                event.code == ABS_X and event.value in ABS_X_VALUES or
                event.code == ABS_Y and event.value in ABS_Y_VALUES or
                event.code in KEY_CODES and event.value in KEY_VALUES
            ):
                publish(event.code, event.value)
    except:
        # Tear down the process directly to trigger external restarting policies
        os._exit(1)


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
    # Find gamepad device
    path = None
    target = config["gamepad_id"]
    mapping = find_input_devices()
    for k, v in mapping.items():
        if v == target:
            path = k
            break
    if path is None:
        print("gamepad not found")
        exit(1)
    print("gamepad found:", path)

    # Capture gamepad events
    device = evdev.InputDevice(path)
    t = threading.Thread(target=capture_events, args=(device,))
    t.start()

    # Start publisher
    publisher.run(config["nats_url"])
