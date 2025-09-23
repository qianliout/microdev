package tests

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
)

// TestBasicFunctionality 测试基础功能
func TestBasicFunctionality(t *testing.T) {
	Convey("基础功能测试", t, func() {

		Convey("日志器测试", func() {
			log := logger.NewLogger()
			So(log, ShouldNotBeNil)

			// 测试日志方法不会panic
			So(func() { log.Info().Msg("测试信息") }, ShouldNotPanic)
			So(func() { log.Error().Msg("测试错误") }, ShouldNotPanic)
			So(func() { log.Debug().Msg("测试调试") }, ShouldNotPanic)
		})

		Convey("错误处理测试", func() {
			Convey("配置错误", func() {
				err := errors.NewConfigError("测试配置错误", nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "测试配置错误")

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "配置错误")
			})

			Convey("LLM错误", func() {
				err := errors.NewLLMError("测试LLM错误", nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "测试LLM错误")

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "AI模型错误")
			})

			Convey("输入错误", func() {
				err := errors.NewInputError("测试输入错误", nil)
				So(err, ShouldNotBeNil)

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "输入错误")
			})

			Convey("输出错误", func() {
				err := errors.NewOutputError("测试输出错误", nil)
				So(err, ShouldNotBeNil)

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "输出错误")
			})

			Convey("网络错误", func() {
				err := errors.NewNetworkError("测试网络错误", nil)
				So(err, ShouldNotBeNil)

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "网络错误")
			})

			Convey("文件错误", func() {
				err := errors.NewFileError("测试文件错误", nil)
				So(err, ShouldNotBeNil)

				friendlyMsg := errors.GetUserFriendlyMessage(err)
				So(friendlyMsg, ShouldContainSubstring, "文件操作错误")
			})
		})

		Convey("配置管理测试", func() {
			Convey("默认配置值", func() {
				// 设置一个临时的API密钥以通过验证
				os.Setenv("DASHSCOPE_API_KEY", "temp-key")
				defer os.Unsetenv("DASHSCOPE_API_KEY")

				cfg, err := config.LoadConfig()
				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)

				// 检查默认值
				So(cfg.ModelName, ShouldEqual, "qwen-plus-latest")
				So(cfg.MaxTokens, ShouldEqual, 2048)
				So(cfg.Temperature, ShouldEqual, 0.7)
				So(cfg.StreamOutput, ShouldBeTrue)
			})

			Convey("API密钥检查方法", func() {
				cfg := &config.Config{
					DashScopeAPIKey:  "dash-key",
					AliBailianAPIKey: "bailian-key",
				}

				So(cfg.HasDashScopeKey(), ShouldBeTrue)
				So(cfg.HasBailianKey(), ShouldBeTrue)
				So(cfg.GetAPIKey(), ShouldEqual, "dash-key") // 优先返回DashScope

				cfg.DashScopeAPIKey = ""
				So(cfg.HasDashScopeKey(), ShouldBeFalse)
				So(cfg.GetAPIKey(), ShouldEqual, "bailian-key") // 返回百炼
			})

			Convey("环境变量读取", func() {
				// 测试DashScope API密钥
				os.Setenv("DASHSCOPE_API_KEY", "test-dashscope-key")
				os.Unsetenv("ALI_BAILIAN_API_KEY")

				cfg, err := config.LoadConfig()
				So(err, ShouldBeNil)
				So(cfg.DashScopeAPIKey, ShouldEqual, "test-dashscope-key")
				So(cfg.AliBailianAPIKey, ShouldBeEmpty)

				// 清理
				os.Unsetenv("DASHSCOPE_API_KEY")

				// 测试百炼API密钥
				os.Setenv("ALI_BAILIAN_API_KEY", "test-bailian-key")

				cfg, err = config.LoadConfig()
				So(err, ShouldBeNil)
				So(cfg.DashScopeAPIKey, ShouldBeEmpty)
				So(cfg.AliBailianAPIKey, ShouldEqual, "test-bailian-key")

				// 清理
				os.Unsetenv("ALI_BAILIAN_API_KEY")
			})
		})
	})
}

// TestEnvironmentSetup 测试环境设置
func TestEnvironmentSetup(t *testing.T) {
	Convey("环境设置测试", t, func() {

		Convey("检查必要的依赖包", func() {
			// 这个测试确保我们能够导入所有必要的包
			So(func() {
				_ = logger.NewLogger()
			}, ShouldNotPanic)

			So(func() {
				_, _ = config.LoadConfig()
			}, ShouldNotPanic)
		})

		Convey("检查API密钥环境变量", func() {
			dashScopeKey := os.Getenv("DASHSCOPE_API_KEY")
			bailianKey := os.Getenv("ALI_BAILIAN_API_KEY")

			if dashScopeKey != "" {
				t.Logf("检测到DASHSCOPE_API_KEY: %s...", dashScopeKey[:min(len(dashScopeKey), 10)])
			}

			if bailianKey != "" {
				t.Logf("检测到ALI_BAILIAN_API_KEY: %s...", bailianKey[:min(len(bailianKey), 10)])
			}

			if dashScopeKey == "" && bailianKey == "" {
				t.Log("警告: 未检测到API密钥环境变量，真实连接测试将被跳过")
			}

			// 这个测试总是通过，只是用来显示环境信息
			So(true, ShouldBeTrue)
		})
	})
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
