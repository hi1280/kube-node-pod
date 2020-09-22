package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hi1280/kubectl-node-pod/pkg"
	"github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig *rest.Config
	rootCmd    = &cobra.Command{
		Use:   "kubectl node-pod",
		Short: "provide an overview of nodes and pods",
		Long:  "provide an overview of nodes and pods",
		Run: func(cmd *cobra.Command, args []string) {
			f := &pkg.Fetch{
				Config: kubeconfig,
			}
			nodeList, podList := f.FetchNodesAndPods()
			pkg.PrintNodeList(nodeList)
			pkg.PrintPodList(podList)
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
