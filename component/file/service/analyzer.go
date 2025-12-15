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

func NewAnalyzerService() *AnalyzerService {
	log := logger.NewLogger(logger.WithModule("FileCheck", ""))
	return &AnalyzerService{log: log}
}

// 把结果规整，美观的输出到控制台
// 每个结果一行，要美观，容易读
func (s *AnalyzerService) Output(res *model.Result) error {
	s.log.UserInfo(fmt.Sprintf("File Path: %s", res.Path))
	s.log.UserInfo(fmt.Sprintf("File Format: %s", res.Format))
	s.log.UserInfo(fmt.Sprintf("Executable: %t", res.IsExecutable))
	s.log.UserInfo(fmt.Sprintf("Has Exec Perms: %t", res.HasExecPerms))
	s.log.UserInfo(fmt.Sprintf("Has Dynamic Libs: %t", res.HasDynamicLibs))
	s.log.UserInfo(fmt.Sprintf("GOOS: %s", res.GOOS))
	s.log.UserInfo(fmt.Sprintf("GOARCH: %s", res.GOARCH))
	s.log.UserInfo(fmt.Sprintf("Is Go Binary: %t", res.IsGoBinary))
	s.log.UserInfo(fmt.Sprintf("Go Version: %s", res.GoVersion))
	s.log.UserInfo(fmt.Sprintf("CGO Enabled: %s", res.CGOEnabled))
	return nil
}

// Analyze 以“逐层识别 + 构建信息补充”的方式分析目标文件：
//  1. 规范化路径并记录；通过 os.Stat 检查是否具备可执行权限位（HasExecPerms）。
//  2. 优先尝试 Mach-O fat：遍历每个架构，判断是否为可执行/可装载（IsExecutable），
//     是否存在动态库依赖（HasDynamicLibs），汇总架构为逗号分隔的 GOARCH，并设置 GOOS=darwin。
//  3. 若不是 fat，则尝试普通 Mach-O（thin），按同样规则填充字段。
//  4. 若不是 Mach-O，则尝试 ELF：依据文件头类型判断可执行/装载，读取 DT_NEEDED 作为依赖，
//     并根据 OSABI/Machine 推断 GOOS/GOARCH。
//  5. 若不是 ELF，则尝试 PE：读取导入库判断是否有动态依赖，设置 GOOS=windows 与架构。
//  6. 都不匹配时标记为 unknown。
//  7. 最后尝试读取 Go 构建信息（debug/buildinfo），若存在则覆盖/补充 GOOS、GOARCH、GoVersion、CGOEnabled，
//     并将 IsGoBinary 置为 true，以便更准确展示目标平台与构建参数。
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

	if fat, err := macho.OpenFat(path); err == nil {
		defer fat.Close()
		res.Format = "mach-o.fat"
		res.GOOS = "darwin"
		exec := false
		dyn := false
		archs := ""
		for i, a := range fat.Arches {
			if a.File == nil {
				continue
			}
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
		res.IsExecutable = exec
		res.HasDynamicLibs = dyn
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
		libs, _ := f.ImportedLibraries()
		res.HasDynamicLibs = len(libs) > 0
		res.GOOS = "windows"
		res.GOARCH = mapPEMachine(f.Machine)
		fillGoBuildInfo(path, &res)
		return res, nil
	}

	fillGoBuildInfo(path, &res)
	res.SetDefault()
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
