package model

type Result struct {
	Path           string
	Format         string
	IsExecutable   bool
	HasExecPerms   bool
	HasDynamicLibs bool
	GOOS           string
	GOARCH         string
	IsGoBinary     bool
	GoVersion      string
	CGOEnabled     string
}
