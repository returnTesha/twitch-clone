# 🎮 twitch-clone

> Twitch 엔지니어링 블로그를 참고하여 Go로 구현한 실시간 스트리밍 플랫폼 미니 클론

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)](https://golang.org)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=flat&logo=redis)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 📌 프로젝트 개요

이 프로젝트는 [Twitch Engineering Blog](blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch)를 읽고 영감을 받아,  
실제 Twitch가 풀었던 기술적 문제들을 직접 Go로 구현해보는 학습 및 포트폴리오 프로젝트입니다.

단순한 튜토리얼 따라하기가 아닌, **실제 대규모 시스템의 아키텍처 의사결정을 이해하고 미니 버전으로 재현**하는 것을 목표로 합니다.

---

## 🏗️ 구현 목표 서비스

### 1. 💬 Chat Service (구현 중)
Twitch 채팅 시스템을 참고한 실시간 메시지 처리 서비스

- WebSocket 기반 실시간 양방향 통신
- 채널별 메시지 팬아웃 (1:N 브로드캐스트)
- Redis Pub/Sub or Redis Stream 기반 메시지 브로커
- 다중 서버 환경에서의 메시지 동기화
- 수평 확장 가능한 구조 (Horizontal Scaling)

**참고:** [Twitch - Breaking the Monolith (채팅 시스템 분리 과정)](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/)

---

### 2. 🎥 Ingest Routing Service (구현 예정)
Twitch의 Intelligest 아키텍처를 참고한 스트림 라우팅 서비스

- PoP(Point of Presence) → Origin 서버 라우팅 시뮬레이션
- 실시간 서버 부하 기반 동적 라우팅 결정
- Greedy 알고리즘 기반 리소스 최적화
- Origin 장애 시 자동 라우팅 전환 (Failover)

**참고:** [Twitch - Ingesting Live Video Streams at Global Scale](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/)

---

## 🛠️ 기술 스택

| 분류 | 기술 |
|------|------|
| Language | Go 1.26.4 |
| 메시지 브로커 | Redis Stream / Redis Pub/Sub |
| 실시간 통신 | WebSocket (gorilla/websocket) |
| 서비스 간 통신 | gRPC |
| 컨테이너 | Docker / Docker Compose |
| 모니터링 | Prometheus + Grafana |
| 인스라스트력쳐 | kubernetes |

---

## 📁 프로젝트 구조

```
twitch-clone/
├── chat-service/         # 채팅 서비스 (WebSocket + 메시지 팬아웃)
│   ├── cmd/
│   ├── internal/
│   │   ├── handler/      # WebSocket 핸들러
│   │   ├── hub/          # 채널 관리 및 팬아웃
│   │   └── stream/       # Redis Stream 연동
│   └── Dockerfile
├── ingest-service/       # 스트림 라우팅 서비스 (구현 예정)
│   ├── cmd/
│   ├── internal/
│   │   ├── routing/      # IRS 라우팅 로직
│   │   └── pop/          # PoP 시뮬레이터
│   └── Dockerfile
├── common/               # 공통 proto, config, util
├── docker-compose.yml
└── README.md
```

---

## 🚀 실행 방법

```bash
# 레포 클론
git clone https://github.com/returntesha/twitch-clone.git
cd twitch-clone

# 전체 서비스 실행
docker-compose up -d

# 채팅 서비스만 실행
cd chat-service
go run cmd/main.go
```

---

## 📚 참고한 Twitch 엔지니어링 블로그

| 글 제목 | 링크 |
|---------|------|
| Twitch Engineering: An Introduction and Overview | [🔗](https://blog.twitch.tv/en/2015/12/18/twitch-engineering-an-introduction-and-overview-a23917b71a25/) |
| Breaking the Monolith at Twitch Part 1 | [🔗](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/) |
| Breaking the Monolith at Twitch Part 2 | [🔗](https://blog.twitch.tv/en/2022/04/12/breaking-the-monolith-at-twitch-part-2/) |
| Ingesting Live Video Streams at Global Scale | [🔗](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/) |
| Twitch State of Engineering 2023 | [🔗](https://blog.twitch.tv/en/2023/09/28/twitch-state-of-engineering-2023) |

---

## ✍️ 구현하면서 배우는 것들

- 대규모 실시간 메시지 시스템 설계
- Go 고루틴과 채널을 활용한 동시성 제어
- Redis Stream을 활용한 내구성 있는 메시지 큐
- 수평 확장 가능한 WebSocket 서버 설계
- 마이크로서비스 간 gRPC 통신
- 실제 트래픽 패턴을 고려한 라우팅 알고리즘

---

✂️ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

---

# 🎮 twitch-clone

> A mini clone of Twitch's real-time streaming platform, inspired by the Twitch Engineering Blog and implemented in Go.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)](https://golang.org)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=flat&logo=redis)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 📌 Overview

This project is a learning and portfolio project inspired by the [Twitch Engineering Blog](https://blog.twitch.tv/en/tags/engineering/).

The goal is not to follow tutorials, but to **understand the architectural decisions behind large-scale systems and reproduce them in a mini version using Go**.

---

## 🏗️ Services

### 1. 💬 Chat Service (In Progress)
A real-time messaging service inspired by Twitch's chat system.

- WebSocket-based real-time bidirectional communication
- Per-channel message fan-out (1:N broadcast)
- Redis Stream / Pub/Sub as message broker
- Message synchronization across multiple server instances
- Horizontally scalable architecture

**Reference:** [Twitch - Breaking the Monolith (chat system extraction)](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/)

---

### 2. 🎥 Ingest Routing Service (Planned)
A stream routing service inspired by Twitch's Intelligest architecture.

- PoP (Point of Presence) → Origin server routing simulation
- Dynamic routing decisions based on real-time server load
- Resource optimization using a Greedy algorithm
- Automatic failover on origin server failure

**Reference:** [Twitch - Ingesting Live Video Streams at Global Scale](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/)

---

## 🛠️ Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.22 |
| Message Broker | Redis Stream / Redis Pub/Sub |
| Real-time Communication | WebSocket (gorilla/websocket) |
| Inter-service Communication | gRPC |
| Container | Docker / Docker Compose |
| Monitoring | Prometheus + Grafana |
| Infrastructure | kubernetes |

---

## 📁 Project Structure

```
twitch-clone/
├── chat-service/         # Chat service (WebSocket + message fan-out)
│   ├── cmd/
│   ├── internal/
│   │   ├── handler/      # WebSocket handler
│   │   ├── hub/          # Channel management & fan-out
│   │   └── stream/       # Redis Stream integration
│   └── Dockerfile
├── ingest-service/       # Stream routing service (planned)
│   ├── cmd/
│   ├── internal/
│   │   ├── routing/      # IRS routing logic
│   │   └── pop/          # PoP simulator
│   └── Dockerfile
├── common/               # Shared proto, config, utils
├── docker-compose.yml
└── README.md
```

---

## 🚀 Getting Started

```bash
# Clone the repo
git clone https://github.com/returntesha/twitch-clone.git
cd twitch-clone

# Run all services
docker-compose up -d

# Run chat service only
cd chat-service
go run cmd/main.go
```

---

## 📚 References

| Article | Link |
|---------|------|
| Twitch Engineering: An Introduction and Overview | [🔗](https://blog.twitch.tv/en/2015/12/18/twitch-engineering-an-introduction-and-overview-a23917b71a25/) |
| Breaking the Monolith at Twitch Part 1 | [🔗](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/) |
| Breaking the Monolith at Twitch Part 2 | [🔗](https://blog.twitch.tv/en/2022/04/12/breaking-the-monolith-at-twitch-part-2/) |
| Ingesting Live Video Streams at Global Scale | [🔗](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/) |
| Twitch State of Engineering 2023 | [🔗](https://blog.twitch.tv/en/2023/09/28/twitch-state-of-engineering-2023) |

---

## ✍️ What I'm Learning

- Designing large-scale real-time messaging systems
- Concurrency control with Go goroutines and channels
- Durable message queuing with Redis Streams
- Horizontally scalable WebSocket server design
- gRPC communication between microservices
- Routing algorithms that account for real-world traffic patterns

---

*Inspired by Twitch Engineering — built to learn, not to compete.*
