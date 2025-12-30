# Distributed Auction System

A small distributed auction system with multiple nodes, a centralized lock, and Lamport clocks.
Nodes accept bids locally and replicate state asynchronously to reach a consistent final result.

---

## Requirements

* Docker
* Docker Compose

---

## Start the System

```bash
docker compose up --build
```

This starts:

* 1 coordinator (gRPC, port 7000)
* 3 auction nodes:

  * node-A: REST on `localhost:8080`
  * node-B: REST on `localhost:8081`
  * node-C: REST on `localhost:8082`

Wait until all services are running.

---

## Basic Interaction

### Submit a bid

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "bidderId": "alice"}'
```

Example response:

```json
{
  "applied": true,
  "state": {
    "highestBid": 100,
    "highestBidder": "alice",
    "lamport": 1
  }
}
```

`applied=true` means the bid was applied locally on this node.

---

### Check current state

```bash
curl http://localhost:8080/state
```

Example output:

```json
{
  "highestBid": 100,
  "highestBidder": "alice",
  "lamport": 1
}
```

---

## Common Test Scenarios

### Lower bid is rejected

```bash
curl -X POST http://localhost:8081/bid \
  -H "Content-Type: application/json" \
  -d '{"amount": 50, "bidderId": "bob"}'
```

Expected behavior:

* `applied=false`
* state remains unchanged

---

### Higher bid replaces current winner

```bash
curl -X POST http://localhost:8082/bid \
  -H "Content-Type: application/json" \
  -d '{"amount": 200, "bidderId": "carol"}'
```

Expected behavior:

* `applied=true`
* new highest bid is replicated to all nodes

---

### Concurrent bids on different nodes

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{"amount": 310, "bidderId": "alice"}' &

curl -X POST http://localhost:8081/bid \
  -H "Content-Type: application/json" \
  -d '{"amount": 300, "bidderId": "bob"}' &

wait
```

Both requests may return `applied=true`, since bids are evaluated locally.
To observe the final auction result, always query /state after replication.

---

### Verify convergence

After a short delay, all nodes converge to the same state:

```bash
curl http://localhost:8080/state
curl http://localhost:8081/state
curl http://localhost:8082/state
```

Expected output on all nodes:

```json
{
  "highestBid": 310,
  "highestBidder": "alice",
  "lamport": 5
}
```

The final state is deterministic due to Lamport timestamps and a tie-breaking rule.

---

## Run Unit Tests

```bash
go test ./...
```

This runs unit tests for:

* auction comparison logic
* Lamport clock behavior
* node-level bid handling
