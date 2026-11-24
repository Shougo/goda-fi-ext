package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Shougo/goda-fi-ext/ext/kind/file"
	"github.com/Shougo/goda-fi.vim/fi"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const head = "[Desktop Entry]\nType=Application\nExec=x\n"

func TestApps(t *testing.T) {
	home, sys := t.TempDir(), t.TempDir()
	// The user's entries hide the system's of the same ID.
	write(t, filepath.Join(home, "editor.desktop"), head+"Name=My Editor\n")
	write(t, filepath.Join(sys, "editor.desktop"), head+"Name=Editor\n")
	write(t, filepath.Join(home, "gone.desktop"), head+"Name=Gone\nHidden=true\n")
	write(t, filepath.Join(sys, "gone.desktop"), head+"Name=Gone\n")
	// An ID of a subdirectory.
	write(t, filepath.Join(sys, "kde", "term.desktop"), head+"Name=Term\n")
	write(t, filepath.Join(home, "kde-term.desktop"), head+"Name=Hidden Term\nNoDisplay=true\n")
	write(t, filepath.Join(sys, "kde", "calc.desktop"), head+"Name=Calc\n")
	// The name of the locale, escaped.
	write(t, filepath.Join(sys, "files.desktop"), head+"Name=Files\nName[ja]=ファイル\\sマネージャー\n# Name=No\n")
	// Not shown.
	write(t, filepath.Join(sys, "link.desktop"), "[Desktop Entry]\nType=Link\nName=Link\nURL=https://example.com\n")
	write(t, filepath.Join(sys, "only.desktop"), head+"Name=Only GNOME\nOnlyShowIn=GNOME;\n")
	write(t, filepath.Join(sys, "not.desktop"), head+"Name=Not KDE\nNotShowIn=Unity;KDE;\n")
	write(t, filepath.Join(sys, "try.desktop"), head+"Name=Missing\nTryExec="+filepath.Join(sys, "none")+"\n")
	write(t, filepath.Join(sys, "noname.desktop"), head)
	write(t, filepath.Join(sys, "readme.txt"), head+"Name=Text\n")
	// Only the group "Desktop Entry".
	write(t, filepath.Join(sys, "actions.desktop"), head+"Name=Main\n[Desktop Action new]\nName=Action\n")

	got := Apps([]string{home, sys}, []string{"ja_JP", "ja"}, []string{"KDE"})
	var names, ids []string
	for _, a := range got {
		names = append(names, a.Name)
		ids = append(ids, a.ID)
	}
	if want := []string{"Calc", "Main", "My Editor", "ファイル マネージャー"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names: got %q, want %q", names, want)
	}
	if want := []string{"kde-calc.desktop", "actions.desktop", "editor.desktop", "files.desktop"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("ids: got %q, want %q", ids, want)
	}
	if got[2].Path != filepath.Join(home, "editor.desktop") {
		t.Errorf("path: got %q", got[2].Path)
	}
}

func TestLocales(t *testing.T) {
	for _, c := range []struct {
		env  map[string]string
		want []string
	}{
		{map[string]string{"LANG": "ja_JP.UTF-8"}, []string{"ja_JP", "ja"}},
		{map[string]string{"LANG": "sr_RS.UTF-8@latin"}, []string{"sr_RS@latin", "sr_RS", "sr@latin", "sr"}},
		{map[string]string{"LANG": "de"}, []string{"de"}},
		{map[string]string{"LC_ALL": "fr_FR", "LANG": "ja_JP"}, []string{"fr_FR", "fr"}},
		{map[string]string{"LC_MESSAGES": "C", "LANG": "ja_JP"}, nil},
		{map[string]string{}, nil},
	} {
		if got := locales(func(k string) string { return c.env[k] }); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v: got %q, want %q", c.env, got, c.want)
		}
	}
}

func TestDirs(t *testing.T) {
	env := map[string]string{"HOME": "/home/u"}
	getenv := func(k string) string { return env[k] }
	j := filepath.Join
	want := []string{j("/home/u", ".local", "share", "applications"), j("/usr/local/share", "applications"), j("/usr/share", "applications")}
	if got := dirs(getenv); !reflect.DeepEqual(got, want) {
		t.Errorf("default: got %q, want %q", got, want)
	}
	env["XDG_DATA_HOME"] = "/data"
	env["XDG_DATA_DIRS"] = "/a::/b"
	want = []string{j("/data", "applications"), j("/a", "applications"), j("/b", "applications")}
	if got := dirs(getenv); !reflect.DeepEqual(got, want) {
		t.Errorf("XDG: got %q, want %q", got, want)
	}
}

func TestList(t *testing.T) {
	if got, want := list(`a;b\;c;d\se;`), []string{"a", "b;c", "d e"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGather(t *testing.T) {
	home := t.TempDir()
	write(t, filepath.Join(home, "applications", "editor.desktop"), head+"Name=Editor\n")
	env := map[string]string{"XDG_DATA_HOME": home, "XDG_DATA_DIRS": filepath.Join(home, "none")}
	for goos, want := range map[string][]fi.Item{
		"linux": {{Word: "Editor", Action: file.ActionData{Path: filepath.Join(home, "applications", "editor.desktop")}}},
		// No applications directories there.
		"windows": {},
		"darwin":  {},
	} {
		s := source{getenv: func(k string) string { return env[k] }, goos: goos}
		var items []fi.Item
		if err := s.Gather(context.Background(), nil, fi.GatherArgs{}, func(it []fi.Item) error {
			items = append(items, it...)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if len(items) != len(want) || (len(want) > 0 && !reflect.DeepEqual(items, want)) {
			t.Errorf("%s: got %+v, want %+v", goos, items, want)
		}
	}
}
