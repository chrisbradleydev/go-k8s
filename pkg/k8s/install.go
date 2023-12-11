package k8s

import (
	"context"
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/k8s/v1alpha1"
)

func (app *Application) watchCloudSql(
	events <-chan SqlInstanceGroupEvent,
	done <-chan interface{}) {
	for {
		select {
		case e := <-events:
			if e.Condition != nil {
				fmt.Fprintln(os.Stdout,
					fmtResourceStatus(e.Type.String(),
						e.Name,
						e.Condition))
			}
			if e.Error != nil {
				app.errors = append(app.errors, e.Error)
			}
		case <-done:
			fmt.Println("watchCloudSql done")
			return
		}
	}
}

func (app *Application) WaitForCloudSQL() {
	done := make(chan interface{})
	defer close(done)

	ctx, cancel := context.WithTimeout(context.Background(), cloudSQLWaitTimeout)
	defer cancel()

	sqlInstanceGroups := NewSqlInstanceGroupList(ctx, app)
	sqlInstanceGroups.InitGroups()

	// fmt.Fprint(os.Stdout, sqlInstanceGroups.String())

	go app.watchCloudSql(sqlInstanceGroups.events, done)

	sqlInstanceGroups.Watch()
	done <- nil // exit watchCloudSql

	for _, e := range app.errors {
		fmt.Println(e.Name, e.Message)
	}
}

func fmtResourceStatus(key, value string, con *v1alpha1.Condition) string {
	return fmt.Sprintf("[%s%s%s] %s%s%s %s%s%s",
		ColorBlue,
		key,
		ColorNc,
		ColorWhite,
		value,
		ColorNc,
		ColorRed,
		con.Reason,
		ColorNc,
	)
}