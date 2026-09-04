package ssh

import (
	"reflect"
	"testing"
)

func TestFilterOutPosixFDState(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want []string
	}{
		{
			name: "drops fd-state vars with any hash prefix",
			env: []string{
				"PATH=C:\\Windows",
				"C28FC6F98A2C44ABBD89D6A3037D0D9_POSIX_FD_STATE=AAAAAAICAwA=",
				"HOME=/home/user",
			},
			want: []string{"PATH=C:\\Windows", "HOME=/home/user"},
		},
		{
			name: "keeps vars whose value contains the marker",
			env: []string{
				"DEBUG_NOTE=see _POSIX_FD_STATE= docs",
			},
			want: []string{"DEBUG_NOTE=see _POSIX_FD_STATE= docs"},
		},
		{
			name: "keeps vars without an equals sign",
			env:  []string{"JUST_A_NAME", "FOO=bar"},
			want: []string{"JUST_A_NAME", "FOO=bar"},
		},
		{
			name: "keeps var whose name only partly matches",
			env:  []string{"FOO_POSIX_FD_STATE_guard=1"},
			want: []string{"FOO_POSIX_FD_STATE_guard=1"},
		},
		{
			name: "drops bare variable name (no hash prefix)",
			env:  []string{"_POSIX_FD_STATE=abc"},
			want: []string{},
		},
		{
			name: "empty environment",
			env:  []string{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterOutPosixFDState(tt.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterOutPosixFDState() = %v, want %v", got, tt.want)
			}
		})
	}
}
