package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupDuplicates(t *testing.T) {
	app := NewApp("")
	sig := NewSqlInstanceGroupList(context.TODO(), app)

	sig.AddGroup(sig.NewGroup(sqlInstances[0].Name))
	sig.AddGroup(sig.NewGroup(sqlInstances[0].Name))
	sig.AddGroup(sig.NewGroup(sqlInstances[1].Name))
	sig.AddGroup(sig.NewGroup(sqlInstances[1].Name))
	sig.AddGroup(sig.NewGroup(sqlInstances[2].Name))
	sig.AddGroup(sig.NewGroup(sqlInstances[2].Name))

	groupsLength := len(sig.Groups)
	expectedLength := 3

	assert.Equal(t, groupsLength, expectedLength, "groups length should match expected length")
}

func TestDatabaseDuplicates(t *testing.T) {
	app := NewApp("")
	sig := NewSqlInstanceGroupList(context.TODO(), app)

	group := sig.NewGroup(sqlInstances[0].Name)
	sig.AddGroup(group)

	group.AddDatabase(sqlDatabases[0])
	group.AddDatabase(sqlDatabases[0])
	group.AddDatabase(sqlDatabases[0])

	databasesLength := len(group.Databases)
	expectedLength := 1

	assert.Equal(t, databasesLength, expectedLength, "databases length should match expected length")
}

func TestUserDuplicates(t *testing.T) {
	app := NewApp("")
	sig := NewSqlInstanceGroupList(context.TODO(), app)

	group := sig.NewGroup(sqlInstances[0].Name)
	sig.AddGroup(group)

	group.AddUser(sqlUsers[0])
	group.AddUser(sqlUsers[0])
	group.AddUser(sqlUsers[0])

	usersLength := len(group.Users)
	expectedLength := 1

	assert.Equal(t, usersLength, expectedLength, "users length should match expected length")
}