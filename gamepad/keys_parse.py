import os

from ruamel.yaml import YAML
yaml = YAML(typ="safe")


def load(paths):
    """load loads a YAML data object from a list of path candidates (paths are tried one-by-one in
    order until one exists)."""
    if not isinstance(paths, list):
        paths = [paths]
    for path in paths:
        if not os.path.exists(path):
            continue
        print("loading from {}...".format(path))
        with open(path) as fp:
            obj = yaml.load(fp)
            return obj


def validate_keys(obj):
    """validate_keys validates the keys object."""
    def _assert(obj, name, typ):
        assert name in obj
        v = obj[name]
        assert isinstance(v, typ)
        return v

    # Widths for formatting
    wk = 12
    wv = 8

    # There must be a "devices" list enumerating all supported devices
    devices = _assert(obj, "devices", list)

    for device in devices:
        print("\nvalidating {}...".format(device))
        # There must be a dedicated dict for each supported device 
        dev = _assert(obj, device, dict)
        # ... which includes codes and values 
        codes = _assert(dev, "codes", dict)
        values = _assert(dev, "values", dict)
        for name, code in codes.items():
            # ... and for each code, there must be a dict of feasible values
            cvs = _assert(values, name, dict)
            print()
            print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))
            print("  | {:{wk}} | {:{wv}} |".format(name, code, wk=wk, wv=wv))
            print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))
            for k, v in cvs.items():
                print("  | {:{wk}} | {:{wv}} |".format(k, v, wk=wk, wv=wv))
            print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))


if __name__ == "__main__":
    obj = load("./keys.yaml")
    validate_keys(obj)
