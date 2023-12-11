package k8s

import (
	"errors"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/k8s/v1alpha1"
	sqlv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/sql/v1beta1"
	"k8s.io/apimachinery/pkg/watch"
)

func (s *SqlInstanceGroup) watchEvents(
	watcher watch.Interface,
	eventsChan chan<- SqlInstanceGroupEvent,
	baseEvent SqlInstanceGroupEvent,
) error {
	for {
		select {
		case e := <-watcher.ResultChan():
			event, healthy := s.processEvent(baseEvent, e)
			eventsChan <- event
			if healthy {
				resourcesCount--
				fmt.Printf("%s %s complete\n", event.Type, event.Name)
				fmt.Printf("remaining resources: %d\n", resourcesCount)
				return nil
			}
		case <-s.ctx.Done():
			return errors.New("resource watch timed out on context")
		}
	}
}

func (s *SqlInstanceGroup) processEvent(baseEvent SqlInstanceGroupEvent, event watch.Event) (SqlInstanceGroupEvent, bool) {
	var sc *v1alpha1.Condition

	switch baseEvent.Type {
	case SqlResourceInstance:
		instance, ok := event.Object.(*sqlv1beta1.SQLInstance)
		if ok && len(instance.Status.Conditions) > 0 {
			sc = &instance.Status.Conditions[0]
		}
	case SqlResourceDatabase:
		database, ok := event.Object.(*sqlv1beta1.SQLDatabase)
		if ok && len(database.Status.Conditions) > 0 {
			sc = &database.Status.Conditions[0]
		}
	case SqlResourceUser:
		user, ok := event.Object.(*sqlv1beta1.SQLUser)
		if ok && len(user.Status.Conditions) > 0 {
			sc = &user.Status.Conditions[0]
		}
	}
	baseEvent.Condition = sc

	return baseEvent, baseEvent.Condition != nil && baseEvent.Condition.Reason == "UpToDate"
}