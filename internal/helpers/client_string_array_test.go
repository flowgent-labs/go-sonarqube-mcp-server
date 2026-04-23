package mcputils

import (
	"reflect"
	"testing"
)

func TestGetStringArray(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		key  string
		want []string
	}{
		{
			name: "cli single string",
			args: map[string]interface{}{"files": "src/com/xx/xxx/HelloWorld.java"},
			key:  "files",
			want: []string{"src/com/xx/xxx/HelloWorld.java"},
		},
		{
			name: "mcp json array",
			args: map[string]interface{}{"files": []interface{}{"a.java", "b.java"}},
			key:  "files",
			want: []string{"a.java", "b.java"},
		},
		{
			name: "comma separated cli string",
			args: map[string]interface{}{"files": "a.java, b.java"},
			key:  "files",
			want: []string{"a.java", "b.java"},
		},
		{
			name: "go string slice",
			args: map[string]interface{}{"files": []string{"a.java"}},
			key:  "files",
			want: []string{"a.java"},
		},
		{
			name: "missing key",
			args: map[string]interface{}{},
			key:  "files",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringArray(tt.args, tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
