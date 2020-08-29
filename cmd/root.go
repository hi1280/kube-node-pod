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
	v1 "k8s.io/api/core/v1"
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
	rootCmd     = &cobra.Command{
		Use:   "kube-node-pod",
		Short: "kube-node-pod provides an overview of nodes and pods",
		Long:  "kube-node-pod provides an overview of nodes and pods",
		Run: func(cmd *cobra.Command, args []string) {

			podList := fetch()

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"NODE", "NAMESPACE", "POD", "STATUS", "AGE"})
			for _, pod := range podList {
				table.Append([]string{
					pod.nodeName,
					pod.namespace,
					pod.name,
					pod.status,
					pod.age,
				})
			}
			table.Render()
		},
	}
)

type printPod struct {
	name      string
	namespace string
	nodeName  string
	status    string
	age       string
}

func fetch() []printPod {
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

	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing Pods: %v\n", err)
		os.Exit(1)
	}

	nodes := make(map[string]string)
	for i, node := range nodeList.Items {
		nodes[node.Name] = changeColor(i, node.Name)
	}

	sortPodList(podList)

	var printPodList []printPod
	for _, pod := range podList.Items {
		printPodList = append(printPodList, printPod{
			name:      pod.Name,
			namespace: pod.Namespace,
			nodeName:  nodes[pod.Spec.NodeName],
			status:    string(pod.Status.Phase),
			age:       translateTimestampSince(*pod.Status.StartTime),
		})
	}

	return printPodList
}

func changeColor(i int, str string) string {
	return aurora.Index(uint8(exceptBlackAndWhiteNumber(i)), str).String()
}

func exceptBlackAndWhiteNumber(i int) int {
	return i%6 + 1
}

func sortPodList(podList *v1.PodList) {
	sort.Slice(podList.Items, func(i, j int) bool {
		p1 := podList.Items[i]
		p2 := podList.Items[j]
		return p1.Spec.NodeName < p2.Spec.NodeName
	})
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
