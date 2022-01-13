import os
import time

path2failures = "./gamepad-failures"
num_failures = 0
start_time = time.time()


def prehook():
    """prehook checks the number of consecutive failures."""
    global num_failures
    if not os.path.exists(path2failures):
        return
    try:
        with open(path2failures) as fp:
            ts, num = fp.readline().split(",")
            ts = float(ts.strip())
            num = int(num.strip())

            if time.time() - ts < 3:
                num_failures = num

            if num_failures >= 3:
                time.sleep(3)
    except:
        pass


def posthook():
    """posthook updates the number of consecutive failures and tear down the process."""
    global num_failures
    if time.time() - start_time > 10:
        num_failures = 1
    else:
        num_failures += 1

    with open(path2failures, "w") as fp:
        fp.write("%f,%d\n" % (time.time(), num_failures))

    os._exit(1)
