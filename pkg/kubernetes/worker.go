package kubernetes

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

type ExecutionStatus string

var (
	StatusPending   = ExecutionStatus("PENDING")
	StatusFailed    = ExecutionStatus("FAILED")
	StatusSucceeded = ExecutionStatus("SUCCEEDED")
)

type JobExecutionStatus struct {
	Status ExecutionStatus
}

func RunJobWorkflow(ctx workflow.Context) (JobExecutionStatus, error) {
	opts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, opts)

	var (
		createdJobName string
		activities     *Activities
	)

	if err := workflow.ExecuteActivity(ctx, activities.StartJobActivity).Get(ctx, &createdJobName); err != nil {
		return JobExecutionStatus{Status: StatusFailed}, fmt.Errorf("could not create job: %w", err)
	}

	var result ExecutionStatus
	waitOpts := workflow.ActivityOptions{ScheduleToCloseTimeout: 30 * time.Minute}
	ctx = workflow.WithActivityOptions(ctx, waitOpts)
	if err := workflow.ExecuteActivity(ctx, activities.WatchJobActivity, createdJobName).Get(ctx, &result); err != nil {
		return JobExecutionStatus{Status: StatusFailed}, err
	}

	return JobExecutionStatus{Status: result}, nil
}

type Activities struct {
	client *kube.Clientset
}

func NewActivities(cf ClientFactory) (*Activities, error) {
	client, err := cf()
	if err != nil {
		return nil, err
	}
	return &Activities{
		client: client,
	}, nil
}

func (w *Activities) StartJobActivity(ctx context.Context) (string, error) {

	j := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "temporal-",
			Namespace:    "default",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: pointer.Int32(90),
			BackoffLimit:            pointer.Int32(0),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:    "perl",
							Image:   "perl",
							Command: []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"},
						},
					},
				},
			},
		},
	}

	created, err := w.client.BatchV1().Jobs("default").Create(context.Background(), &j, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	return created.GetName(), nil
}

func (w *Activities) WatchJobActivity(ctx context.Context, name string) (ExecutionStatus, error) {
	return WatchJob(ctx, w.client.BatchV1().Jobs("default"), name, 5*time.Second)
}

type ClientFactory func() (*kube.Clientset, error)

func ClientFromKubeConfig(pth string) ClientFactory {
	return func() (*kube.Clientset, error) {
		config, err := clientcmd.BuildConfigFromFlags("", pth)
		if err != nil {
			return nil, fmt.Errorf("could not create kube config: %w", err)
		}

		return kube.NewForConfig(config)
	}
}
