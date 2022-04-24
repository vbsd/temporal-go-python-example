package main

import (
	"flag"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	var host string
	var namespace string

	flag.StringVar(&host, "host", "temporal:7233", "Temporal host with a port")
	flag.StringVar(&namespace, "namespace", "default", "Temporal namespace")
	flag.Parse()

	c, err := client.NewClient(client.Options{
		HostPort:  host,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, TASK_QUEUE_WORKFLOWS, worker.Options{})

	w.RegisterWorkflowWithOptions(
		BookTrip,
		workflow.RegisterOptions{
			Name: WORKFLOW_BOOK_TRIP,
		})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
