package file

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Shougo/goda-fi.vim/fi"
)

func TestPreview(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.desktop")
	for _, c := range []struct {
		name string
		data ActionData
		want *fi.Previewer
	}{
		{"lines", ActionData{Path: path, Preview: []string{"App", "", path}},
			&fi.Previewer{Kind: "nofile", Contents: []string{"App", "", path}}},
		{"file", ActionData{Path: path},
			&fi.Previewer{Kind: "nofile", Contents: []string{path + " is not found"}}},
	} {
		got, err := Kind{}.Preview(context.Background(), nil, fi.PreviewArgs{Item: fi.Item{Word: "a", Action: c.data}})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}
