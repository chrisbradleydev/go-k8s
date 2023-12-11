package k8s

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (app *Application) GetDeployment(ctx context.Context, name string) (*v1.Deployment, error) {
	var err error
	var deploy *v1.Deployment
	deploy, err = app.kubeClient.
		AppsV1().
		Deployments(app.namespace).
		Get(ctx,
			name,
			metav1.GetOptions{})
	if err != nil {
		return deploy, err
	}
	return deploy, nil
}