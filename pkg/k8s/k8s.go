package k8s

import (
	"context"
	"path"

	"github.com/chrisbradleydev/go-k8s/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClient() kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(utils.GetHomeDir(), ".kube/config"))
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func GetPod(podName string, getOptions metav1.GetOptions) (*corev1.Pod, error) {
	pod, err := GetClient().CoreV1().Pods("").Get(
		context.TODO(),
		podName,
		getOptions,
	)
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func GetPods() (*corev1.PodList, error) {
	pods, err := GetClient().CoreV1().Pods("").List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, err
	}
	return pods, nil
}
