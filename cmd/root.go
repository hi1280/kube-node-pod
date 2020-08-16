package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig *rest.Config

var rootCmd = &cobra.Command{
	Use:   "kube-node-pod",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := kubernetes.NewForConfig(kubeconfig)
		if err != nil {
			fmt.Printf("Error connecting to Kubernetes: %v\n", err)
			os.Exit(1)
		}
		nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing Nodes: %v\n", err)
			os.Exit(2)
		}
		for _, node := range nodeList.Items {
			fmt.Println(node.GetName())
		}
	},
}

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
		panic(err.Error())
	}
	kubeconfig = config
}
