# ⚡ JMS Core X

### Network Discovery • Diagnostics • Monitoring Core

**JMS Core X** is a modular network discovery and diagnostics engine written in Go.

It is designed as a foundation for discovering network devices, collecting runtime information, performing connectivity diagnostics, processing network state changes, and building higher-level monitoring and management systems.

> **Discover. Diagnose. Understand the Network.**

---

## 🚀 Overview

JMS Core X is built around a modular architecture where individual subsystems can operate independently while being coordinated through a central runtime orchestrator.

The project currently includes components for:

* 🌐 Network Discovery
* 📡 Device Detection
* 🩺 ICMP / Auto-Ping Diagnostics
* 🔄 Runtime Orchestration
* 📊 Unified Network Snapshots
* 🕓 History Tracking
* ⚡ Event Processing
* 🚨 Alert Infrastructure
* 🔌 API Integration
* 🔁 WebSocket Support
* 🗄️ Database / State Management
* 🔍 Raw Network Snapshot Processing

---

## 🧠 Architecture

```text
                 ┌─────────────────────┐
                 │     JMS Core X      │
                 └──────────┬──────────┘
                            │
                    Runtime Orchestrator
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
      Discovery        Diagnostics        Unified State
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │
                     Network Snapshot
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
       Events            History           Alerts
          │
          ├──────── API
          └──────── WebSocket
```

The architecture is intentionally modular so individual services can evolve without tightly coupling the entire system.

---

## 🩺 Diagnostics

The current runtime includes the **Diagnostics Layer** with Auto-Ping support.

The runtime orchestrator periodically executes connectivity diagnostics and provides the base for future device-health monitoring.

```text
Device Discovered
       ↓
Diagnostics Job
       ↓
ICMP Probe
       ↓
Reachability Result
       ↓
Runtime / Network State
```

---

## 📁 Project Structure

```text
jms-core-x/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── alerts/
│   ├── api/
│   ├── bootstrap/
│   ├── change/
│   ├── core/
│   ├── db/
│   ├── discovery/
│   ├── events/
│   ├── history/
│   ├── rawsnapshot/
│   ├── runtime/
│   ├── unified/
│   └── ws/
│
├── go.mod
├── go.sum
└── README.md
```

---

## 🛠️ Technology

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go\&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-Networking-FCC624?logo=linux\&logoColor=black)
![Status](https://img.shields.io/badge/Status-Development-orange)

Core technologies include:

* **Go 1.21**
* `gopacket`
* `gorilla/mux`
* `gorilla/websocket`
* `golang.org/x/net`
* `golang.org/x/sys`

---

## ⚙️ Build

Clone the repository:

```bash
git clone https://github.com/mando2022/jms-core-x.git
cd jms-core-x
```

Download dependencies:

```bash
go mod download
```

Build JMS Core X:

```bash
go build -o jms-core-x ./cmd/server
```

Run:

```bash
./jms-core-x
```

Or run directly:

```bash
go run ./cmd/server
```

---

## 🖥️ Runtime

A normal startup follows the core lifecycle:

```text
Bootstrap
    ↓
Core Init
    ↓
Core Start
    ↓
Diagnostics Layer
    ↓
Runtime Orchestrator
    ↓
JMS Core X Running
```

Shutdown is handled gracefully through `SIGINT` / `SIGTERM`.

```bash
CTRL+C
```

causes the runtime orchestrator and core layers to shut down cleanly.

---

## 🔐 Permissions

Some network discovery and ICMP functionality may require additional operating-system privileges depending on the platform and network configuration.

On Linux, raw network operations may require elevated capabilities or privileges.

---

## 🧩 Design Goals

JMS Core X is being developed around several principles:

**Modularity**
Each network subsystem should remain independently maintainable.

**Runtime Safety**
Services should start, operate, and shut down predictably.

**Extensibility**
New discovery methods, diagnostics, APIs, and monitoring systems can be added without redesigning the core.

**Network Visibility**
Transform low-level network information into a unified representation of devices and connectivity.

**Evidence First**
Network state should be determined from observable runtime information rather than assumptions.

---

## 🔭 Project Direction

JMS Core X is intended to evolve into a broader network intelligence and management core capable of combining:

```text
Discovery
   +
Diagnostics
   +
Device State
   +
Events
   +
History
   +
Monitoring
   +
Management
```

into one unified system.

---

## 👨‍💻 Author

### Abo Jasser

Embedded Systems • Firmware • Linux • Networking • Reverse Engineering

**J-M-S / JMS OS**

```text
>_ DEV • BUILD • REVERSE • EXPLORE
```

---

### ⚡ JMS Core X

**Built to discover what is really happening on the network.**
