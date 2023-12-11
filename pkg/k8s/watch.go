package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
)

func (s *SqlInstanceGroupList) Watch() {
	for _, group := range s.Groups {
		s.wg.Add(1)
		go group.Watch(s.events, s.wg)
	}
	s.wg.Wait()
}

func (s *SqlInstanceGroup) Watch(eventsChan chan<- SqlInstanceGroupEvent, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(s.ctx, cloudSQLWaitTimeout)
	defer cancel()

	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceInstance,
		Name:  s.Name,
	}

	err := s.CheckInstance(ctx)
	if err != nil {
		baseEvent.Error = err
		eventsChan <- baseEvent
		return
	}

	if ok := s.WatchInstance(ctx, eventsChan); !ok {
		eventsChan <- SqlInstanceGroupEvent{}
		return
	}

	for _, db := range s.Databases {
		s.wg.Add(1)
		go s.WatchDatabase(eventsChan, db)
	}

	for _, user := range s.Users {
		s.wg.Add(1)
		go s.WatchUser(eventsChan, user)
	}

	s.wg.Wait()
}

func (s *SqlInstanceGroup) CheckInstance(ctx context.Context) *AppError {
	_, err := s.app.cnrmClient.SqlV1beta1().
		SQLInstances(s.app.namespace).
		Get(ctx, s.Name, v1.GetOptions{})
	if err != nil {
		return &AppError{Name:"CheckInstance", Message: fmt.Sprint(err)}
	}
	return nil
}

func (s *SqlInstanceGroup) CheckDatabase(eventsChan chan<- SqlInstanceGroupEvent, name string) (bool, error) {
	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceDatabase,
		Name:  name,
	}

	database, err := s.app.cnrmClient.SqlV1beta1().
		SQLDatabases(s.app.namespace).
		Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		baseEvent.Error = &AppError{Name:"CheckDatabase", Message: fmt.Sprint(err)}
		eventsChan <- baseEvent
		return false, err
	}
	return database.Status.Conditions[0].Reason != "UpdateFailed", nil
}

func (s *SqlInstanceGroup) CheckUser(eventsChan chan<- SqlInstanceGroupEvent, name string) (bool, error) {
	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceUser,
		Name:  name,
	}

	database, err := s.app.cnrmClient.SqlV1beta1().
		SQLUsers(s.app.namespace).
		Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		baseEvent.Error = &AppError{Name:"CheckUser", Message: fmt.Sprint(err)}
		eventsChan <- baseEvent
		return false, err
	}
	return database.Status.Conditions[0].Reason != "UpdateFailed", nil
}

func (s *SqlInstanceGroup) WatchInstance(ctx context.Context, eventsChan chan<- SqlInstanceGroupEvent) bool {
	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceInstance,
		Name:  s.Name,
	}

	rw, err := toolsWatch.NewRetryWatcher(
		"1",
		&cache.ListWatch{WatchFunc: s.getInstanceWatchFunc(s.Name)})
	defer rw.Stop()
	if err != nil {
		eventsChan <- baseEvent
		return false
	}

	if err := s.watchEvents(rw, eventsChan, baseEvent); err != nil {
		baseEvent.Error = &AppError{Name:"WatchInstance", Message: fmt.Sprint(err)}
		eventsChan <- baseEvent
		return false
	}

	return true
}

func (s *SqlInstanceGroup) WatchDatabase(eventsChan chan<- SqlInstanceGroupEvent, db *SqlDatabase) {
	defer s.wg.Done()

	// config connector resources may initialize with UpdateFailed
	// wait until these resources are updated to continue
	ticker := time.NewTicker(resourceCheckInterval)
	databaseOk := false
	for {
		select {
		case <-ticker.C:
			ok, err := s.CheckDatabase(eventsChan, db.Name)
			if err != nil {
				fmt.Printf("%s CheckDatabase err\n", db.Name)
				return
			}
			if ok {
				fmt.Printf("%s CheckDatabase ok\n", db.Name)
				databaseOk = true
			}
		case <-s.ctx.Done():
			fmt.Printf("%s CheckDatabase timeout\n", db.Name)
			return
		}
		if databaseOk {
			break
		}
	}

	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceDatabase,
		Name:  db.Name,
	}

	rw, err := toolsWatch.NewRetryWatcher(
		"1",
		&cache.ListWatch{WatchFunc: s.getDatabaseWatchFunc(db.Name)})
	defer rw.Stop()
	if err != nil {
		baseEvent.Error = &AppError{Name:"WatchDatabase", Message: fmt.Sprint(err)}
		eventsChan <- baseEvent
		return
	}

	if err := s.watchEvents(rw, eventsChan, baseEvent); err != nil {
		eventsChan <- baseEvent
	}
}

func (s *SqlInstanceGroup) WatchUser(eventsChan chan<- SqlInstanceGroupEvent, user *SqlUser) {
	defer s.wg.Done()

	// config connector resources may initialize with UpdateFailed
	// wait until these resources are updated to continue
	ticker := time.NewTicker(resourceCheckInterval)
	userOk := false
	for {
		select {
		case <-ticker.C:
			ok, err := s.CheckUser(eventsChan, user.Name)
			if err != nil {
				fmt.Printf("%s CheckUser err\n", user.Name)
				return
			}
			if ok {
				fmt.Printf("%s CheckUser ok\n", user.Name)
				userOk = true
			}
		case <-s.ctx.Done():
			fmt.Printf("%s CheckUser timeout\n", user.Name)
			return
		}
		if userOk {
			break
		}
	}

	baseEvent := SqlInstanceGroupEvent{
		Group: s,
		Type:  SqlResourceUser,
		Name:  user.Name,
	}

	rw, err := toolsWatch.NewRetryWatcher(
		"1",
		&cache.ListWatch{WatchFunc: s.getUserWatchFunc(user.Name)})
	defer rw.Stop()
	if err != nil {
		baseEvent.Error = &AppError{Name:"WatchUser", Message: fmt.Sprint(err)}
		eventsChan <- baseEvent
		return
	}

	if err := s.watchEvents(rw, eventsChan, baseEvent); err != nil {
		eventsChan <- baseEvent
	}
}

func (s *SqlInstanceGroup) getInstanceWatchFunc(name string) func(_ v1.ListOptions) (watch.Interface, error) {
	return func(_ v1.ListOptions) (watch.Interface, error) {
		var watch watch.Interface

		watch, err := s.app.cnrmClient.SqlV1beta1().
			SQLInstances(s.app.namespace).
			Watch(s.ctx, v1.ListOptions{FieldSelector: "metadata.name=" + name})

		if err != nil {
			return watch, err
		}
		return watch, nil
	}
}

func (s *SqlInstanceGroup) getDatabaseWatchFunc(name string) func(_ v1.ListOptions) (watch.Interface, error) {
	return func(_ v1.ListOptions) (watch.Interface, error) {
		var watch watch.Interface

		watch, err := s.app.cnrmClient.SqlV1beta1().
			SQLDatabases(s.app.namespace).
			Watch(s.ctx, v1.ListOptions{FieldSelector: "metadata.name=" + name})

		if err != nil {
			return watch, err
		}
		return watch, nil
	}
}

func (s *SqlInstanceGroup) getUserWatchFunc(name string) func(_ v1.ListOptions) (watch.Interface, error) {
	return func(_ v1.ListOptions) (watch.Interface, error) {
		var watch watch.Interface

		watch, err := s.app.cnrmClient.SqlV1beta1().
			SQLUsers(s.app.namespace).
			Watch(s.ctx, v1.ListOptions{FieldSelector: "metadata.name=" + name})

		if err != nil {
			return watch, err
		}
		return watch, nil
	}
}