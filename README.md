# HEX Auth HTTPS Scanner

A Python-based HTTPS scanner that tests **15-character hexadecimal authentication tokens** via a query parameter and identifies **response patterns based on Content-Length and HTTP status codes**.

This tool is designed for **analysis, debugging, and security research** where response fingerprinting is required.

---

## ✨ Features

* Tests 15-character hexadecimal values (`0-9a-f`)
* Sends requests over **HTTPS**
* Appends tokens to `?auth=` query parameter
* Matches responses by:

  * HTTP status code (`200`)
  * `Content-Length` value
* Logs **all matching results** to a separate TXT file
* Resume-capable via configurable HEX ranges
* Optional request delay (rate-limit friendly)
* Robust error handling and logging

---

## 📌 Use Case Examples

* Response fingerprinting
* Authentication token behavior analysis
* Web application testing
* Debugging inconsistent server-side auth logic

---

## ⚠️ Important Notice

The full 15-character HEX keyspace contains:

```
16^15 ≈ 1.15 × 10^18 combinations
```

A full scan is **theoretical only** and not practically feasible. This tool is intended to be used with:

* Restricted ranges
* Known prefixes/suffixes
* Incremental or resumed scans

Use responsibly.

---

## 🛠 Requirements

* Python **3.9+**
* `requests` library

Install dependencies:

```bash
pip install requests
```

---

## ⚙️ Configuration

Edit the script configuration section:

* `BASE_URL` – Target HTTPS endpoint
* `TARGET_CONTENT_LENGTH` – Response length to match
* `START_HEX` / `END_HEX` – HEX range (resume support)
* `SLEEP_BETWEEN_REQUESTS` – Optional delay between requests

---

## ▶️ Usage

```bash
python auth_hex_https_scan_extended.py
```

The scanner will iterate through the defined HEX range and log all matches.

---

## 📄 Output Files

### Matching Results

`hits_content_length_22874.txt`

```
2026-01-16T13:42:11 auth=deadbeefcafebab status=200 content-length=22874
```

### Error Log

`errors.log`

```
000000000000abc ERROR ReadTimeout
```

---

## 🔁 Resume Scanning

To continue after interruption, simply update:

```python
START_HEX = "00000001a3f9000"
```

and restart the script.

---

## 🔒 Legal & Ethical Disclaimer

This project is provided for **educational and authorized testing purposes only**.

* Do **not** use against systems you do not own or have explicit permission to test
* The author assumes **no liability** for misuse
* Ensure compliance with applicable laws and policies

---

## 📜 License

MIT License

---

## 🤝 Contributions

Pull requests, improvements, and documentation updates are welcome.

Please ensure all contributions align with ethical and legal usage guidelines.
