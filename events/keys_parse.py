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


def parse(obj, pick=None):
    """parse parses and validates the keys object and returns the last lookup table."""
    def _assert(obj, name, typ):
        assert name in obj
        v = obj[name]
        assert isinstance(v, typ)
        return v

    # There must be a "models" list enumerating all supported models
    models = { mod: None for mod in _assert(obj, "models", list) }
    for mod in list(models.keys()):
        # There must be a dedicated dict for each supported model
        model = _assert(obj, mod, dict)
        for key in model:
            # ... and each key in the model maps to a list of feasible values
            li = _assert(model, key, list)
            # ... w/o duplicates
            assert len(li) == len(set(li))
        models[mod] = model

    if pick is None:
        # There must be a "devices" list enumerating all supported devices
        devices = _assert(obj, "devices", list)
    else:
        # ... or one can selectively pick
        devices = [ pick ]

    lookup = None
    for device in devices:
        lookup = {}
        print("parsing {}...".format(device))
        # There must be a dedicated dict for each supported device 
        dev = _assert(obj, device, dict)
        # ... which specifies a supported model
        assert "model" in dev
        mod = dev["model"]
        assert mod in models
        model = models[mod]
        # ... and the other keys define binding rules
        for key, rules in dev.items():
            if key == "model":
                continue

            assert "code" in rules
            code = rules["code"]

            assert "bind" in rules
            bind = rules["bind"]
            assert bind in model
            values = model[bind]

            assert "vbind" in rules
            vbind = _assert(rules, "vbind", list)
            assert len(vbind) == len(set(vbind))
            assert len(vbind) == len(values)

            for vb, v in zip(vbind, values):
                lookup[(code, vb)] = (bind, v)

    return lookup


if __name__ == "__main__":
    obj = load("./keys.yaml")
    lookup = parse(obj)
    print("example lookup:\n")
    for k, v in lookup.items():
        print("  ({:>{wc}}, {:>{wc}}) -> ({:>{wc}}, {:>{wc}})".format(
            str(k[0]), str(k[1]), str(v[0]), str(v[1]), wc=6))
