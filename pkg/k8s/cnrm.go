package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"

	v1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/sql/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (app *Application) GetInstance(ctx context.Context, name string) (*v1beta1.SQLInstance, error) {
	var err error
	var instance *v1beta1.SQLInstance
	instance, err = app.cnrmClient.SqlV1beta1().
		SQLInstances(app.namespace).
		Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return instance, err
	}
	return instance, nil
}

func (app *Application) GetInstanceList(ctx context.Context) (*v1beta1.SQLInstanceList, error) {
	var err error
	var list *v1beta1.SQLInstanceList
	list, err = app.cnrmClient.SqlV1beta1().
		SQLInstances(app.namespace).
		List(ctx, v1.ListOptions{})
	if err != nil {
		return list, err
	}
	return list, nil
}

func (app *Application) WatchInstance(ctx context.Context, name string) {
	watcher, err := app.cnrmClient.SqlV1beta1().
		SQLInstances(app.namespace).
		Watch(ctx, v1.ListOptions{FieldSelector: "metadata.name=" + name})
	defer watcher.Stop()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case e := <-watcher.ResultChan():
			fmt.Printf("%#v\n", e)
		case <-ctx.Done():
			err := errors.New("context timeout")
			log.Fatal(err)
		}
	}
}