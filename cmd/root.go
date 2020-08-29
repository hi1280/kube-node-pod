package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	configFlags *genericclioptions.ConfigFlags
	kubeconfig  *rest.Config
)

var rootCmd = &cobra.Command{
	Use:   "kube-node-pod",
	Short: "kube-node-pod provides an overview of nodes and pods",
	Long:  "kube-node-pod provides an overview of nodes and pods",
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := kubernetes.NewForConfig(kubeconfig)
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
	var configString string
	if home, _ := homedir.Dir(); home != "" {
		rootCmd.PersistentFlags().StringVar(&configString, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().StringVar(&configString, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configString)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	kubeconfig = config
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
