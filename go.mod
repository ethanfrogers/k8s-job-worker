module github.com/ethanfrogers/k8s-job-worker

go 1.15

require (
	go.temporal.io/sdk v1.10.0
	go.uber.org/zap v1.14.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	logur.dev/adapter/zap v0.5.0
	logur.dev/logur v0.17.0
)
