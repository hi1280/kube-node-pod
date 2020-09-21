package cmd

import (
	"fmt"
	"testing"

	"github.com/logrusorgru/aurora"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_changeColor(t *testing.T) {
	type args struct {
		i   int
		str string
	}
	tests := []struct {
		args args
		want string
	}{
		{args{i: 0, str: "a"}, aurora.Index(1, "a").String()},
		{args{i: 1, str: "a"}, aurora.Index(2, "a").String()},
		{args{i: 2, str: "a"}, aurora.Index(3, "a").String()},
		{args{i: 3, str: "a"}, aurora.Index(4, "a").String()},
		{args{i: 4, str: "a"}, aurora.Index(5, "a").String()},
		{args{i: 5, str: "a"}, aurora.Index(6, "a").String()},
		{args{i: 6, str: "a"}, aurora.Index(1, "a").String()},
		{args{i: 7, str: "a"}, aurora.Index(2, "a").String()},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%v,%v", tt.args.i, tt.args.str)
		t.Run(testname, func(t *testing.T) {
			if got := changeColor(tt.args.i, tt.args.str); got != tt.want {
				t.Errorf("changeColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchKindOwnedBy(t *testing.T) {
	u := &fakeUnstructured{
		fakeFetch: func(ownerRef metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
			return &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
				},
			}, nil
		},
	}
	type args struct {
		rs  resource
		pod v1.Pod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"if pod doesn't have owner kind is empty", args{rs: u, pod: v1.Pod{}}, ""},
		{"if pod has owner kind is owner's", args{
			rs: u,
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: "Deployment"},
					},
				},
			}}, "Deployment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fetchKindOwnedBy(tt.args.rs, tt.args.pod); got != tt.want {
				t.Errorf("fetchKindOwnedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeUnstructured struct {
	fakeFetch func(ownerRef metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error)
}

func (u *fakeUnstructured) fetch(ownerRef metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
	return u.fakeFetch(ownerRef, namespace)
}
