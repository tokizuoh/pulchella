package main

import (
	"reflect"
	"testing"
)

// TODO: テスト書き直す（mainに処理の実態がないため）
func TestRemoveEmpty(t *testing.T) {
	type args struct {
		arr []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "success", args: args{arr: []string{"1", "2", "", "3"}}, want: []string{"1", "2", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := removeEmpty(tt.args.arr); !reflect.DeepEqual(res, tt.want) {
				t.Fatalf("removeEmpty() = %v, want %v", res, tt.want)
			}
		})
	}
}
