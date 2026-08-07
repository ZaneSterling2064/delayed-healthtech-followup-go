package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/delayed-healthtech-followup/internal/scheduler"
)

func main() {
	hours := flag.Int("hours", 6, "hours before the follow-up callback")
	taskURL := flag.String("task-url", os.Getenv("FOLLOWUP_TASK_URL"), "HTTPS endpoint that performs the follow-up")
	patientRef := flag.String("patient-ref", "", "non-sensitive workflow reference used for idempotency")
	flag.Parse()

	if *hours < 1 || *taskURL == "" || *patientRef == "" {
		log.Fatal("hours must be positive; task-url and patient-ref are required")
	}

	runAt := time.Now().UTC().Add(time.Duration(*hours) * time.Hour).Truncate(time.Minute)
	client, err := scheduler.NewClient(os.Getenv("INFRAI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.CronCreate(context.Background(), scheduler.CronCreateRequest{
		CronExpr: cronAt(runAt),
		Task:     *taskURL,
		MaxRuns:  1,
	}, idempotencyKey(*patientRef, runAt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("scheduled follow-up job %s for %s\n", result.JobID, runAt.Format(time.RFC3339))
}

func cronAt(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d *", t.Minute(), t.Hour(), t.Day(), int(t.Month()))
}

func idempotencyKey(patientRef string, runAt time.Time) string {
	sum := sha256.Sum256([]byte(patientRef + "|" + runAt.Format(time.RFC3339)))
	return "followup-" + hex.EncodeToString(sum[:16])
}
