package cmd

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

type printNode struct {
	name           string
	taint          string
	tolerationPods string
}

type printPod struct {
	name       string
	namespace  string
	nodeName   string
	status     string
	age        string
	toleration string
	kind       string
}

func printNodeList(nodeList []printNode) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NODE", "ALLOW POD"})
	for _, node := range nodeList {
		table.Append([]string{
			node.name,
			node.tolerationPods,
		})
	}
	table.Render()
}

func printPodList(podList []printPod) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NODE", "NAMESPACE", "POD", "STATUS", "AGE", "KIND OWNER"})
	for _, pod := range podList {
		table.Append([]string{
			pod.nodeName,
			pod.namespace,
			pod.name,
			pod.status,
			pod.age,
			pod.kind,
		})
	}
	table.Render()
}
