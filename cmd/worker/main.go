package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethanfrogers/k8s-job-worker/pkg/kubernetes"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

func main() {
	workflowTaskQueue := flag.String("workflow-task-queue", "kubernetes_worker", "")
	temporalHost := flag.String("temporal-host", "localhost:7233", "")
	kubeconfig := flag.String("kubeconfig", "", "")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	opts := client.Options{
		HostPort: *temporalHost,
		Logger:   logur.LoggerToKV(zapadapter.New(logger)),
	}

	c, err := client.NewClient(opts)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	if *kubeconfig == "" {
		pth := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		kubeconfig = &pth
		logger.Sugar().Infof("kubeconfig is empty, using default %s", *kubeconfig)
	}

	kw, err := kubernetes.NewActivities(kubernetes.ClientFromKubeConfig(*kubeconfig))
	if err != nil {
		panic(err)
	}

	w := worker.New(c, *workflowTaskQueue, worker.Options{})
	w.RegisterWorkflow(kubernetes.RunJobWorkflow)
	w.RegisterActivity(kw.StartJobActivity)
	w.RegisterActivity(kw.WatchJobActivity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		fmt.Printf("worker failed to start: %s", err.Error())
		os.Exit(1)
	}
}
