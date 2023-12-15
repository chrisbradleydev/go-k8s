package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/k8s/v1alpha1"
)

type SqlInstance struct {
	Name string `yaml:"name" json:"name"`
}

type SqlDatabase struct {
	Name         string `yaml:"name" json:"name"`
	InstanceName string `yaml:"instanceName" json:"instanceName"`
}

type SqlUser struct {
	Name         string `yaml:"name" json:"name"`
	InstanceName string `yaml:"instanceName" json:"instanceName"`
}

type SqlInstanceGroup struct {
	Name      string
	Instance  *SqlInstance
	Databases []*SqlDatabase
	Users     []*SqlUser
	wg        sync.WaitGroup
	ctx       context.Context
	app       *Application
}

type SqlInstanceGroupList struct {
	Groups []*SqlInstanceGroup
	ctx    context.Context
	events chan SqlInstanceGroupEvent
	wg     *sync.WaitGroup
	app    *Application
}

type SqlInstanceGroupEvent struct {
	Group     *SqlInstanceGroup
	Type      DependencyType
	Name      string
	Condition *v1alpha1.Condition
	Error     *AppError
}

type DependencyType string

var (
	SqlResourceInstance DependencyType = "SqlInstance"
	SqlResourceDatabase DependencyType = "SqlDatabase"
	SqlResourceUser     DependencyType = "SqlUser"

	sqlInstances = [3]SqlInstance{
		{Name: "test-deployments-mysql-uno"},
		{Name: "test-deployments-mysql-dos"},
		{Name: "test-deployments-mysql-tres"},
	}
	sqlDatabases = [3]SqlDatabase{
		{Name: "td-uno-db", InstanceName: "test-deployments-mysql-uno"},
		{Name: "td-dos-db", InstanceName: "test-deployments-mysql-dos"},
		{Name: "td-tres-db", InstanceName: "test-deployments-mysql-tres"},
	}
	sqlUsers     = [5]SqlUser{
		{Name: "td-uno-user", InstanceName: "test-deployments-mysql-uno"},
		{Name: "td-dos-user", InstanceName: "test-deployments-mysql-dos"},
		{Name: "td-tres-user-1", InstanceName: "test-deployments-mysql-tres"},
		{Name: "td-tres-user-2", InstanceName: "test-deployments-mysql-tres"},
		{Name: "td-tres-user-3", InstanceName: "test-deployments-mysql-tres"},
	}
)

func (t *DependencyType) String() string {
	return string(*t)
}

func NewSqlInstanceGroupList(ctx context.Context, app *Application) *SqlInstanceGroupList {
	return &SqlInstanceGroupList{
		Groups: make([]*SqlInstanceGroup, 0),
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		events: make(chan SqlInstanceGroupEvent),
		app:    app,
	}
}

func (s *SqlInstanceGroupList) NewGroup(name string) *SqlInstanceGroup {
	return &SqlInstanceGroup{
		Name:      name,
		Databases: make([]*SqlDatabase, 0),
		Users:     make([]*SqlUser, 0),
		ctx:       s.ctx,
		app:       s.app,
	}
}

func (s *SqlInstanceGroupList) AddGroup(g *SqlInstanceGroup) error {
	if s.HasGroup(g.Name) {
		return fmt.Errorf("can't add group to group list, already exists")
	}
	s.Groups = append(s.Groups, g)
	return nil
}

func (s *SqlInstanceGroupList) GetGroup(name string) *SqlInstanceGroup {
	var group *SqlInstanceGroup
	for _, s := range s.Groups {
		if s.Name == name {
			return s
		}
	}
	return group
}

func (s *SqlInstanceGroupList) HasGroup(name string) bool {
	for _, s := range s.Groups {
		if s.Name == name {
			return true
		}
	}
	return false
}

func (s *SqlInstanceGroupList) AddInstance(i SqlInstance) {
	if s.HasGroup(i.Name) {
		return
	}
	s.AddGroup(s.NewGroup(i.Name))
}

func (s *SqlInstanceGroupList) AddDatabase(d SqlDatabase) {
	var group *SqlInstanceGroup
	if group = s.GetGroup(d.InstanceName); group == nil {
		group = s.NewGroup(d.InstanceName)
	}
	group.AddDatabase(d)
}

func (s *SqlInstanceGroupList) AddUser(u SqlUser) {
	var group *SqlInstanceGroup
	if group = s.GetGroup(u.InstanceName); group == nil {
		group = s.NewGroup(u.InstanceName)
	}
	group.AddUser(u)
}

func (g *SqlInstanceGroup) AddDatabase(d SqlDatabase) {
	if !g.HasDatabase(d) {
		g.Databases = append(g.Databases, &d)
	}
}

func (g *SqlInstanceGroup) HasDatabase(d SqlDatabase) bool {
	for _, database := range g.Databases {
		if d.Name == database.Name {
			return true
		}
	}
	return false
}

func (g *SqlInstanceGroup) AddUser(u SqlUser) {
	if !g.HasUser(u) {
		g.Users = append(g.Users, &u)
	}
}

func (g *SqlInstanceGroup) HasUser(u SqlUser) bool {
	for _, user := range g.Users {
		if u.Name == user.Name {
			return true
		}
	}
	return false
}

func (s *SqlInstanceGroupList) String() string {
	str := strings.Builder{}
	for _, group := range s.Groups {
		str.WriteString("Instance: ")
		str.WriteString(group.Name)
		str.WriteString("\n")
		str.WriteString("  Databases:\n")
		for _, db := range group.Databases {
			str.WriteString("  - ")
			str.WriteString(db.Name)
			str.WriteRune('\n')
		}
		str.WriteString("  Users:\n")
		for _, user := range group.Users {
			str.WriteString("  - ")
			str.WriteString(user.Name)
			str.WriteRune('\n')
		}
	}
	return str.String()
}

func (s *SqlInstanceGroupList) InitGroups() {
	for _, instance := range sqlInstances {
		s.AddInstance(instance)
	}
	for _, database := range sqlDatabases {
		s.AddDatabase(database)
	}
	for _, user := range sqlUsers {
		s.AddUser(user)
	}
}