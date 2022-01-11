import os
import time

path2failures = "./gamepad-failures"
num_failures = 0
start_time = time.time()


def prehook():
    """prehook checks the number of consecutive failures."""
    if os.path.exists(path2failures):
        with open(path2failures) as fp:
            num_failures = int(fp.readline().strip())

    if num_failures >= 3:
        time.sleep(3)


def posthook():
    """posthook updates the number of consecutive failures and tear down the process."""
    global num_failures
    if time.time() - start_time > 10:
        num_failures = 1
    else:
        num_failures += 1

    with open(path2failures, "w") as fp:
        fp.write("%d\n" % (num_failures))

    os._exit(1)
