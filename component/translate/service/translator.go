package service

import (
	"context"

	"microdev/infra/consts"
	"microdev/infra/llm"
	"microdev/infra/logger"
	translatemodel "microdev/infra/model/translate"
	"microdev/infra/utils"
)

type TranslatorSrv struct {
	llmCli *llm.Client
	log    *logger.Logger
}

func NewTranslatorSrv(llmCli *llm.Client) *TranslatorSrv {
	return &TranslatorSrv{
		llmCli: llmCli,
		log:    logger.New(logger.WithModule("translate"), logger.WithSubModule("service")),
	}
}

// Translate 按块调用大模型翻译 param.Text，译文流式写入 param.Output。
func (s *TranslatorSrv) Translate(ctx context.Context, param *translatemodel.TranslateParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	direction := param.Direction
	if direction == "" {
		direction = detectDirection(param.Text)
	}

	chunks := chunkText(param.Text, consts.TranslateChunkSize)
	for i, chunk := range chunks {
		chunkParam := &llm.TranslateTextParam{Text: chunk, Direction: direction, Output: param.Output}
		if err := s.llmCli.TranslateText(ctx, chunkParam); err != nil {
			s.log.Error(ctx).Err(err).
				Str("direction", direction).
				Int("chunkIndex", i).
				Int("chunkCount", len(chunks)).
				Msg("translate chunk failed")
			return err
		}
	}
	return nil
}

// detectDirection 文本含任意汉字时按中译英处理，否则按英译中。
func detectDirection(text string) string {
	if utils.HasChinese(text) {
		return consts.DirectionZh2En
	}
	return consts.DirectionEn2Zh
}

// chunkText 按 rune 数量切块，size 非正时使用默认块大小。
func chunkText(text string, size int64) []string {
	if size <= 0 {
		size = consts.TranslateChunkSize
	}
	runes := []rune(text)
	step := int(size)
	chunks := make([]string, 0, len(runes)/step+1)
	for start := 0; start < len(runes); start += step {
		end := min(start+step, len(runes))
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}
