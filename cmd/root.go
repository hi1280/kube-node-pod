package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	configFlags          *genericclioptions.ConfigFlags
	resourceBuilderFlags *genericclioptions.ResourceBuilderFlags
)

var rootCmd = &cobra.Command{
	Use:   "kube-node-pod",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := configFlags.ToRESTConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error connecting to Kubernetes: %v\n", err)
			os.Exit(1)
		}

		nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing Nodes: %v\n", err)
			os.Exit(1)
		}
		nodes := map[string]string{}
		for i, node := range nodeList.Items {
			nodes[node.Name] = aurora.Index(uint8(i%6+1), node.Name).String()
		}

		podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing Pods: %v\n", err)
			os.Exit(1)
		}
		sort.Slice(podList.Items, func(i, j int) bool {
			p1 := podList.Items[i]
			p2 := podList.Items[j]
			return p1.Spec.NodeName < p2.Spec.NodeName
		})
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"NODE", "NAMESPACE", "POD", "STATUS", "AGE"})
		for _, pod := range podList.Items {
			table.Append([]string{
				nodes[pod.Spec.NodeName],
				pod.Namespace,
				pod.Name,
				string(pod.Status.Phase),
				translateTimestampSince(*pod.Status.StartTime),
			})
		}
		table.Render()
	},
}

// Execute is entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	configFlags = genericclioptions.NewConfigFlags(true)
	resourceBuilderFlags = genericclioptions.NewResourceBuilderFlags()
	resourceBuilderFlags.WithAllNamespaces(false)
	configFlags.AddFlags(rootCmd.PersistentFlags())
	resourceBuilderFlags.AddFlags(rootCmd.PersistentFlags())
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
