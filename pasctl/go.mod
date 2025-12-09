module pasctl

go 1.21

require (
	github.com/chrisranney/gopas v0.0.0
	github.com/chzyer/readline v1.5.1
	github.com/fatih/color v1.16.0
	github.com/olekukonko/tablewriter v0.0.5
	golang.org/x/term v0.15.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

replace github.com/chrisranney/gopas => ../
