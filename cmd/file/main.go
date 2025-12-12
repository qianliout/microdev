package main

import (
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
	"microdev/pkg/output"
)

type analysisResult struct {
	Path            string
	Format          string
	IsExecutableFmt bool
	HasExecPerms    bool
	HasDynamicLibs  bool
	PlatformOS      string
	PlatformArch    string
	IsGoBinary      bool
	GoVersion       string
	CgoEnabled      string
}

func main() {
	log := logger.NewLogger()

	var rootCmd = &cobra.Command{
		Use:   "file <path>",
		Short: "检测二进制文件信息",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			cfg, err := config.LoadConfig()
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			_ = cfg

			res, err := analyze(target)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			out := output.NewProcessor("", log)
			defer func() { _ = out.Close() }()

			emitResult(out, res)
			return nil
		},
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func analyze(path string) (analysisResult, error) {
	var res analysisResult
	abs := path
	if !filepath.IsAbs(abs) {
		p, err := filepath.Abs(path)
		if err == nil {
			abs = p
		}
	}
	res.Path = abs

	st, err := os.Stat(path)
	if err != nil {
		return res, err
	}
	mode := st.Mode()
	res.HasExecPerms = mode&0111 != 0

	if f, err := macho.OpenFat(path); err == nil {
		res.Format = "mach-o.fat"
		exec := false
		dyn := false
		osName := "darwin"
		archs := ""
		for i, a := range f.Arches {
			if a.File != nil {
				if a.File.Type == macho.TypeExec || a.File.Type == macho.TypeDylib {
					exec = true
				}
				libs, _ := a.File.ImportedLibraries()
				if len(libs) > 0 {
					dyn = true
				}
				aStr := mapMachOCpu(a.Cpu)
				if i == 0 {
					archs = aStr
				} else {
					archs += "," + aStr
				}
			}
		}
		res.IsExecutableFmt = exec
		res.HasDynamicLibs = dyn
		res.PlatformOS = osName
		res.PlatformArch = archs
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		res.Format = "mach-o"
		res.IsExecutableFmt = f.Type == macho.TypeExec || f.Type == macho.TypeDylib
		libs, _ := f.ImportedLibraries()
		res.HasDynamicLibs = len(libs) > 0
		res.PlatformOS = "darwin"
		res.PlatformArch = mapMachOCpu(f.Cpu)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		res.Format = "elf"
		res.IsExecutableFmt = f.FileHeader.Type == elf.ET_EXEC || f.FileHeader.Type == elf.ET_DYN
		libs, _ := f.ImportedLibraries()
		res.HasDynamicLibs = len(libs) > 0
		res.PlatformOS = mapELFOSABI(f.FileHeader.OSABI)
		res.PlatformArch = mapELFMachine(f.FileHeader.Machine)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := pe.Open(path); err == nil {
		defer f.Close()
		res.Format = "pe"
		res.IsExecutableFmt = true
		res.HasDynamicLibs = false
		res.PlatformOS = "windows"
		res.PlatformArch = mapPEMachine(f.Machine)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	res.Format = "unknown"
	res.IsExecutableFmt = false
	res.HasDynamicLibs = false
	res.PlatformOS = ""
	res.PlatformArch = ""
	fillGoBuildInfo(path, &res)
	return res, nil
}

func fillGoBuildInfo(path string, res *analysisResult) {
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return
	}
	res.IsGoBinary = true
	res.GoVersion = bi.GoVersion
	for _, s := range bi.Settings {
		if s.Key == "GOOS" {
			res.PlatformOS = s.Value
		}
		if s.Key == "GOARCH" {
			res.PlatformArch = s.Value
		}
		if s.Key == "CGO_ENABLED" {
			res.CgoEnabled = s.Value
		}
	}
}

func mapMachOCpu(c macho.Cpu) string {
	switch c {
	case macho.CpuAmd64:
		return "amd64"
	case macho.CpuArm64:
		return "arm64"
	case macho.Cpu386:
		return "386"
	case macho.CpuArm:
		return "arm"
	default:
		return c.String()
	}
}

func mapELFMachine(m elf.Machine) string {
	switch m {
	case elf.EM_X86_64:
		return "amd64"
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_386:
		return "386"
	case elf.EM_ARM:
		return "arm"
	default:
		return m.String()
	}
}

func mapELFOSABI(o elf.OSABI) string {
	switch o {
	case elf.ELFOSABI_LINUX:
		return "linux"
	case elf.ELFOSABI_FREEBSD:
		return "freebsd"
	case elf.ELFOSABI_NETBSD:
		return "netbsd"
	case elf.ELFOSABI_SOLARIS:
		return "solaris"
	default:
		return "linux"
	}
}

func mapPEMachine(m uint16) string {
	switch m {
	case 0x8664:
		return "amd64"
	case 0x014c:
		return "386"
	case 0x01c4:
		return "arm"
	case 0xaa64:
		return "arm64"
	default:
		return fmt.Sprintf("0x%x", m)
	}
}

func emitResult(out *output.Processor, res analysisResult) {
	fmt.Fprintf(out, "文件: %s\n", res.Path)
	fmt.Fprintf(out, "二进制格式: %s\n", res.Format)
	fmt.Fprintf(out, "可执行格式: %t\n", res.IsExecutableFmt)
	fmt.Fprintf(out, "执行权限: %t\n", res.HasExecPerms)
	fmt.Fprintf(out, "动态链接库: %t\n", res.HasDynamicLibs)
	if res.PlatformOS != "" || res.PlatformArch != "" {
		fmt.Fprintf(out, "平台: %s/%s\n", res.PlatformOS, res.PlatformArch)
	} else {
		fmt.Fprintf(out, "平台: 未知\n")
	}
	if res.IsGoBinary {
		fmt.Fprintf(out, "Go构建: true\n")
		fmt.Fprintf(out, "Go版本: %s\n", res.GoVersion)
		if res.CgoEnabled != "" {
			fmt.Fprintf(out, "CGO_ENABLED: %s\n", res.CgoEnabled)
		}
	} else {
		fmt.Fprintf(out, "Go构建: false\n")
	}
}
