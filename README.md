# Layer 4 Load Balancer Prototype

A lightweight, concurrency-safe Layer 4 (TCP) Load Balancer written in Go. It distributes incoming TCP traffic across multiple backend servers using a **Round-Robin** algorithm, continuously monitors backend health via periodic TCP probes, and dynamically loads runtime configurations from a `config.yaml` file.

## ✨ Features

- **Layer 4 (TCP) Proxying:** High-speed bi-directional data streams streaming via `io.Copy`.
- **Dynamic YAML Configuration:** Load listening ports, backend targets, and health check intervals directly from `config.yaml`.
- **Round-Robin Load Balancing:** Atomic operations (`sync/atomic`) ensure high-performance concurrency-safe traffic distribution.
- **Automated Health Monitoring:** Periodic background probes using `net.DialTimeout` to dynamically remove unhealthy nodes and restore recovered backends.
- **Thread-Safe Architecture:** Uses `sync.RWMutex` to protect runtime backend status mutation without race conditions.

---

## 🛠️ Configuration (`config.yaml`)

Define your load balancer listening address and target backend servers in `config.yaml`:

```yaml
server:
  port: "127.0.0.1:8080"
  health_check_interval: "5s"

backends:
  - "127.0.0.1:8081"
  - "127.0.0.1:8082"
  - "127.0.0.1:8083"
```


## 🚦 Getting Started

### 1. Prerequisites
- **Go:** `1.20+` ইনস্টলড থাকতে হবে ([Install Go](https://go.dev/doc/install))
- **Python 3:** টেস্ট ব্যাকএন্ড সার্ভার রান করার জন্য

---

### 2. Installation & Setup



```bash
git clone https://github.com/Sakin790/L4-Load-Balancer-.git

cd L4-Load-Balancer-

go get gopkg.in/yaml.v3

add server IP in config.yaml file

go run main.go
```

c
Open 3 separate terminal tabs or windows and run a light HTTP server in each:

* **Terminal 1 (Backend 1):**
  ```bash
  python3 -m http.server 8081

* **Terminal 2 (Backend 2):**
  ```bash
  python3 -m http.server 8082

* **Terminal 3 (Backend 3):**
  ```bash
  python3 -m http.server 8083

### Step 4: Run the L4 Load Balancer
* **Terminal 4 (Load balancer):**
  ```bash
  go run main.go

🚀 L4 Load Balancer running on 127.0.0.1:8080 with Health Checking...

### Step 4: Test Traffic Routing (Round-Robin)

* **Hit the endpoint:**
  ```bash
  curl http://127.0.0.1:8080
  curl http://127.0.0.1:8080
  curl http://127.0.0.1:8080

Output should be visible