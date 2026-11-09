package installer

import (
	"context"
	"reflect"
	"testing"

	"github.com/lsewcx/toolup/internal/execx"
)

type call struct {
	name     string
	args     []string
	elevated bool
}

type fakeRunner struct {
	calls []call
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	f.calls = append(f.calls, call{name: name, args: args})
	return nil
}

func (f *fakeRunner) RunElevated(_ context.Context, name string, args ...string) error {
	f.calls = append(f.calls, call{name: name, args: args, elevated: true})
	return nil
}

func TestInstallCommands(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		command  string
		args     []string
		elevated bool
	}{
		{
			name:     "apt",
			manager:  "apt",
			command:  "apt-get",
			args:     []string{"install", "-y", "--no-install-recommends", "git"},
			elevated: true,
		},
		{
			name:     "dnf",
			manager:  "dnf",
			command:  "dnf",
			args:     []string{"install", "-y", "git"},
			elevated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeRunner{}
			inst, err := New(tt.manager, runner)
			if err != nil {
				t.Fatal(err)
			}

			if err := inst.Install(context.Background(), "git"); err != nil {
				t.Fatal(err)
			}

			want := []call{{
				name:     tt.command,
				args:     tt.args,
				elevated: tt.elevated,
			}}
			if !reflect.DeepEqual(runner.calls, want) {
				t.Fatalf("calls = %#v, want %#v", runner.calls, want)
			}
		})
	}
}

var _ execx.Runner = (*fakeRunner)(nil)
