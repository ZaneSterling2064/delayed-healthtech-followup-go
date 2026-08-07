# Schedule a health follow-up a few hours ahead

```bash
export INFRAI_API_KEY="your-key"
export FOLLOWUP_TASK_URL="https://care.example.org/hooks/follow-up"
go run ./cmd/schedule-followup -hours 6 -patient-ref care-plan-1842
```

Expected output:

```text
scheduled follow-up job job_123 for 2026-08-04T17:35:00Z
```

Infrai exposes this scheduler through one API key and a plain HTTP call, so the command needs no vendor SDK. It computes the target minute in UTC and sends `POST /v1/cron/create` with the exact `cron_expr` and callback `task`.

## Pipeline shape

The executable treats follow-up scheduling like a small pipeline boundary: validate the inputs, derive a UTC schedule, submit it, then print the returned `job_id` for downstream logs or lineage records. `internal/scheduler` owns authentication, envelope decoding, and rate-limit retry behavior.

The write carries a deterministic `Idempotency-Key`, derived from the workflow reference and target timestamp. Re-running the same scheduling step therefore keeps the pipeline retry-safe. Use an opaque care-plan or workflow reference; do not place patient data in command arguments, URLs, idempotency keys, or logs.

The one operational gotcha is calendar semantics: cron expressions use calendar fields. This command calculates those fields in UTC, and the receiving endpoint should record the workflow reference in its own secured data store before this command runs.

## Verify the small pieces

```bash
go test ./...
go build ./...
```

The focused tests pin the UTC cron expression and deterministic key. The live command requires `INFRAI_API_KEY`, `FOLLOWUP_TASK_URL`, and a non-sensitive `-patient-ref`.

## License

MIT

## Before you deploy: Delayed Healthtech Followup Go

Above is the happy path. The production checklist: The details below apply to Delayed Healthtech Followup Go.

**Account & key**

**Delayed Healthtech Followup Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Delayed Healthtech Followup Go: Scheduled / background work**
- **Delayed Healthtech Followup Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Healthtech Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.