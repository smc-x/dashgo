import os
import sys

import evdev


if __name__ == "__main__":
    args = sys.argv
    if not len(args) == 2:
        print("[usage]: python keys_profile.py id_vendor:id_product")
        os._exit(1)

    # Get target device
    target = args[1].lower()
    dev = None
    for device in [ evdev.InputDevice(path) for path in evdev.list_devices() ]:
        id_vendor = device.info.vendor
        id_product = device.info.product
        if id_vendor > 0 and id_product > 0:
            cur = "%04x:%04x" % (id_vendor, id_product)
            cur = cur.lower()
            if cur == target:
                dev = device
                break

    if dev is None:
        print("target {} not found".format(target))
        os._exit(1)

    # Profile device
    print("press <ctrl-c> to stop profiling...")
    types = { evdev.ecodes.EV_KEY, evdev.ecodes.EV_ABS }
    for event in device.read_loop():
        if event.type in types:
            print(event)
