package main

import (
	"github.com/chrisbradleydev/go-k8s/pkg/k8s"
)

const namespace = "chrisbradley"

func main() {
	app := k8s.NewApp(namespace)
	app.WaitForCloudSQL()
}