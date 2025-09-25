package tests

import (
	"testing"
	"time"

	"microdev/component/prompt/session"
	"microdev/pkg/logger"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSessionManager(t *testing.T) {
	Convey("会话管理器测试", t, func() {
		log := logger.NewLogger()
		manager := session.NewManager(log)

		Convey("创建新会话", func() {
			sessionObj := manager.StartSession()
			So(sessionObj, ShouldNotBeNil)
			So(sessionObj.ID, ShouldNotBeEmpty)
			So(len(sessionObj.Messages), ShouldEqual, 0)
			So(manager.GetCurrentSession(), ShouldEqual, sessionObj)
		})

		Convey("添加消息", func() {
			manager.StartSession()
			
			// 添加用户消息
			manager.AddUserMessage("测试用户消息")
			sessionObj := manager.GetCurrentSession()
			So(len(sessionObj.Messages), ShouldEqual, 1)
			So(sessionObj.Messages[0].Role, ShouldEqual, "user")
			So(sessionObj.Messages[0].Content, ShouldEqual, "测试用户消息")
			
			// 添加助手消息
			manager.AddAssistantMessage("测试助手回复")
			So(len(sessionObj.Messages), ShouldEqual, 2)
			So(sessionObj.Messages[1].Role, ShouldEqual, "assistant")
			So(sessionObj.Messages[1].Content, ShouldEqual, "测试助手回复")
		})

		Convey("获取上下文信息", func() {
			manager.StartSession()
			
			// 空会话应该返回空上下文
			context := manager.GetContextForOptimization()
			So(context, ShouldEqual, "")
			
			// 添加一些消息
			manager.AddUserMessage("第一个问题")
			manager.AddAssistantMessage("第一个回答")
			manager.AddUserMessage("第二个问题")
			manager.AddAssistantMessage("第二个回答")
			
			context = manager.GetContextForOptimization()
			So(context, ShouldNotBeEmpty)
			So(context, ShouldContainSubstring, "最近对话")
			So(context, ShouldContainSubstring, "第一个问题")
			So(context, ShouldContainSubstring, "第二个问题")
		})

		Convey("记忆压缩", func() {
			manager.StartSession()
			
			// 添加超过最大消息数的消息
			for i := 0; i < 10; i++ {
				manager.AddUserMessage("用户消息 " + string(rune(i+'1')))
				manager.AddAssistantMessage("助手回复 " + string(rune(i+'1')))
			}
			
			sessionObj := manager.GetCurrentSession()
			// 应该触发记忆压缩
			So(len(sessionObj.Messages), ShouldBeLessThan, 20)
			So(sessionObj.Summary, ShouldNotBeEmpty)
		})

		Convey("会话统计", func() {
			manager.StartSession()
			
			// 空会话统计
			stats := manager.GetSessionStats()
			So(stats["session_active"], ShouldBeTrue)
			So(stats["total_messages"], ShouldEqual, 0)
			So(stats["user_messages"], ShouldEqual, 0)
			So(stats["assistant_messages"], ShouldEqual, 0)
			So(stats["has_summary"], ShouldBeFalse)
			
			// 添加消息后的统计
			manager.AddUserMessage("测试")
			manager.AddAssistantMessage("回复")
			
			stats = manager.GetSessionStats()
			So(stats["total_messages"], ShouldEqual, 2)
			So(stats["user_messages"], ShouldEqual, 1)
			So(stats["assistant_messages"], ShouldEqual, 1)
		})

		Convey("结束会话", func() {
			manager.StartSession()
			So(manager.GetCurrentSession(), ShouldNotBeNil)
			
			manager.EndSession()
			So(manager.GetCurrentSession(), ShouldBeNil)
			
			// 结束后的统计
			stats := manager.GetSessionStats()
			So(stats["session_active"], ShouldBeFalse)
		})
	})
}

func TestMessage(t *testing.T) {
	Convey("消息结构测试", t, func() {
		message := session.Message{
			Role:      "user",
			Content:   "测试内容",
			Timestamp: time.Now(),
		}
		
		So(message.Role, ShouldEqual, "user")
		So(message.Content, ShouldEqual, "测试内容")
		So(message.Timestamp, ShouldNotBeZeroValue)
	})
}
