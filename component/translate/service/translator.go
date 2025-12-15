package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"microdev/component/translate/model"
	"microdev/pkg/config"
	"microdev/pkg/llm"
	"microdev/pkg/logger"
)

// TranslatorService 翻译服务
type TranslatorService struct {
	llm    *llm.Client
	cfg    *config.Config
	logger *logger.Logger
}

// NewTranslatorService 创建新的翻译服务
func NewTranslatorService(cli *llm.Client, cfg *config.Config) *TranslatorService {
	log := logger.NewLogger(logger.WithModule("translate", ""))
	ss := &TranslatorService{
		llm:    cli,
		cfg:    cfg,
		logger: log,
	}
	return ss
}

// Translate 读取req.Input,调用llm进行翻译，并把翻译结果写入req.Output
func (s *TranslatorService) Translate(req *model.Translate) error {
	if req == nil || req.Input == nil || req.Output == nil {
		return fmt.Errorf("invalid translate request")
	}
	// 读取输入
	defer req.Input.Close()
	data, err := io.ReadAll(req.Input)
	if err != nil {
		s.logger.Err(err).Msg("read input failed")
		return err
	}
	text := string(data)
	if text == "" {
		return io.EOF
	}

	// 判断Direct
	direct := req.Direct
	if direct == "" {
		direct = detectLang(text)
	}

	// 分块翻译并写入输出（支持大文本与流式）
	chunks := chunkText(text, 2000)
	for _, c := range chunks {
		if err := s.llm.TranslateText(c, direct, req.Output); err != nil {
			s.logger.Err(err).Msg("translate chunk failed")
			return err
		}
	}

	return nil
}

// 读出res.Out,把结果写入命令行中
func (s *TranslatorService) Output(res *model.Translate) error {
	if res == nil || res.Output == nil {
		return fmt.Errorf("invalid translate result")
	}
	defer res.Output.Close()

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	_, err := io.Copy(w, res.Output)
	if err != nil {
		s.logger.Err(err).Msg("write output failed")
		return err
	}
	return nil
}

// Input 处理输入参数为 ReadCloser，支持文件、stdin、直接文本
func (s *TranslatorService) Input(in string) (io.ReadCloser, error) {
	// 空输入直接报错
	if strings.TrimSpace(in) == "" {
		return nil, fmt.Errorf("empty input")
	}
	// 使用 - 代表从stdin读取
	if in == "-" {
		s.logger.Info().Msg("read from stdin")
		return io.NopCloser(os.Stdin), nil
	}
	// 尝试作为文件路径
	if fi, err := os.Stat(in); err == nil && fi.Mode().IsRegular() {
		f, err := os.Open(in)
		if err != nil {
			s.logger.Err(err).Str("file", in).Msg("open file failed")
			return nil, err
		}
		s.logger.Info().Str("file", filepath.Base(in)).Int64("size", fi.Size()).Msg("read from file")
		return f, nil
	}
	// 否则作为直接文本
	s.logger.Info().Int("text_len", len(in)).Msg("read from text")
	return io.NopCloser(strings.NewReader(in)), nil
}

// detectLang 检测语言方向
func detectLang(s string) string {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return "zh2en"
		}
	}
	return "en2zh"
}

// chunkText 按 rune 分块
func chunkText(s string, size int) []string {
	if size <= 0 {
		size = 2000
	}
	out := make([]string, 0, (len(s)/size)+1)
	runes := []rune(s)
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}
