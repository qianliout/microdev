package api

import (
	"bufio"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"microdev/component/prompt/service"
	"microdev/infra/console"
	"microdev/infra/consts"
	promptmodel "microdev/infra/model/prompt"
)

const separatorWidth = 60

type interactiveRunner struct {
	cmd        *cobra.Command
	optimizer  *service.OptimizerSrv
	sessionSrv *service.SessionSrv
	out        *console.Console
	sessionID  int64
}

func newInteractiveRunner(
	cmd *cobra.Command,
	optimizer *service.OptimizerSrv,
	sessionSrv *service.SessionSrv,
) *interactiveRunner {
	return &interactiveRunner{
		cmd:        cmd,
		optimizer:  optimizer,
		sessionSrv: sessionSrv,
		out:        console.New(),
	}
}

// run 创建会话后逐行读取输入并优化，输入退出命令或标准输入结束时返回。
func (vi *interactiveRunner) run(ctx context.Context) error {
	session, err := vi.sessionSrv.CreateSession(ctx, &promptmodel.CreateSessionParam{})
	if err != nil {
		return err
	}
	vi.sessionID = session.ID

	vi.out.Println("🤖 欢迎使用提示词优化交互模式！")
	vi.out.Println("💡 特性：保持对话历史、智能记忆压缩、上下文感知优化")
	vi.out.Println("📝 输入您的提示词，我将为您优化。输入 'exit'、'quit' 或 '退出' 结束会话。")
	vi.out.Println(strings.Repeat("-", separatorWidth))
	vi.prompt()

	scanner := bufio.NewScanner(vi.cmd.InOrStdin())
	done := false
	for !done && scanner.Scan() {
		done = vi.handleLine(ctx, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	return nil
}

// handleLine 处理一行输入，返回 true 表示应结束会话。
func (vi *interactiveRunner) handleLine(ctx context.Context, line string) bool {
	input := strings.TrimSpace(line)
	if isExitCommand(input) {
		vi.out.Println("\n👋 感谢使用！会话已结束。")
		return true
	}
	if input == "" {
		vi.out.Println("⚠️  输入不能为空，请重新输入。")
		vi.prompt()
		return false
	}

	vi.out.Println("\n🔄 正在优化提示词...")
	param := &promptmodel.OptimizeWithSessionParam{
		SessionID:      vi.sessionID,
		OriginalPrompt: input,
		Language:       consts.DefaultPromptLanguage,
		Output:         vi.cmd.OutOrStdout(),
	}
	_, err := vi.optimizer.OptimizeWithSession(ctx, param)
	vi.out.Println("")
	if err != nil {
		vi.out.Printf("❌ 优化失败: %s\n", err.Error())
	}
	vi.printStats(ctx)
	vi.out.Println(strings.Repeat("-", separatorWidth))
	vi.prompt()
	return false
}

func (vi *interactiveRunner) printStats(ctx context.Context) {
	stats, err := vi.sessionSrv.GetSessionStats(ctx, &promptmodel.SessionIDParam{SessionID: vi.sessionID})
	if err != nil {
		return
	}
	if stats.SessionActive != consts.TrueString {
		return
	}
	msg := fmt.Sprintf("\n📊 会话统计: %d条消息", stats.TotalMessages)
	if stats.HasSummary == consts.TrueString {
		msg += " (已压缩记忆)"
	}
	vi.out.Println(msg)
}

func (vi *interactiveRunner) prompt() {
	vi.out.Println("\n💬 请输入提示词: ")
}

func isExitCommand(input string) bool {
	return slices.Contains(consts.ExitCommands, strings.ToLower(input))
}
