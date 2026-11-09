package catalog

var gitTool = Tool{
	Name: "git",
	Packages: map[string]string{
		"apt": "git",
		"dnf": "git",
	},
}
