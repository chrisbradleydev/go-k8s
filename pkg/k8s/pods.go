package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (app *Application) GetPod(podName string, getOptions metav1.GetOptions) (*corev1.Pod, error) {
	pod, err := app.kubeClient.
		CoreV1().
		Pods(app.namespace).
		Get(
			context.TODO(),
			podName,
			getOptions,
		)
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (app *Application) GetPods() (*corev1.PodList, error) {
	pods, err := app.kubeClient.
		CoreV1().
		Pods(app.namespace).
		List(
			context.TODO(),
			metav1.ListOptions{},
		)
	if err != nil {
		return nil, err
	}
	return pods, nil
}