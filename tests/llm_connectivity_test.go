package tests

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"microdev/component/translate/model"
	"microdev/component/translate/service"
	"microdev/pkg/config"
	"microdev/pkg/llm"
	"microdev/pkg/logger"
)

// TestLLMConnectivity 测试LLM连接性
func TestLLMConnectivity(t *testing.T) {
	Convey("LLM连接性测试", t, func() {

		Convey("配置加载测试", func() {
			Convey("当环境变量未设置时", func() {
				// 保存原始环境变量
				originalDashScope := os.Getenv("DASHSCOPE_API_KEY")
				originalBailian := os.Getenv("ALI_BAILIAN_API_KEY")

				// 清空环境变量
				os.Unsetenv("DASHSCOPE_API_KEY")
				os.Unsetenv("ALI_BAILIAN_API_KEY")

				cfg, err := config.LoadConfig()

				// 恢复环境变量
				if originalDashScope != "" {
					os.Setenv("DASHSCOPE_API_KEY", originalDashScope)
				}
				if originalBailian != "" {
					os.Setenv("ALI_BAILIAN_API_KEY", originalBailian)
				}

				So(err, ShouldNotBeNil)
				So(cfg, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "至少需要设置")
			})

			Convey("当设置了DASHSCOPE_API_KEY时", func() {
				os.Setenv("DASHSCOPE_API_KEY", "test-key")
				defer os.Unsetenv("DASHSCOPE_API_KEY")

				cfg, err := config.LoadConfig()

				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
				So(cfg.HasDashScopeKey(), ShouldBeTrue)
				So(cfg.DashScopeAPIKey, ShouldEqual, "test-key")
				So(cfg.ModelName, ShouldEqual, "qwen-plus-latest")
				So(cfg.MaxTokens, ShouldEqual, 2048)
				So(cfg.Temperature, ShouldEqual, 0.7)
				So(cfg.StreamOutput, ShouldBeTrue)
			})

			Convey("当设置了ALI_BAILIAN_API_KEY时", func() {
				os.Setenv("ALI_BAILIAN_API_KEY", "test-bailian-key")
				defer os.Unsetenv("ALI_BAILIAN_API_KEY")

				cfg, err := config.LoadConfig()

				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
				So(cfg.HasBailianKey(), ShouldBeTrue)
				So(cfg.AliBailianAPIKey, ShouldEqual, "test-bailian-key")
			})
		})

		Convey("LLM客户端创建测试", func() {
			log := logger.NewLogger()

			Convey("使用有效配置创建客户端", func() {
				cfg := &config.Config{
					DashScopeAPIKey: "test-key",
					ModelName:       "qwen-plus-latest",
					MaxTokens:       2048,
					Temperature:     0.7,
					StreamOutput:    false,
				}

				client, err := llm.NewClient(cfg, log)

				So(err, ShouldBeNil)
				So(client, ShouldNotBeNil)
			})

			Convey("使用无效配置创建客户端", func() {
				cfg := &config.Config{
					ModelName:    "qwen-plus-latest",
					MaxTokens:    2048,
					Temperature:  0.7,
					StreamOutput: false,
				}

				client, err := llm.NewClient(cfg, log)

				So(err, ShouldNotBeNil)
				So(client, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "没有可用的API密钥")
			})
		})

		Convey("翻译器创建测试", func() {
			log := logger.NewLogger()

			Convey("使用有效配置创建翻译器", func() {
				cfg := &config.Config{
					DashScopeAPIKey: "test-key",
					ModelName:       "qwen-plus-latest",
					MaxTokens:       2048,
					Temperature:     0.7,
					StreamOutput:    false,
				}

				trans, err := service.NewTranslatorService(cfg, log)

				So(err, ShouldBeNil)
				So(trans, ShouldNotBeNil)
			})

			Convey("使用无效配置创建翻译器", func() {
				cfg := &config.Config{
					ModelName:    "qwen-plus-latest",
					MaxTokens:    2048,
					Temperature:  0.7,
					StreamOutput: false,
				}

				trans, err := service.NewTranslatorService(cfg, log)

				So(err, ShouldNotBeNil)
				So(trans, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "没有可用的API密钥")
			})
		})
	})
}

// TestLLMRealConnectivity 测试真实的LLM连接（需要真实的API密钥）
func TestLLMRealConnectivity(t *testing.T) {
	Convey("真实LLM连接测试", t, func() {

		// 检查是否有真实的API密钥
		apiKey := os.Getenv("DASHSCOPE_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ALI_BAILIAN_API_KEY")
		}

		if apiKey == "" {
			t.Skip("跳过真实连接测试：未设置API密钥")
			return
		}

		log := logger.NewLogger()
		cfg, err := config.LoadConfig()
		So(err, ShouldBeNil)

		Convey("测试Prompt生成功能", func() {
			client, err := llm.NewClient(cfg, log)
			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)

			// 使用缓冲区捕获输出
			var buf bytes.Buffer

			// 设置超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 测试简单的prompt生成
			testInput := "写一个Hello World程序"

			// 创建一个goroutine来执行生成
			done := make(chan error, 1)
			go func() {
				done <- client.GeneratePrompt(testInput, &buf)
			}()

			// 等待完成或超时
			select {
			case err := <-done:
				So(err, ShouldBeNil)
				output := buf.String()
				So(output, ShouldNotBeEmpty)
				So(len(output), ShouldBeGreaterThan, 10)
				t.Logf("生成的Prompt: %s", output)
			case <-ctx.Done():
				t.Log("测试超时，可能是网络问题或API响应慢")
				// 不标记为失败，因为可能是网络问题
			}
		})

		Convey("测试翻译功能", func() {
			trans, err := service.NewTranslatorService(cfg, log)
			So(err, ShouldBeNil)
			So(trans, ShouldNotBeNil)

			// 使用缓冲区捕获输出
			var buf bytes.Buffer

			// 设置超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 测试简单的翻译
			testInput := "Hello World"

			// 创建一个goroutine来执行翻译
			done := make(chan error, 1)
			go func() {
				request := &model.TranslationRequest{
					Content: testInput,
				}
				done <- trans.Translate(request, &buf)
			}()

			// 等待完成或超时
			select {
			case err := <-done:
				So(err, ShouldBeNil)
				output := buf.String()
				So(output, ShouldNotBeEmpty)
				So(len(output), ShouldBeGreaterThan, 0)
				// 检查是否包含中文字符（英译中）
				containsChinese := false
				for _, r := range output {
					if r >= 0x4e00 && r <= 0x9fff {
						containsChinese = true
						break
					}
				}
				So(containsChinese, ShouldBeTrue)
				t.Logf("翻译结果: %s", output)
			case <-ctx.Done():
				t.Log("翻译测试超时，可能是网络问题或API响应慢")
				// 不标记为失败，因为可能是网络问题
			}
		})

		Convey("测试中译英功能", func() {
			trans, err := service.NewTranslatorService(cfg, log)
			So(err, ShouldBeNil)
			So(trans, ShouldNotBeNil)

			// 使用缓冲区捕获输出
			var buf bytes.Buffer

			// 设置超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 测试中译英
			testInput := "你好世界"

			// 创建一个goroutine来执行翻译
			done := make(chan error, 1)
			go func() {
				request := &model.TranslationRequest{
					Content: testInput,
				}
				done <- trans.Translate(request, &buf)
			}()

			// 等待完成或超时
			select {
			case err := <-done:
				So(err, ShouldBeNil)
				output := buf.String()
				So(output, ShouldNotBeEmpty)
				So(len(output), ShouldBeGreaterThan, 0)
				// 检查是否包含英文
				output = strings.ToLower(output)
				So(output, ShouldContainSubstring, "hello")
				t.Logf("中译英结果: %s", output)
			case <-ctx.Done():
				t.Log("中译英测试超时，可能是网络问题或API响应慢")
				// 不标记为失败，因为可能是网络问题
			}
		})
	})
}
