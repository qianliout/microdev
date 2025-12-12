package service

import (
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"

	"microdev/component/file/model"
	"microdev/pkg/logger"
)

type AnalyzerService struct {
	log *logger.Logger
}

func NewAnalyzerService(log *logger.Logger) *AnalyzerService {
	return &AnalyzerService{log: log}
}

func (s *AnalyzerService) Analyze(path string) (model.Result, error) {
	var res model.Result
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
	res.HasExecPerms = st.Mode()&0111 != 0

	if f, err := macho.OpenFat(path); err == nil {
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
		res.Format = "mach-o.fat"
		res.IsExecutable = exec
		res.HasDynamicLibs = dyn
		res.GOOS = osName
		res.GOARCH = archs
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		res.Format = "mach-o"
		res.IsExecutable = f.Type == macho.TypeExec || f.Type == macho.TypeDylib
		libs, _ := f.ImportedLibraries()
		res.HasDynamicLibs = len(libs) > 0
		res.GOOS = "darwin"
		res.GOARCH = mapMachOCpu(f.Cpu)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		res.Format = "elf"
		res.IsExecutable = f.FileHeader.Type == elf.ET_EXEC || f.FileHeader.Type == elf.ET_DYN
		libs, _ := f.ImportedLibraries()
		res.HasDynamicLibs = len(libs) > 0
		res.GOOS = mapELFOSABI(f.FileHeader.OSABI)
		res.GOARCH = mapELFMachine(f.FileHeader.Machine)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	if f, err := pe.Open(path); err == nil {
		defer f.Close()
		res.Format = "pe"
		res.IsExecutable = true
		res.HasDynamicLibs = false
		res.GOOS = "windows"
		res.GOARCH = mapPEMachine(f.Machine)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	res.Format = "unknown"
	res.IsExecutable = false
	res.HasDynamicLibs = false
	fillGoBuildInfo(path, &res)
	return res, nil
}

func fillGoBuildInfo(path string, res *model.Result) {
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return
	}
	res.IsGoBinary = true
	res.GoVersion = bi.GoVersion
	for _, s := range bi.Settings {
		if s.Key == "GOOS" {
			res.GOOS = s.Value
		}
		if s.Key == "GOARCH" {
			res.GOARCH = s.Value
		}
		if s.Key == "CGO_ENABLED" {
			res.CGOEnabled = s.Value
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
