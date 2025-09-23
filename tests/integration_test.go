package tests

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	promptCmd "microdev/component/prompt/cmd"
	translateCmd "microdev/component/translate/cmd"
	"microdev/pkg/config"
	"microdev/pkg/input"
	"microdev/pkg/logger"
	"microdev/pkg/output"
	"microdev/pkg/utils"
)

// TestCommandIntegration 测试命令集成
func TestCommandIntegration(t *testing.T) {
	Convey("命令集成测试", t, func() {

		Convey("Prompt命令测试", func() {
			cmd := promptCmd.NewPromptCommand()
			So(cmd, ShouldNotBeNil)
			So(cmd.Use, ShouldEqual, "prompt <input>")
			So(cmd.Aliases, ShouldContain, "p")
			So(cmd.Short, ShouldEqual, "优化并生成更好的提示词")
		})

		Convey("Translate命令测试", func() {
			cmd := translateCmd.NewTranslateCommand()
			So(cmd, ShouldNotBeNil)
			So(cmd.Use, ShouldEqual, "translate <input>")
			So(cmd.Aliases, ShouldContain, "t")
			So(cmd.Short, ShouldEqual, "中英互译，支持文件、目录或直接文本")

			// 检查flags
			flags := cmd.Flags()
			So(flags.Lookup("directory"), ShouldNotBeNil)
			So(flags.Lookup("file"), ShouldNotBeNil)
			So(flags.Lookup("type"), ShouldNotBeNil)
			So(flags.Lookup("concurrency"), ShouldNotBeNil)
			So(flags.Lookup("force"), ShouldNotBeNil)
		})
	})
}

// TestInputProcessing 测试输入处理
func TestInputProcessing(t *testing.T) {
	Convey("输入处理测试", t, func() {
		log := logger.NewLogger()
		processor := input.NewProcessor(log)
		So(processor, ShouldNotBeNil)

		Convey("直接文本输入", func() {
			testInput := "这是一个测试文本"
			result, err := processor.Process(testInput)

			So(err, ShouldBeNil)
			So(result, ShouldEqual, testInput)
		})

		Convey("空输入处理", func() {
			result, err := processor.Process("")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeEmpty)
		})

		Convey("不存在的文件", func() {
			result, err := processor.Process("/path/to/nonexistent/file.txt")

			// 应该将其作为直接文本处理
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "/path/to/nonexistent/file.txt")
		})
	})
}

// TestOutputProcessing 测试输出处理
func TestOutputProcessing(t *testing.T) {
	Convey("输出处理测试", t, func() {
		log := logger.NewLogger()

		Convey("控制台输出", func() {
			processor := output.NewProcessor("", log)
			So(processor, ShouldNotBeNil)
			So(processor.IsConsoleOutput(), ShouldBeTrue)
			So(processor.GetOutputPath(), ShouldBeEmpty)

			// 测试写入（这会输出到stdout，但不会失败）
			n, err := processor.WriteString("测试输出")
			So(err, ShouldBeNil)
			So(n, ShouldEqual, len("测试输出"))

			err = processor.Close()
			So(err, ShouldBeNil)
		})

		Convey("文件输出", func() {
			// 使用临时文件
			tempFile := "/tmp/test_output.txt"
			defer os.Remove(tempFile) // 清理

			processor := output.NewProcessor(tempFile, log)
			So(processor, ShouldNotBeNil)
			So(processor.IsConsoleOutput(), ShouldBeFalse)
			So(processor.GetOutputPath(), ShouldEqual, tempFile)

			// 测试写入
			testContent := "这是测试内容"
			n, err := processor.WriteString(testContent)
			So(err, ShouldBeNil)
			So(n, ShouldEqual, len(testContent))

			// 刷新缓冲区
			err = processor.Flush()
			So(err, ShouldBeNil)

			// 关闭文件
			err = processor.Close()
			So(err, ShouldBeNil)

			// 验证文件内容
			content, err := os.ReadFile(tempFile)
			So(err, ShouldBeNil)
			So(string(content), ShouldEqual, testContent)
		})
	})
}

// TestLanguageDetection 测试语言检测
func TestLanguageDetection(t *testing.T) {
	Convey("语言检测测试", t, func() {

		Convey("中文文本检测", func() {
			testTexts := []string{
				"你好世界",
				"这是一个中文测试文本",
				"中英混合 mixed text 测试",
			}

			for _, text := range testTexts {
				direction := utils.DetectLanguage(text)
				So(direction, ShouldEqual, "zh2en")
			}
		})

		Convey("英文文本检测", func() {
			testTexts := []string{
				"Hello World",
				"This is an English test text",
				"Pure English content for testing",
			}

			for _, text := range testTexts {
				direction := utils.DetectLanguage(text)
				So(direction, ShouldEqual, "en2zh")
			}
		})

		Convey("空文本检测", func() {
			direction := utils.DetectLanguage("")
			So(direction, ShouldEqual, "zh2en") // 默认中译英
		})

		Convey("数字和符号", func() {
			direction := utils.DetectLanguage("123456!@#$%^&*()")
			So(direction, ShouldEqual, "zh2en") // 默认中译英
		})
	})
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	Convey("配置验证测试", t, func() {

		Convey("完整配置验证", func() {
			cfg := &config.Config{
				DashScopeAPIKey:  "test-key",
				AliBailianAPIKey: "test-bailian-key",
				ModelName:        "qwen-plus-latest",
				MaxTokens:        2048,
				Temperature:      0.7,
				StreamOutput:     true,
			}

			So(cfg.HasDashScopeKey(), ShouldBeTrue)
			So(cfg.HasBailianKey(), ShouldBeTrue)
			So(cfg.GetAPIKey(), ShouldEqual, "test-key")
			So(cfg.ModelName, ShouldNotBeEmpty)
			So(cfg.MaxTokens, ShouldBeGreaterThan, 0)
			So(cfg.Temperature, ShouldBeBetween, 0.0, 2.0)
		})

		Convey("配置边界值测试", func() {
			cfg := &config.Config{
				DashScopeAPIKey: "test-key",
				ModelName:       "test-model",
				MaxTokens:       1,
				Temperature:     0.0,
				StreamOutput:    false,
			}

			So(cfg.MaxTokens, ShouldEqual, 1)
			So(cfg.Temperature, ShouldEqual, 0.0)
			So(cfg.StreamOutput, ShouldBeFalse)
		})
	})
}

// TestEndToEndWorkflow 端到端工作流测试
func TestEndToEndWorkflow(t *testing.T) {
	Convey("端到端工作流测试", t, func() {

		// 检查是否有API密钥
		apiKey := os.Getenv("DASHSCOPE_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ALI_BAILIAN_API_KEY")
		}

		if apiKey == "" {
			t.Skip("跳过端到端测试：未设置API密钥")
			return
		}

		Convey("完整的配置加载流程", func() {
			cfg, err := config.LoadConfig()
			So(err, ShouldBeNil)
			So(cfg, ShouldNotBeNil)

			log := logger.NewLogger()
			So(log, ShouldNotBeNil)

			// 测试输入处理
			inputProcessor := input.NewProcessor(log)
			content, err := inputProcessor.Process("测试输入")
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "测试输入")

			// 测试输出处理
			outputProcessor := output.NewProcessor("", log)
			n, err := outputProcessor.Write([]byte("测试输出"))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, len("测试输出"))

			err = outputProcessor.Close()
			So(err, ShouldBeNil)
		})
	})
}
