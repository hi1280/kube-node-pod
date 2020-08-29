package cmd

import (
	"testing"

	"github.com/logrusorgru/aurora"
)

func Test_changeColor(t *testing.T) {
	type args struct {
		i   int
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"a", args{i: 0, str: "a"}, aurora.Index(1, "a").String()},
		{"a", args{i: 1, str: "a"}, aurora.Index(2, "a").String()},
		{"a", args{i: 2, str: "a"}, aurora.Index(3, "a").String()},
		{"a", args{i: 3, str: "a"}, aurora.Index(4, "a").String()},
		{"a", args{i: 4, str: "a"}, aurora.Index(5, "a").String()},
		{"a", args{i: 5, str: "a"}, aurora.Index(6, "a").String()},
		{"a", args{i: 6, str: "a"}, aurora.Index(1, "a").String()},
		{"a", args{i: 7, str: "a"}, aurora.Index(2, "a").String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changeColor(tt.args.i, tt.args.str); got != tt.want {
				t.Errorf("changeColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
