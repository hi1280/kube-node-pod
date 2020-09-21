package pkg

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/logrusorgru/aurora"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type resource interface {
	fetch(ownerRef metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error)
}

type Fetch struct {
	Config *rest.Config
}

type unstructuredResource struct {
	mapper *restmapper.DeferredDiscoveryRESTMapper
	dyn    dynamic.Interface
}

func (f *Fetch) FetchNodesAndPods() ([]printNode, []printPod) {
	clientset, err := kubernetes.NewForConfig(f.Config)
	if err != nil {
		fmt.Printf("Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	dyn, err := dynamic.NewForConfig(f.Config)
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(f.Config)
	if err != nil {
		fmt.Printf("Error creating discovery client: %v\n", err)
		os.Exit(1)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

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

	var nodes []printNode
	colorNodeNameMap := make(map[string]string)
	for i, node := range nodeList.Items {
		nodeName := changeColor(i, node.Name)
		podNames := []string{}
		if node.Spec.Taints != nil {
			for _, pod := range podList.Items {
				if isMatchingTolerations(node.Spec.Taints, pod.Spec.Tolerations) {
					podNames = append(podNames, pod.Name)
				}
			}
		}
		if len(podNames) > 0 {
			for _, pod := range podNames {
				nodes = append(nodes, printNode{
					name:           nodeName,
					tolerationPods: pod,
				})
			}
		} else {
			nodes = append(nodes, printNode{
				name:           nodeName,
				tolerationPods: "*",
			})
		}

		colorNodeNameMap[node.Name] = nodeName
	}

	sortPodList(podList)

	var pods []printPod
	for _, pod := range podList.Items {
		u := &unstructuredResource{
			mapper: mapper,
			dyn:    dyn,
		}
		pods = append(pods, printPod{
			name:      pod.Name,
			namespace: pod.Namespace,
			nodeName:  colorNodeNameMap[pod.Spec.NodeName],
			status:    string(pod.Status.Phase),
			age:       translateTimestampSince(pod.Status.StartTime),
			kind:      fetchKindOwnedBy(u, pod),
		})
	}

	return nodes, pods
}

func fetchKindOwnedBy(rs resource, pod v1.Pod) string {
	kind := ""
	owners := pod.GetOwnerReferences()
	namespace := pod.GetNamespace()
	obj, err := ownedBy(rs, owners, namespace)
	if err != nil {
		fmt.Printf("Error getting Resource: %v\n", err)
		os.Exit(1)
	}
	if obj == nil {
		return kind
	}
	kind = obj.GetKind()
	for {
		owners = obj.GetOwnerReferences()
		namespace = obj.GetNamespace()
		obj, err = ownedBy(rs, owners, namespace)
		if err != nil {
			fmt.Printf("Error getting Resource: %v\n", err)
			os.Exit(1)
		}
		if obj == nil {
			break
		}
		kind = obj.GetKind()
	}
	return kind
}

func ownedBy(rs resource, owners []metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
	for _, ownerRef := range owners {
		return rs.fetch(ownerRef, namespace)
	}
	return nil, nil
}

func (u unstructuredResource) fetch(ownerRef metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
	gv, err := schema.ParseGroupVersion(ownerRef.APIVersion)
	if err != nil {
		return nil, err
	}
	mapping, err := u.mapper.RESTMapping(schema.GroupKind{
		Group: gv.Group,
		Kind:  ownerRef.Kind,
	}, gv.Version)
	if err != nil {
		return nil, err
	}
	var rs dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		rs = u.dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		rs = u.dyn.Resource(mapping.Resource)
	}
	return rs.Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
}

func isMatchingTolerations(taints []v1.Taint, tolerations []v1.Toleration) bool {
	if len(taints) == 0 {
		return true
	}
	if len(tolerations) == 0 && len(taints) > 0 {
		return false
	}
	for i := range taints {
		tolerated := false
		for j := range tolerations {
			if tolerations[j].ToleratesTaint(&taints[i]) {
				tolerated = true
				break
			}
		}
		if !tolerated {
			return false
		}
	}
	return true
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

func translateTimestampSince(timestamp *metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
