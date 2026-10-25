package service

import (
	"context"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"microdev/infra/consts"
	"microdev/infra/logger"
	filemodel "microdev/infra/model/file"
)

type AnalyzerSrv struct {
	log *logger.Logger
}

type archInfo struct {
	exec bool
	dyn  bool
	arch string
}

func NewAnalyzerSrv() *AnalyzerSrv {
	return &AnalyzerSrv{
		log: logger.New(logger.WithModule("file"), logger.WithSubModule("service")),
	}
}

// Analyze 先按文件格式逐层识别，再用 Go 构建信息覆盖 GOOS/GOARCH 并补充版本与 CGO 开关。
func (s *AnalyzerSrv) Analyze(ctx context.Context, param *filemodel.AnalyzeParam) (*filemodel.Result, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(param.Path)
	if err != nil {
		s.log.Error(ctx).Err(err).Str("path", param.Path).Msg("resolve absolute path failed")
		return nil, fmt.Errorf("resolve absolute path %s: %w", param.Path, err)
	}
	info, err := os.Stat(param.Path)
	if err != nil {
		s.log.Error(ctx).Err(err).Str("path", param.Path).Msg("stat file failed")
		return nil, fmt.Errorf("stat file: %w", err)
	}

	res := &filemodel.Result{Path: abs}
	if uint32(info.Mode().Perm())&consts.ExecPermMask != 0 {
		res.HasExecPerms = consts.TrueString
	}

	s.detectFormat(param.Path, res)
	s.fillGoBuildInfo(param.Path, res)
	return res.Serialize(), nil
}

// detectFormat 依次尝试 Mach-O fat、Mach-O、ELF、PE，命中即停；都不匹配则保持未知格式。
func (s *AnalyzerSrv) detectFormat(path string, res *filemodel.Result) {
	matched := s.detectMachOFat(path, res)
	if matched {
		return
	}
	matched = s.detectMachO(path, res)
	if matched {
		return
	}
	matched = s.detectELF(path, res)
	if matched {
		return
	}
	s.detectPE(path, res)
}

func (s *AnalyzerSrv) detectMachOFat(path string, res *filemodel.Result) bool {
	fat, err := macho.OpenFat(path)
	if err != nil {
		return false
	}
	defer fat.Close()

	infos := make([]archInfo, 0, len(fat.Arches))
	for _, a := range fat.Arches {
		infos = append(infos, s.machoArchInfo(a))
	}
	archs := make([]string, 0, len(infos))
	exec := false
	dyn := false
	for _, info := range infos {
		exec = exec || info.exec
		dyn = dyn || info.dyn
		archs = append(archs, info.arch)
	}
	archs = slices.DeleteFunc(archs, func(arch string) bool { return arch == "" })

	res.Format = consts.FormatMachOFat
	res.GOOS = consts.OSDarwin
	res.IsExecutable = boolString(exec)
	res.HasDynamicLibs = boolString(dyn)
	res.GOARCH = strings.Join(archs, ",")
	return true
}

func (s *AnalyzerSrv) machoArchInfo(a macho.FatArch) archInfo {
	if a.File == nil {
		return archInfo{}
	}
	// 读取导入库失败时按无动态依赖处理。
	libs, _ := a.File.ImportedLibraries()
	isExec := a.File.Type == macho.TypeExec || a.File.Type == macho.TypeDylib
	return archInfo{exec: isExec, dyn: len(libs) > 0, arch: mapMachOCpu(a.Cpu)}
}

func (s *AnalyzerSrv) detectMachO(path string, res *filemodel.Result) bool {
	f, err := macho.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// 读取导入库失败时按无动态依赖处理。
	libs, _ := f.ImportedLibraries()
	isExec := f.Type == macho.TypeExec || f.Type == macho.TypeDylib
	res.Format = consts.FormatMachO
	res.IsExecutable = boolString(isExec)
	res.HasDynamicLibs = boolString(len(libs) > 0)
	res.GOOS = consts.OSDarwin
	res.GOARCH = mapMachOCpu(f.Cpu)
	return true
}

func (s *AnalyzerSrv) detectELF(path string, res *filemodel.Result) bool {
	f, err := elf.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// 读取依赖库失败时按无动态依赖处理。
	libs, _ := f.ImportedLibraries()
	isExec := f.FileHeader.Type == elf.ET_EXEC || f.FileHeader.Type == elf.ET_DYN
	res.Format = consts.FormatELF
	res.IsExecutable = boolString(isExec)
	res.HasDynamicLibs = boolString(len(libs) > 0)
	res.GOOS = mapELFOSABI(f.FileHeader.OSABI)
	res.GOARCH = mapELFMachine(f.FileHeader.Machine)
	return true
}

func (s *AnalyzerSrv) detectPE(path string, res *filemodel.Result) bool {
	f, err := pe.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// 读取导入库失败时按无动态依赖处理。
	libs, _ := f.ImportedLibraries()
	res.Format = consts.FormatPE
	res.IsExecutable = consts.TrueString
	res.HasDynamicLibs = boolString(len(libs) > 0)
	res.GOOS = consts.OSWindows
	res.GOARCH = mapPEMachine(f.Machine)
	return true
}

// fillGoBuildInfo 读到 Go 构建信息时，用其中的 GOOS/GOARCH 覆盖推断值并补充版本与 CGO 开关。
func (s *AnalyzerSrv) fillGoBuildInfo(path string, res *filemodel.Result) {
	bi, err := buildinfo.ReadFile(path)
	// 非 Go 构建的文件没有构建信息，属于正常情况。
	if err != nil {
		return
	}
	res.IsGoBinary = consts.TrueString
	res.GoVersion = bi.GoVersion

	settings := make(map[string]string, len(bi.Settings))
	for _, setting := range bi.Settings {
		settings[setting.Key] = setting.Value
	}
	goos, hasGOOS := settings[consts.BuildKeyGOOS]
	if hasGOOS {
		res.GOOS = goos
	}
	goarch, hasGOARCH := settings[consts.BuildKeyGOARCH]
	if hasGOARCH {
		res.GOARCH = goarch
	}
	cgo, hasCGO := settings[consts.BuildKeyCGOEnabled]
	if hasCGO {
		res.CGOEnabled = cgo
	}
}

func boolString(value bool) string {
	if value {
		return consts.TrueString
	}
	return consts.FalseString
}

func mapMachOCpu(c macho.Cpu) string {
	switch c {
	case macho.CpuAmd64:
		return consts.ArchAMD64
	case macho.CpuArm64:
		return consts.ArchARM64
	case macho.Cpu386:
		return consts.Arch386
	case macho.CpuArm:
		return consts.ArchARM
	default:
		return c.String()
	}
}

func mapELFMachine(m elf.Machine) string {
	switch m {
	case elf.EM_X86_64:
		return consts.ArchAMD64
	case elf.EM_AARCH64:
		return consts.ArchARM64
	case elf.EM_386:
		return consts.Arch386
	case elf.EM_ARM:
		return consts.ArchARM
	default:
		return m.String()
	}
}

func mapELFOSABI(o elf.OSABI) string {
	switch o {
	case elf.ELFOSABI_FREEBSD:
		return consts.OSFreeBSD
	case elf.ELFOSABI_NETBSD:
		return consts.OSNetBSD
	case elf.ELFOSABI_SOLARIS:
		return consts.OSSolaris
	default:
		return consts.OSLinux
	}
}

func mapPEMachine(m uint16) string {
	switch m {
	case consts.PEMachineAMD64:
		return consts.ArchAMD64
	case consts.PEMachine386:
		return consts.Arch386
	case consts.PEMachineARM:
		return consts.ArchARM
	case consts.PEMachineARM64:
		return consts.ArchARM64
	default:
		return fmt.Sprintf("0x%x", m)
	}
}
