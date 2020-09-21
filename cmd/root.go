package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	configFlags *genericclioptions.ConfigFlags
	kubeconfig  *rest.Config
	rootCmd     = &cobra.Command{
		Use:   "kube-node-pod",
		Short: "kube-node-pod provides an overview of nodes and pods",
		Long:  "kube-node-pod provides an overview of nodes and pods",
		Run: func(cmd *cobra.Command, args []string) {

			nodeList, podList := fetchNodesAndPods()
			printNodeList(nodeList)
			printPodList(podList)
		},
	}
)

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

// Execute is entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
