# MemoryOS Server - HTTP API Service

<p align="center">
  <strong>RESTful API server for MemoryOS - Deploy memory as a service</strong>
</p>

HTTP API server providing MemoryOS functionality as a service. Part of the MemoryOS ecosystem.

## Features

- **RESTful API**: Full CRUD operations for memories
- **Multiple Endpoints**: /memory, /context, /agent, /team, /shared, /skill
- **Health Checks**: System health monitoring
- **Easy Integration**: Any HTTP client can use it

## Installation

```bash
go install github.com/yksanjo/memoryos-server@latest
```

## Quick Start

```bash
# Run server
memoryos-server

# Server starts on :8080
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/memory` | POST | Store memory |
| `/memory` | GET | Get memory |
| `/memory` | DELETE | Delete memory |
| `/memory/search` | GET | Search memories |
| `/context` | GET | Get compressed context |
| `/agent` | POST | Register agent |
| `/team` | POST | Create team |
| `/shared` | POST | Create shared value |
| `/skill` | POST | Register skill |
| `/stats` | GET | Get statistics |

## Example Usage

```bash
# Store a memory
curl -X POST http://localhost:8080/memory \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"agent1","type":"episodic","content":"User question"}'

# Get context
curl "http://localhost:8080/context?agent_id=agent1&max_tokens=2000"

# Search
curl "http://localhost:8080/memory/search?agent_id=agent1&q=pricing"
```

## Related Projects

- [memoryos](https://github.com/yksanjo/memoryos) - Full framework
- [memoryos-engine](https://github.com/yksanjo/memoryos-engine) - Core library
- [memoryos-cli](https://github.com/yksanjo/memoryos-cli) - CLI tool

## License

MIT
