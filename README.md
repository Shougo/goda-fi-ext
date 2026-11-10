# goda-fi-ext

The sources, filters, kinds and columns of
[goda-fi.vim](https://github.com/Shougo/goda-fi.vim): Go packages for its
configuration. Their Vim script is embedded in them (sdk/vimscript of
goda.vim): they are not Vim plugins, nothing to install but the module.

```go
import (
	"github.com/Shougo/goda-fi-ext/ext/filter/matchersubstring"
	kindfile "github.com/Shougo/goda-fi-ext/ext/kind/file"
	"github.com/Shougo/goda-fi-ext/ext/source/filerec"
)

b.AddSource(filerec.New(filerec.Params{}))
b.AddFilter(matchersubstring.New(matchersubstring.Params{}))
b.AddKind(kindfile.New())
```

The UIs are [goda-fi-ui-ff.vim](https://github.com/Shougo/goda-fi-ui-ff.vim)
and [goda-fi-ui-filer.vim](https://github.com/Shougo/goda-fi-ui-filer.vim).

## Sources

ddu.vim's names lose their "_": "file_rec" is "filerec". The directory is
sourceOptions.path, or the current one.

### file

ext/source/file (ddu-source-file): the files of a directory; the directories are trees (expandItem). Params: ignoreDirectories, new.

### filerec

ext/source/filerec (ddu-source-file_rec): the files under a directory. Params: chunkSize (1000), ignoredDirectories ([".git"]), expandSymbolicLink.

### fileold

ext/source/fileold (ddu-source-file_old): v:oldfiles, the viminfo or ShaDa read again first.

### buffer

ext/source/buffer (ddu-source-buffer): the listed buffers of files, the alternate one first.

### filepoint

ext/source/filepoint (ddu-source-file_point): the file under the cursor, at the line and column after it ("path:10:3", "path", line 10, a diff...). A URL there (http://, https://) is an item of the kind url: add it too.

### fileexternal

ext/source/fileexternal (ddu-source-file_external): the files a command lists. Params: cmd (["fd", "-t", "f"]), updateItems.

### rg

ext/source/rg (ddu-source-rg): what ripgrep finds. Params: input (the pattern; the input when the source is volatile), args (["--column", "--no-heading", "--color", "never"]; with "--json" its JSON is read), category (a line of each file before its matches), cmd ("rg"), globs, highlights ({path, lineNr, word}), maxEnqueSize (10000), minVolatileInputLength (2), paths. inputType "migemo" is not supported.

### pathhistory

ext/source/pathhistory (ddu-source-path_history): the paths the session of the buffer it is started from left (narrow), the last first.

### line

ext/source/line (ddu-source-line): the lines of the buffer, from the cursor line. Params: bufNr, ignoreEmptyInput, range ("buffer" or "window"), winId.

### register

ext/source/register (ddu-source-register): the registers that are not empty, of the kind "word".

## Filters

### matchersubstring

ext/filter/matchersubstring (ddu-filter-matcher_substring): every word of the input; "!word" excludes, "<word" starts a word, "^word" is a prefix, "word$" a suffix, "\ " a space. Params: highlightMatched (""), limit (1000), maxLength (500).

### matcherhidden

ext/filter/matcherhidden (ddu-filter-matcher_hidden): no hidden files, unless the input has a ".".

### matcherrelative

ext/filter/matcherrelative (ddu-filter-matcher_relative): only the files under the directory of the source.

### matcherfiles

ext/filter/matcherfiles (ddu-filter-matcher_files): only the paths of params.globs or params.patterns.

### sorteralpha

ext/filter/sorteralpha (ddu-filter-sorter_alpha): by word.

### sortermtime

ext/filter/sortermtime (ddu-filter-sorter_mtime): the files changed last first.

### converterhldir

ext/filter/converterhldir (ddu-filter-converter_hl_dir): the directories of the path highlighted. Params: hlGroup ("Directory", or a list used by turns).

## Kinds

### file

ext/kind/file (ddu-kind-file). Actions: open (params: command "edit",
maxSize), append, insert, cd, execute (params: command), feedkeys, loclist,
quickfix, yank, narrow (the directory of the item is the path of the source;
params: path, ".." for the parent; several items: a session pushed with
their directories). Previews files and buffers.

The actions of the filer, done by the binary, asked in Vim (`input()`,
`confirm()`):

| Action | |
|---|---|
| copy, move, link (params: mode "absolute", "relative" or "hard") | the items to the clipboard |
| paste | what the clipboard holds, in the directory of the item (the tree under the cursor when it is expanded); a file there already: Force, Time (when newer), Underbar ("name_"), No, Rename |
| clipYank | the paths to the register `"+`, for clipCopy and clipMove (another Vim) |
| clipCopy, clipMove | the files of the register `"+` |
| newFile, newDirectory | files (comma separated; "name/" is a directory), directories |
| rename, edit | the items: a new name is asked for each |
| delete | (asked) |
| trash | (asked) with kindParams trashCommand (`["gio", "trash"]`) |
| undo | takes back the last copy, move, link, rename, newFile or newDirectory |
| executeSystem | opens the file with the program of the system (params: method, "" for the one of the system, "nvim-open", or a command: "xdg-open", "open", "wslview", ...) |

The clipboard and undo are shared by the sessions; the buffers of the files
moved, renamed or deleted follow them, and a modified one is not moved.

### word

ext/kind/word (ddu-kind-word). Actions: append, insert, feedkeys, yank. Previews the text.

### url

ext/kind/url (ddu-kind-url). Actions: browse (with kindParams externalOpener: "systemopen" (default: xdg-open, open, the shell of MS-Windows), "openbrowser" (open-browser.vim), "external" (vim-external), "uiopen" (Neovim's vim.ui.open())), open (params: command "edit"; as buffers, for netrw and the like), yank (params: register; one URL a line; the UI stays), fetch (params: method "GET", headers, body, showStatus, showHeader: the response in a buffer).

## Columns

Columns make what the items of a source show (sourceOptions columns), one after the other, as ddu.vim's: they are made as the items are drawn, so they show the trees as they are.

### filename

ext/column/filename (ddu-column-filename): the name of the file, indented by its level, after an icon. Params: collapsedIcon ("+"), expandedIcon ("-"), iconWidth (1), indentationWidth (1), linkIcon ("@"), maxSymlinkTargetWidth (48: "name -> target" is cut), highlights (directoryIcon Special, directoryName Directory, linkIcon and linkName Comment).
