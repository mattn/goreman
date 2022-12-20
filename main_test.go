package main

import (
	"reflect"
	"testing"
)

func Test_extractVariables(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr error
	}{
		{name: "extract one variable",
			args: args{
				command: "curl something.com:{{.port}}",
			},
			want: []string{"port"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, wantErr := extractVariables(tt.args.command)
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractVariables() = %v, want %v", got, tt.want)
			}
			if wantErr != tt.wantErr {
				t.Errorf("extractVariables() = %v, wantErr %v", got, tt.wantErr)
			}
		})
	}
}

func TestService_Valid(t *testing.T) {
	type fields struct {
		Command     string
		Environment string
		Enable      bool
		Variables   []map[string]string
	}
	t1 := make(map[string]string)
	t1["port"] = "1001"

	t2 := make(map[string]string)
	t2["port"] = "1002"

	tests := []struct {
		name    string
		service Service
		want    bool
	}{
		{
			name:    "disabled service",
			service: Service{Command: "something {{.port}}", Enable: false, Variables: []map[string]string{t1, t2}},
			want:    false,
		},
		{
			name:    "enabled service, wrong variable count",
			service: Service{Command: "something {{.port}} and some other variable {{.testing}}", Enable: true, Variables: []map[string]string{t1}},
			want:    false,
		},
		{
			name:    "enabled service",
			service: Service{Command: "something {{.Port}}", Enable: true, Variables: []map[string]string{t1}},
			want:    true,
		},
		{
			name:    "enabled service, but missing variable",
			service: Service{Command: "something {{.trop}}", Enable: true, Variables: []map[string]string{t1}},
			want:    false,
		},
		{
			name:    "enabled service, but invalid variable syntax",
			service: Service{Command: "something {trop}}", Enable: true, Variables: []map[string]string{t1}},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{
				Command:     tt.service.Command,
				Environment: tt.service.Environment,
				Enable:      tt.service.Enable,
				Variables:   tt.service.Variables,
			}
			if got := s.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
