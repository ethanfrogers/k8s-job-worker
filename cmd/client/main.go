package main

import (
	"context"
	"fmt"

	"github.com/ethanfrogers/k8s-job-worker/pkg/kubernetes"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	opts := client.StartWorkflowOptions{
		TaskQueue: "kubernetes_worker",
	}

	execution, err := c.ExecuteWorkflow(ctx, opts, kubernetes.RunJobWorkflow)
	if err != nil {
		panic(err)
	}

	var status kubernetes.JobExecutionStatus
	if err := execution.Get(ctx, &status); err != nil {
		panic(err)
	}
	fmt.Println(status.Status)
}
