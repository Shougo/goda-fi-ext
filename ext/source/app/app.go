// Package app is the source "app" of goda-fi.vim: the applications of the
// desktop, as an application launcher. They are the desktop entries (the
// Desktop Entry Specification of freedesktop.org) of the applications
// directories of the XDG data directories: $XDG_DATA_HOME/applications,
// then those of $XDG_DATA_DIRS.
//
// The items are of the kind file (their desktop files): its executeSystem
// launches them with params.method "gio" (or "dex", "kioclient"), and its
// open edits them.
//
// Linux and the BSDs only: there are no items on MS-Windows and macOS.
package app

import (
	"bufio"
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/Shougo/goda-fi-ext/ext/kind/file"
	"github.com/Shougo/goda-fi.vim/fi"
	"github.com/Shougo/goda.vim/sdk"
)

type source struct {
	getenv func(string) string
	goos   string
}

// New returns the source.
func New() fi.Source { return source{getenv: os.Getenv, goos: runtime.GOOS} }

func (source) Name() string { return "app" }
func (source) Kind() string { return "file" }

func (s source) Gather(_ context.Context, _ sdk.Host, _ fi.GatherArgs, emit func([]fi.Item) error) error {
	if s.goos == "windows" || s.goos == "darwin" {
		return emit([]fi.Item{})
	}
	apps := Apps(dirs(s.getenv), locales(s.getenv), desktops(s.getenv))
	items := make([]fi.Item, 0, len(apps))
	for _, a := range apps {
		items = append(items, fi.Item{Word: a.Label(), Action: file.ActionData{Path: a.Path, Preview: a.Description()}})
	}
	return emit(items)
}

// App is an application: a desktop entry to show.
type App struct {
	// ID is the desktop file ID: its path under the applications
	// directory, "/" made "-" ("kde-foo.desktop" for kde/foo.desktop).
	ID   string
	Path string
	// Name and GenericName ("Web Browser") are those of the locale.
	Name        string
	GenericName string
	// Comment ("Browse the Web") and Keywords are those of the locale too.
	Comment  string
	Keywords []string
	Exec     string
}

// Description is what the preview shows: the names, the comment, the
// keywords, the command and the desktop file.
func (a App) Description() []string {
	lines := []string{a.Name}
	if a.GenericName != "" && !strings.EqualFold(a.GenericName, a.Name) {
		lines = append(lines, a.GenericName)
	}
	if a.Comment != "" {
		lines = append(lines, strings.Split(a.Comment, "\n")...)
	}
	lines = append(lines, "")
	if len(a.Keywords) > 0 {
		lines = append(lines, "Keywords: "+strings.Join(a.Keywords, ", "))
	}
	if a.Exec != "" {
		lines = append(lines, "Exec: "+a.Exec)
	}
	return append(lines, a.Path)
}

// Label is what the item shows and is matched by: the name, and the generic
// name if it says more ("Firefox (Web Browser)"; matched by "browser").
func (a App) Label() string {
	if a.GenericName == "" || strings.EqualFold(a.GenericName, a.Name) {
		return a.Name
	}
	return a.Name + " (" + a.GenericName + ")"
}

// dirs are the applications directories, the first first: an ID there
// hides the same ID in the others.
func dirs(getenv func(string) string) []string {
	home := getenv("XDG_DATA_HOME")
	if home == "" && getenv("HOME") != "" {
		home = filepath.Join(getenv("HOME"), ".local", "share")
	}
	data := getenv("XDG_DATA_DIRS")
	if data == "" {
		data = "/usr/local/share:/usr/share"
	}
	var ds []string
	for _, d := range append([]string{home}, strings.Split(data, ":")...) {
		if d != "" {
			ds = append(ds, filepath.Join(d, "applications"))
		}
	}
	return ds
}

// locales are the keys of the localized values for the locale of the
// messages, the best first: lang_COUNTRY@MODIFIER, lang_COUNTRY,
// lang@MODIFIER, lang.
func locales(getenv func(string) string) []string {
	var l string
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if l = getenv(v); l != "" {
			break
		}
	}
	if l == "" || l == "C" || l == "POSIX" {
		return nil
	}
	var mod string
	if i := strings.IndexByte(l, '@'); i >= 0 {
		l, mod = l[:i], l[i:]
	}
	if i := strings.IndexByte(l, '.'); i >= 0 {
		l = l[:i]
	}
	lang, _, hasCountry := strings.Cut(l, "_")
	var ls []string
	if hasCountry && mod != "" {
		ls = append(ls, l+mod)
	}
	if hasCountry {
		ls = append(ls, l)
	}
	if mod != "" {
		ls = append(ls, lang+mod)
	}
	return append(ls, lang)
}

// desktops are those of $XDG_CURRENT_DESKTOP, for OnlyShowIn and
// NotShowIn.
func desktops(getenv func(string) string) []string {
	var ds []string
	for _, d := range strings.Split(getenv("XDG_CURRENT_DESKTOP"), ":") {
		if d != "" {
			ds = append(ds, d)
		}
	}
	return ds
}

// Apps are the applications of the directories, by name. An entry hidden
// (Hidden, NoDisplay, OnlyShowIn, NotShowIn, TryExec) still hides the same
// ID in the directories after it, as the specification says: a user
// hides an application with a desktop file of the same ID.
func Apps(dirs, locales, desktops []string) []App {
	seen := map[string]bool{}
	var apps []App
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".desktop") {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}
			id := strings.ReplaceAll(filepath.ToSlash(rel), "/", "-")
			if seen[id] {
				return nil
			}
			seen[id] = true
			b, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			e := parse(b)
			if name, ok := e.show(locales, desktops); ok {
				apps = append(apps, App{ID: id, Path: path, Name: name,
					GenericName: e.localized("GenericName", locales),
					Comment:     e.localized("Comment", locales),
					Keywords:    list(e.localizedRaw("Keywords", locales)),
					Exec:        unescape(e["Exec"])})
			}
			return nil
		})
	}
	slices.SortStableFunc(apps, func(a, b App) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return apps
}

// entry is the group "Desktop Entry" of a desktop file, its values not
// unescaped.
type entry map[string]string

func parse(b []byte) entry {
	e := entry{}
	in := false
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if line[0] == '[' {
			in = line == "[Desktop Entry]"
			continue
		}
		if !in {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if _, dup := e[k]; !dup {
			e[k] = strings.TrimSpace(v)
		}
	}
	return e
}

// show is the name of the entry, if it is an application to show.
func (e entry) show(locales, desktops []string) (string, bool) {
	if e["Type"] != "Application" || e["Hidden"] == "true" || e["NoDisplay"] == "true" {
		return "", false
	}
	if only, ok := e["OnlyShowIn"]; ok && !intersects(list(only), desktops) {
		return "", false
	}
	if intersects(list(e["NotShowIn"]), desktops) {
		return "", false
	}
	if try := unescape(e["TryExec"]); try != "" && !executable(try) {
		return "", false
	}
	name := e.localized("Name", locales)
	return name, name != ""
}

func (e entry) localized(key string, locales []string) string {
	return unescape(e.localizedRaw(key, locales))
}

// localizedRaw is the value not unescaped: that of a list is split first.
func (e entry) localizedRaw(key string, locales []string) string {
	for _, l := range locales {
		if v, ok := e[key+"["+l+"]"]; ok {
			return v
		}
	}
	return e[key]
}

// list is a value of strings separated by ";" ("\;" is one in a string).
func list(v string) []string {
	var l []string
	var cur strings.Builder
	for i := 0; i < len(v); i++ {
		switch {
		case v[i] == '\\' && i+1 < len(v) && v[i+1] == ';':
			cur.WriteByte(';')
			i++
		case v[i] == ';':
			l = append(l, unescape(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(v[i])
		}
	}
	if cur.Len() > 0 {
		l = append(l, unescape(cur.String()))
	}
	return l
}

// unescape the escapes of a string value: \s, \n, \t, \r and \\.
func unescape(v string) string {
	if !strings.Contains(v, `\`) {
		return v
	}
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		if v[i] != '\\' || i+1 == len(v) {
			b.WriteByte(v[i])
			continue
		}
		i++
		switch v[i] {
		case 's':
			b.WriteByte(' ')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		default:
			b.WriteByte(v[i])
		}
	}
	return b.String()
}

func intersects(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}

// executable: TryExec, a path or a name in $PATH.
func executable(p string) bool {
	if filepath.IsAbs(p) {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir() && st.Mode()&0o111 != 0
	}
	_, err := exec.LookPath(p)
	return err == nil
}
