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


def parse(obj, pick=None, print_out=True):
    """parse parses and validates the keys object and returns the last lookup table."""
    def _assert(obj, name, typ):
        assert name in obj
        v = obj[name]
        assert isinstance(v, typ)
        return v

    def _print(*args):
        if print_out:
            print(*args)

    # Widths for formatting
    wk = 12
    wv = 8

    if pick is None:
        # There must be a "devices" list enumerating all supported devices
        devices = _assert(obj, "devices", list)
    else:
        devices = [ pick ]

    lookup = None
    for device in devices:
        _print("\nparsing {}...".format(device))
        # There must be a dedicated dict for each supported device 
        dev = _assert(obj, device, dict)
        # ... which includes codes and values 
        codes = _assert(dev, "codes", dict)
        values = _assert(dev, "values", dict)
        lookup = {}
        for name, code in codes.items():
            lookup[code] = set()
            # ... and for each code, there must be a dict of feasible values
            cvs = _assert(values, name, dict)
            _print()
            _print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))
            _print("  | {:{wk}} | {:{wv}} |".format(name, code, wk=wk, wv=wv))
            _print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))
            for k, v in cvs.items():
                lookup[code].add(v)
                _print("  | {:{wk}} | {:{wv}} |".format(k, v, wk=wk, wv=wv))
            _print("  +-{:{wk}}-+-{:{wv}}-+".format("-"*wk, "-"*wv, wk=wk, wv=wv))
            assert len(lookup[code]) == len(cvs)
        assert len(lookup) == len(codes)
    return lookup


if __name__ == "__main__":
    obj = load("./keys.yaml")
    lookup = parse(obj)
    print("\nexample lookup:\n")
    print(" ", lookup)
