# Futures Trading

A safety-first **Go-based Binance Futures trading bot** using a Maker / Checker / Executor architecture.

## Overview

```mermaid
graph TD
    market[Market]
    analyze[Analyze]
    signals[(Signals & FSM)]
    maker[Maker/<br/>Decision Engine]
    checker[Checker/<br/>Risk Engine]
    executor[Executor/<br/>Order Engine]
    notify[Notify]
    guarder[Guarder/<br/>Safety & Circuit Breaker]
    exchange[Exchange<br/>Binance]

    market --> analyze --> |Ingest| signals
    maker -->|Intent| checker
    checker -->|Approved| executor
    executor -->|Place| exchange

    maker -.-> signals
    guarder -.-> signals
    checker -.-> signals

    guarder --> |Terminate| executor
    exchange --> notify
```

## Singals Queue

### Intervals

```mermaid
flowchart LR
    A[Analyze Service] -->|Raw Signals| B[Signal Ingress Queue]

    B --> C[Signal Service]

    subgraph C[Signal Service]
        direction TB

        subgraph Q5m["Interval Queue (5m)"]
            H5m["Max Heap<br/>Priority by Score"]
        end

        subgraph Q15m["Interval Queue (15m)"]
            H15m["Max Heap<br/>Priority by Score"]
        end

        subgraph Q1h["Interval Queue (1h)"]
            H1h["Max Heap<br/>Priority by Score"]
        end
    end

    C -->|Peek / Pop Best Signal| D[Decision / Maker Service]
    D -->|Trading Intent| E[Checker / Risk Engine]
    E -->|Approved Order| F[Executor / Order Placement]
```

### Interval

```mermaid
flowchart TB
    subgraph Signals Queue
        H["Priority Heap (Max Heap)"]

        H --> S1["Signal: BTCUSDT<br/>Score: 78.2"]
        H --> S2["Signal: ETHUSDT<br/>Score: 74.5"]
        H --> S3["Signal: SOLUSDT<br/>Score: 69.1"]
    end

```

```mermaid
sequenceDiagram
    participant Analyzer
    participant SignalSvc
    participant Decision
    participant Risk
    participant Executor

    Analyzer->>SignalSvc: Ingest(signal, interval)
    SignalSvc->>SignalSvc: Build priority & push to heap

    Decision->>SignalSvc: Peek(interval)
    SignalSvc-->>Decision: Best signal

    Decision->>Risk: Trading intent
    Risk-->>Decision: Approved / Rejected

    Decision->>SignalSvc: Pop(interval)
    Decision->>Executor: Execute order
```
