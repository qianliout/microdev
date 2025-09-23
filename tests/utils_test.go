package tests

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"microdev/pkg/utils"
)

// TestUtilsFunctionality 测试工具函数
func TestUtilsFunctionality(t *testing.T) {
	Convey("工具函数测试", t, func() {

		Convey("语言检测功能", func() {
			Convey("中文文本检测", func() {
				text := "这是一个中文测试文本"
				direction := utils.DetectLanguage(text)
				So(direction, ShouldEqual, "zh2en")
			})

			Convey("英文文本检测", func() {
				text := "This is an English test text"
				direction := utils.DetectLanguage(text)
				So(direction, ShouldEqual, "en2zh")
			})
		})

		Convey("输出路径生成", func() {
			Convey("中文内容生成英文输出路径", func() {
				inputPath := "/path/to/document.md"
				content := "这是中文内容"
				outputPath := utils.GetTranslationOutputPath(inputPath, content)
				expected := "/path/to/document_en.md"
				So(outputPath, ShouldEqual, expected)
			})

			Convey("英文内容生成中文输出路径", func() {
				inputPath := "/path/to/document.md"
				content := "This is English content"
				outputPath := utils.GetTranslationOutputPath(inputPath, content)
				expected := "/path/to/document_zh.md"
				So(outputPath, ShouldEqual, expected)
			})

			Convey("不同扩展名处理", func() {
				inputPath := "/path/to/file.txt"
				content := "English content"
				outputPath := utils.GetTranslationOutputPath(inputPath, content)
				expected := "/path/to/file_zh.txt"
				So(outputPath, ShouldEqual, expected)
			})
		})

		Convey("翻译文件存在性检查", func() {
			// 创建临时目录
			tempDir, err := os.MkdirTemp("", "micro_test")
			So(err, ShouldBeNil)
			defer os.RemoveAll(tempDir)

			Convey("文件不存在时返回false", func() {
				inputPath := filepath.Join(tempDir, "test.md")
				content := "English content"
				exists := utils.IsTranslationExists(inputPath, content)
				So(exists, ShouldBeFalse)
			})

			Convey("文件存在时返回true", func() {
				inputPath := filepath.Join(tempDir, "test.md")
				content := "English content"
				outputPath := utils.GetTranslationOutputPath(inputPath, content)
				
				// 创建输出文件
				err := os.WriteFile(outputPath, []byte("translated content"), 0644)
				So(err, ShouldBeNil)
				
				exists := utils.IsTranslationExists(inputPath, content)
				So(exists, ShouldBeTrue)
			})
		})
	})
}
