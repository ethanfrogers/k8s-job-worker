package kubernetes

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	batchAPI "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

func WatchJob(ctx context.Context, client batchv1.JobInterface, jobName string, interval time.Duration) (ExecutionStatus, error) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return StatusFailed, ctx.Err()
		case <-ticker.C:
			activity.RecordHeartbeat(ctx, 0)
			job, err := client.Get(ctx, jobName, v1.GetOptions{})
			if err != nil {
				return StatusFailed, fmt.Errorf("failed to get job %s: %w", jobName, err)
			}

			if _, ok := conditionByType(batchAPI.JobFailed, job.Status.Conditions); ok {
				return StatusFailed, nil
			}

			if _, ok := conditionByType(batchAPI.JobComplete, job.Status.Conditions); ok {
				return StatusSucceeded, nil
			}

		}
	}
}

func conditionByType(conditionType batchAPI.JobConditionType, conditions []batchAPI.JobCondition) (batchAPI.JobCondition, bool) {
	for _, c := range conditions {
		if c.Type == conditionType {
			return c, true
		}
	}

	return batchAPI.JobCondition{}, false
}
