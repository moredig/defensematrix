# 🛡️ Defense Matrix

> *"The best trap is one the attacker walks into willingly."*

A dynamic **Cyber Deception and Automated Moving Target Defense (AMTD)** system.
Built in Go and Python. Designed to exhaust attackers, burn zero-days, and leave production untouched.

---

## 🧠 What It Does

Instead of blocking traffic like a firewall, or passively logging like a honeypot —
Defense Matrix treats infrastructure as **entirely ephemeral**.

It intercepts incoming connections, routes threats into ultra-isolated **"Boats"** (micro-containers),
and automatically destroys and regenerates those environments the moment an attack is detected.

Attackers waste their tools. Your real system never gets touched.

---

## ⚙️ Architecture

    [ Internet ] → [ Wamai Gateway ] → [ Boat-1: Frontline ]
                                            ↓ (attack detected)
                                       [ Boat-2: Shadow ] ← promoted instantly
                                       [ Boat-1 ] → nuked & recycled

### Core Layers

| Layer | Tech | Role |
|---|---|---|
| Wamai Gateway | Go / net/http/httputil | Reverse proxy — intercepts all traffic |
| Fleet Manager | Go / Docker SDK | Manages the rolling three-boat rotation |
| Recycler | Go goroutines | Auto-triggers rotation on nuke events |
| Scanner | Go / regexp | Pattern-matches raw payloads for threats |
| Deception Engine | Go | Injects fake banners, breadcrumbs, lure files |
| CLI Dashboard | Python / Textual | Live terminal perimeter monitor |

---

## 🚢 The Rolling Trinity

| Boat | Role | State |
|---|---|---|
| Boat-1 | Frontline | 🔵 Absorbing live traffic, broadcasting bait |
| Boat-2 | Shadow | ⚫ Sterile clone, hot standby |
| Boat-3 | Graveyard | 🔴 Just nuked — recycling into new shadow |

The moment an attack signature triggers:
1. Traffic reroutes to Boat-2 in **under 10ms**
2. Boat-1 is force-killed and wiped from RAM and disk
3. Boat-3 reprovisioned as the new Shadow
4. Loop resets — attacker is back at square one

---

## 🎭 Deception Layer

Boats broadcast irresistible fake identities:

- **Outdated banners** — Apache 2.2, IIS 6.0, nginx 1.10, OpenSSH 7.2
- **Open ports** — SSH, MySQL, Postgres, RDP left deliberately open
- **Weak credentials** — admin/admin, root/toor accepted on purpose
- **Lure files** — fake credentials.txt, .env, partial SQL dumps dropped inside

---

## 🔍 Detection Signatures

| Signature | Threat Level |
|---|---|
| SQL Injection | 🔴 HIGH |
| Shell Injection | 🔴 HIGH |
| Directory Traversal | 🔴 HIGH |
| XSS Attempt | 🟠 MEDIUM |
| Brute Force | 🟠 MEDIUM |
| Port Scan | 🟡 LOW |
| Credential Stuffing | 🟡 LOW |

---

## 🖥️ CLI Dashboard

    ┌─────────────────────────────────────────────────┐
    │           DEFENSE MATRIX — LIVE PERIMETER        │
    ├──────────────┬──────────────┬───────────────────┤
    │   BOAT-1     │   BOAT-2     │   BOAT-3          │
    │ 🔵 IDLE      │ 🔵 IDLE      │ 🔴 NUKED          │
    │ FRONTLINE    │ SHADOW       │ GRAVEYARD         │
    ├─────────────────────────────────────────────────┤
    │ ⚡ Connections: 3  🎯 Hits: 12  🔄 Rotations: 2 │
    ├─────────────────────────────────────────────────┤
    │ [WAMAI] HIT — SQL Injection | Level: HIGH       │
    │ [WAMAI] Rotation triggered — nuking frontline   │
    │ [WAMAI] boat-2 promoted to frontline            │
    └─────────────────────────────────────────────────┘

---

## 🚀 Quick Start

### Prerequisites
- Docker + Docker Compose
- Go 1.21+
- Python 3.8+

### Run the Engine

    git clone https://github.com/moredig/defensematrix.git
    cd defensematrix
    go run ./cmd/wamai

### Run the Dashboard

    cd cli
    pip install -r ../requirements.txt
    python dashboard.py

### Deploy with Docker

    cd deploy
    docker-compose up --build

---

## 📁 Project Structure

    defensematrix/
    ├── cmd/wamai/          # Go entrypoint
    ├── internal/
    │   ├── docker/         # Docker daemon client + orchestrator
    │   ├── fleet/          # Boat state machine + rotation
    │   ├── detection/      # Traffic scanner + signatures
    │   ├── deception/      # Bait profiles + breadcrumbs
    │   └── proxy/          # Reverse proxy gateway + multiplexer
    ├── cli/                # Python terminal dashboard
    ├── config/             # YAML configuration
    ├── deploy/             # Dockerfiles + compose
    └── logs/               # Engine output

---

## ⚖️ License

Apache 2.0 — see [LICENSE](LICENSE)

> **Legal Notice:** This tool is designed for defensive research and infrastructure protection on systems you own or have explicit permission to defend. Deploying deception infrastructure against unauthorized targets may violate computer fraud laws in your jurisdiction.

