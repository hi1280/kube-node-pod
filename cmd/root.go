package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
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

			nodeList, podList := fetch()
			printNodeList(nodeList, podList)
			printPodList(nodeList, podList)
		},
	}
)

type Node struct {
	name  string
	taint string
}

type Pod struct {
	name       string
	namespace  string
	nodeName   string
	status     string
	age        string
	toleration string
	kind       string
}

func printNodeList(nodeList []Node, podList []Pod) {
	table := tablewriter.NewWriter(os.Stdout)
	for _, node := range nodeList {
		table.Append([]string{
			node.name,
			node.taint,
		})
	}
	table.Render()
}

func printPodList(nodeList []Node, podList []Pod) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NODE", "NAMESPACE", "POD", "STATUS", "AGE"})
	for _, pod := range podList {
		table.Append([]string{
			pod.nodeName,
			pod.namespace,
			pod.name,
			pod.status,
			pod.age,
			pod.kind,
			// pod.toleration,
		})
	}
	table.Render()
}

func fetch() ([]Node, []Pod) {
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

	var nodes []Node
	colorNodeNameMap := make(map[string]string)
	for i, node := range nodeList.Items {
		nodes = append(nodes, Node{
			name:  node.Name,
			taint: convertTaints(node),
		})
		colorNodeNameMap[node.Name] = changeColor(i, node.Name)
	}

	sortPodList(podList)

	dyn, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("Error creating discovery client: %v\n", err)
		os.Exit(1)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	var pods []Pod
	for _, pod := range podList.Items {
		kind := ""
		obj, err := ownedByPod(pod, mapper, dyn)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		kind = obj.GetKind()
		for {
			obj, err = ownedBy(obj, mapper, dyn)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if obj == nil {
				break
			}
			kind = obj.GetKind()
		}
		pods = append(pods, Pod{
			name:      pod.Name,
			namespace: pod.Namespace,
			nodeName:  colorNodeNameMap[pod.Spec.NodeName],
			status:    string(pod.Status.Phase),
			age:       translateTimestampSince(pod.Status.StartTime),
			kind:      kind,
			// toleration: convertTolerations(pod),
		})
	}

	return nodes, pods
}

func ownedByPod(obj v1.Pod, mapper *restmapper.DeferredDiscoveryRESTMapper, dyn dynamic.Interface) (*unstructured.Unstructured, error) {
	var errResult error
	var out *unstructured.Unstructured

	for _, ownerRef := range obj.GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			errResult = err
		}
		mapping, err := mapper.RESTMapping(schema.GroupKind{
			Group: gv.Group,
			Kind:  ownerRef.Kind,
		}, gv.Version)
		if err != nil {
			errResult = err
		}
		var rs dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			rs = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			rs = dyn.Resource(mapping.Resource)
		}
		out, err = rs.Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			errResult = err
		}
	}
	return out, errResult
}

func ownedBy(obj *unstructured.Unstructured, mapper *restmapper.DeferredDiscoveryRESTMapper, dyn dynamic.Interface) (*unstructured.Unstructured, error) {
	var errResult error
	var out *unstructured.Unstructured
	for _, ownerRef := range obj.GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			errResult = err
		}
		mapping, err := mapper.RESTMapping(schema.GroupKind{
			Group: gv.Group,
			Kind:  ownerRef.Kind,
		}, gv.Version)
		if err != nil {
			errResult = err
		}
		var rs dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			rs = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			rs = dyn.Resource(mapping.Resource)
		}
		out, err = rs.Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			errResult = err
		}
	}
	return out, errResult
}

func convertTaints(node v1.Node) string {
	var strs []string
	for _, taint := range node.Spec.Taints {
		if taint.Key == "" && taint.Value == "" {
			continue
		}
		strs = append(strs, fmt.Sprintf("%v:%v", taint.Key, taint.Value))
	}
	return strings.Join(strs, ",")
}

func convertTolerations(pod v1.Pod) string {
	var strs []string
	for _, toleration := range pod.Spec.Tolerations {
		if toleration.Key == "" && toleration.Value == "" {
			continue
		}
		strs = append(strs, fmt.Sprintf("%v:%v", toleration.Key, toleration.Value))
	}
	return strings.Join(strs, ",")
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

func translateTimestampSince(timestamp *metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
