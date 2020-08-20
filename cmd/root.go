package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var config *rest.Config

var rootCmd = &cobra.Command{
	Use:   "kube-node-pod",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error connecting to Kubernetes: %v\n", err)
			os.Exit(1)
		}
		nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing Nodes: %v\n", err)
			os.Exit(2)
		}
		for _, node := range nodeList.Items {
			fmt.Println(node.GetName())
		}
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
	configFlags := genericclioptions.NewConfigFlags(true)
	resourceBuilderFlags := genericclioptions.NewResourceBuilderFlags()
	resourceBuilderFlags.WithAllNamespaces(false)
	resourceBuilderFlags.WithAll(false)
	configFlags.AddFlags(rootCmd.PersistentFlags())
	resourceBuilderFlags.AddFlags(rootCmd.PersistentFlags())
	c, err := configFlags.ToRESTConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config = c
}
