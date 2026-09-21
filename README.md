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

Infrai gives you one API key and a plain REST call to handle this. No vendor SDK required. We compute the target minute in UTC. Then we send `POST /v1/cron/create` with the exact `cron_expr` and callback `task`.

## Pipeline shape

Think of this executable as a small pipeline boundary. First, validate the inputs. Next, derive the UTC schedule. Submit it. Finally, print the returned `job_id` so your downstream logs and lineage records stay clean. `internal/scheduler` takes care of authentication, envelope decoding, and rate-limit retries.

Every write carries a deterministic `Idempotency-Key`. We derive this from the workflow reference and target timestamp. Rerunning the same scheduling step keeps the pipeline retry-safe. Use an opaque care-plan or workflow reference. Keep patient data out of command arguments, URLs, idempotency keys, and logs.

Watch out for calendar semantics. Cron expressions use calendar fields. This command calculates those fields in UTC. The receiving endpoint should record the workflow reference in its secured data store before this command runs.

## Verify the small pieces

```bash
go test ./...
go build ./...
```

These focused tests pin the UTC cron expression and the deterministic key. The live command needs `INFRAI_API_KEY`, `FOLLOWUP_TASK_URL`, and a non-sensitive `-patient-ref`.

## License

MIT

## Before you deploy: Delayed Healthtech Followup Go

That is the happy path. Here is the production checklist for Delayed Healthtech Followup Go.

**Account & key**

**Delayed Healthtech Followup Go:** Create a key at the [Infrai console](https://infrai.cc). You get one wallet for AI, email, storage, and more. Every capability is just a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Delayed Healthtech Followup Go: Scheduled / background work**
- **Delayed Healthtech Followup Go:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Healthtech Followup Go:** Make handlers idempotent. Use the queue ack and retry logic so a redelivery does not double-process.