package main

import (
	"fmt"

	"github.com/chrisbradleydev/go-k8s/pkg/k8s"
)

func main() {
	pods, err := k8s.GetPods()
	if err != nil {
		panic(err.Error())
	}

	if len(pods.Items) > 0 {
		for i, pod := range pods.Items {
			i++
			fmt.Printf("%03d: %s\n", i, pod.Name)
		}

		fmt.Printf("There are %d pods in the cluster.\n", len(pods.Items))
	}
}