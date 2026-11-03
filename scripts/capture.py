"""Capture the running app's window through the MyGo automation socket.

usage: capture.py SOCKET OUT.png [--eval JS]... [--wait MS]

Each --eval runs in the page in order (statements may `await` and `return`);
--wait sleeps after them, for animations to settle, before the capture.
"""
import base64, json, socket, sys, time

args = sys.argv[1:]
sock_path, out = args[0], args[1]
evals, wait = [], 600
i = 2
while i < len(args):
    if args[i] == "--eval":
        evals.append(args[i + 1]); i += 2
    elif args[i] == "--wait":
        wait = int(args[i + 1]); i += 2
    else:
        sys.exit(f"unknown argument {args[i]}")

s = socket.socket(socket.AF_UNIX)
s.connect(sock_path)
f = s.makefile("rw")
seq = 0

def call(method, params=None):
    global seq
    seq += 1
    f.write(json.dumps({"id": seq, "method": method, "params": params or {}}) + "\n")
    f.flush()
    r = json.loads(f.readline())
    if "error" in r:
        sys.exit(f"{method}: {r['error']}")
    return r["result"]

for _ in range(60):  # the socket listens before the window opens
    if call("windows")["windows"]:
        break
    time.sleep(0.5)
call("idle", {"timeout": 20000})
for js in evals:
    print("eval →", call("eval", {"script": js}).get("value"))
time.sleep(wait / 1000)
r = call("capture")
open(out, "wb").write(base64.b64decode(r["png"]))
print(f"wrote {out} ({r['width']}×{r['height']})")
