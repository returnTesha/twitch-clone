실시간 스트리밍 플랫폼 미니 클론

[![Go](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat&logo=go)](https://golang.org)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=flat&logo=redis)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 📌 프로젝트 개요

이 프로젝트는 [Twitch Engineering Blog](https://blog.twitch.tv/en/tags/engineering/)를 읽고 영감을 받아,
실제 Twitch가 풀었던 기술적 문제들을 직접 Go로 구현해보는 학습 및 포트폴리오 프로젝트입니다.

단순한 튜토리얼 따라하기가 아닌, **실제 대규모 시스템의 아키텍처 의사결정을 이해하고 미니 버전으로 재현**하는 것을 목표로 합니다.

---

## 🏗️ 구현 목표 서비스

### 1. 💬 Chat Service (구현 완료)
Twitch 채팅 시스템을 참고한 실시간 메시지 처리 서비스

- WebSocket 기반 실시간 양방향 통신
- 채널별 메시지 팬아웃 (1:N 브로드캐스트)
- Redis Stream 기반 메시지 브로커 (PEL로 메시지 유실 방지)
- Redis Sorted Set 기반 채널 유저 목록 / 인기 채널 랭킹
- SSE(Server-Sent Events) 기반 대기열 순번 실시간 알림
- 수평 확장 가능한 구조 (Horizontal Scaling)

**참고:** [Twitch - Breaking the Monolith (채팅 시스템 분리 과정)](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/)

---

### 2. 🎥 Ingest Routing Service (구현 예정)
Twitch의 Intelligest 아키텍처를 참고한 스트림 라우팅 서비스

- PoP(Point of Presence) → Origin 서버 라우팅 시뮬레이션
- 실시간 서버 부하 기반 동적 라우팅 결정
- Greedy 알고리즘 기반 리소스 최적화
- Origin 장애 시 자동 라우팅 전환 (Failover)
- chat-service와 gRPC 연동 (시청자 수 기반 라우팅)

**참고:** [Twitch - Ingesting Live Video Streams at Global Scale](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/)

---

## 🛠️ 기술 스택

| 분류 | 기술 |
|------|------|
| Language | Go 1.26.4 |
| 웹 프레임워크 | Gin |
| 메시지 브로커 | Redis Stream (XReadGroup + PEL) |
| 실시간 통신 | WebSocket (gorilla/websocket) |
| 실시간 푸시 | SSE (Server-Sent Events) |
| Redis 자료구조 | Sorted Set (유저 목록, 채널 랭킹, 대기열 순번) |
| 인증 | JWT (golang-jwt/jwt v5) |
| 서비스 간 통신 | gRPC (예정) |
| 컨테이너 | Docker / Docker Compose |
| 인프라스트럭처 | Kubernetes |
| 모니터링 | Prometheus + Grafana |

---

## 📁 프로젝트 구조
twitch-clone/

├── chat-service/

│   ├── cmd/

│   │   └── main.go              # 서버 진입점, 라우트, graceful shutdown

│   ├── internal/

│   │   ├── auth/                # JWT 발급/검증, Gin 미들웨어

│   │   ├── config/              # 환경변수 기반 설정

│   │   ├── handler/             # WebSocket 핸들러 (readPump/writePump)

│   │   ├── hub/                 # 채널별 클라이언트 관리 + 팬아웃 + Sorted Set

│   │   ├── model/               # Message, Client 구조체

│   │   └── stream/              # Redis Stream producer/consumer

│   └── Dockerfile

├── ingest-service/              # 스트림 라우팅 서비스 (구현 예정)

│   ├── cmd/

│   └── internal/

│       ├── routing/             # IRS 라우팅 로직

│       └── pop/                 # PoP 시뮬레이터

├── common/                      # 공통 proto, config, util

├── docker-compose.yml

└── README.md

---

## 🌐 API

| Method | Path | 설명 |
|--------|------|------|
| GET | `/health` | 헬스체크 |
| POST | `/token` | JWT 발급 (개발용) |
| GET | `/ws/:channelId` | WebSocket 연결 (JWT 필요) |
| GET | `/channels/:channelId/users` | 채널 활성 유저 목록 (입장순) |
| GET | `/channels/ranking` | 인기 채널 TOP 10 (시청자순) |
| GET | `/channels/:channelId/rank/:userId` | 채널 내 입장 순번 |
| GET | `/queue/:channelId/watch/:userId` | 대기열 순번 SSE 스트림 |

---

## 🚀 실행 방법

```bash
# 레포 클론
git clone https://github.com/returntesha/twitch-clone.git
cd twitch-clone

# Redis 실행
docker run -d --name redis-chat -p 6379:6379 redis:7.0

# chat-service 실행
cd chat-service
go run cmd/main.go

# 토큰 발급
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1", "username": "returntesha"}'

# WebSocket 연결
wscat -c "ws://localhost:8080/ws/channel1?token={발급받은토큰}"
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
- Redis Stream을 활용한 내구성 있는 메시지 큐 (PEL, XReadGroup)
- Redis Sorted Set을 활용한 실시간 랭킹 및 순번 시스템
- 수평 확장 가능한 WebSocket 서버 설계
- SSE와 WebSocket의 적절한 선택 기준
- 마이크로서비스 간 gRPC 통신
- 실제 트래픽 패턴을 고려한 라우팅 알고리즘

---

✂️ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

---

# 🎮 twitch-clone

> A mini clone of Twitch's real-time streaming platform, inspired by the Twitch Engineering Blog and implemented in Go.

[![Go](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat&logo=go)](https://golang.org)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=flat&logo=redis)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## 📌 Overview

This project is a learning and portfolio project inspired by the [Twitch Engineering Blog](https://blog.twitch.tv/en/tags/engineering/).

The goal is not to follow tutorials, but to **understand the architectural decisions behind large-scale systems and reproduce them in a mini version using Go**.

---

## 🏗️ Services

### 1. 💬 Chat Service (Implemented)
A real-time messaging service inspired by Twitch's chat system.

- WebSocket-based real-time bidirectional communication
- Per-channel message fan-out (1:N broadcast)
- Redis Stream as message broker (PEL for message durability)
- Redis Sorted Set for channel user list / popular channel ranking
- SSE-based real-time queue position notification
- Horizontally scalable architecture

**Reference:** [Twitch - Breaking the Monolith (chat system extraction)](https://blog.twitch.tv/en/2022/03/30/breaking-the-monolith-at-twitch/)

---

### 2. 🎥 Ingest Routing Service (Planned)
A stream routing service inspired by Twitch's Intelligest architecture.

- PoP (Point of Presence) → Origin server routing simulation
- Dynamic routing decisions based on real-time server load
- Resource optimization using a Greedy algorithm
- Automatic failover on origin server failure
- gRPC integration with chat-service (viewer-count-based routing)

**Reference:** [Twitch - Ingesting Live Video Streams at Global Scale](https://blog.twitch.tv/en/2022/04/26/ingesting-live-video-streams-at-global-scale/)

---

## 🛠️ Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.26.4 |
| Web Framework | Gin |
| Message Broker | Redis Stream (XReadGroup + PEL) |
| Real-time Communication | WebSocket (gorilla/websocket) |
| Real-time Push | SSE (Server-Sent Events) |
| Redis Data Structure | Sorted Set (user list, channel ranking, queue position) |
| Authentication | JWT (golang-jwt/jwt v5) |
| Inter-service Communication | gRPC (planned) |
| Container | Docker / Docker Compose |
| Infrastructure | Kubernetes |
| Monitoring | Prometheus + Grafana |

---

## 📁 Project Structure
twitch-clone/

├── chat-service/

│   ├── cmd/

│   │   └── main.go              # Entry point, routes, graceful shutdown

│   ├── internal/

│   │   ├── auth/                # JWT generation/validation, Gin middleware

│   │   ├── config/              # Environment-based configuration

│   │   ├── handler/             # WebSocket handler (readPump/writePump)

│   │   ├── hub/                 # Per-channel client management + fan-out + Sorted Set

│   │   ├── model/               # Message, Client structs

│   │   └── stream/              # Redis Stream producer/consumer

│   └── Dockerfile

├── ingest-service/              # Stream routing service (planned)

│   ├── cmd/

│   └── internal/

│       ├── routing/             # IRS routing logic

│       └── pop/                 # PoP simulator

├── common/                      # Shared proto, config, utils

├── docker-compose.yml

└── README.md

---

## 🌐 API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/token` | Issue JWT (dev only) |
| GET | `/ws/:channelId` | WebSocket connection (JWT required) |
| GET | `/channels/:channelId/users` | Active user list (by join order) |
| GET | `/channels/ranking` | Top 10 channels (by viewer count) |
| GET | `/channels/:channelId/rank/:userId` | User join rank in channel |
| GET | `/queue/:channelId/watch/:userId` | Queue position SSE stream |

---

## 🚀 Getting Started

```bash
# Clone the repo
git clone https://github.com/returntesha/twitch-clone.git
cd twitch-clone

# Start Redis
docker run -d --name redis-chat -p 6379:6379 redis:7.0

# Run chat-service
cd chat-service
go run cmd/main.go

# Issue a token
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1", "username": "returntesha"}'

# Connect via WebSocket
wscat -c "ws://localhost:8080/ws/channel1?token={your_token}"
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
- Durable message queuing with Redis Streams (PEL, XReadGroup)
- Real-time ranking and queue positioning with Redis Sorted Set
- Horizontally scalable WebSocket server design
- Choosing between SSE and WebSocket based on use case
- gRPC communication between microservices
- Routing algorithms that account for real-world traffic patterns

---

*Inspired by Twitch Engineering — built to learn, not to compete.*