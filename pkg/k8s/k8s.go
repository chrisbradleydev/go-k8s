package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	cnrm "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Application struct {
	restConfig *rest.Config
	kubeClient *kubernetes.Clientset
	cnrmClient *cnrm.Clientset
	namespace  string
	errors     AppErrorsList
}

type AppError struct {
	Name     string
	Message  string
}

type AppErrorsList []*AppError

const (
	cloudSQLWaitTimeout          =  20 * time.Minute
	resourceCheckInterval        =  500 * time.Millisecond
)

const (
	ColorRed   = "\033[0;31m"
	ColorWhite = "\033[0;37m"
	ColorBlue  = "\033[1;34m"
	ColorNc    = "\x1b[0m"
)

func NewApp(namespace string) *Application {
	var err error
	var restConfig *rest.Config
	configPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	restConfig, err = clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		fmt.Println(err)
	}

	app := &Application{
		restConfig: restConfig,
		namespace:  namespace,
	}

	kubeClient, err := app.createClient()
	if err != nil {
		fmt.Println(err)
	}
	app.kubeClient = kubeClient

	cnrmClient, err := app.createCNRM()
	if err != nil {
		fmt.Println(err)
	}
	app.cnrmClient = cnrmClient

	return app
}

func (app *Application) createClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(app.restConfig)
}

func (app *Application) createCNRM() (*cnrm.Clientset, error) {
	return cnrm.NewForConfig(app.restConfig)
}