package connect

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	c.Convey("基础用例", t, func() {
		url := "https://www.liwenzhou.com/posts/Go/unit-test-5/"
		got := Get(url)
		// 断言
		c.So(got, c.ShouldEqual, true) // 断言
	})

	c.Convey("url请求不通的示例", t, func() {
		url := "posts/Go/unit-test-5/"
		got := Get(url)
		// 断言
		c.ShouldBeFalse(got)
	})

	// 添加更多测试用例
	c.Convey("无效URL测试", t, func() {
		url := "invalid-url"
		got := Get(url)
		c.So(got, c.ShouldBeFalse)
	})

	c.Convey("空URL测试", t, func() {
		url := ""
		got := Get(url)
		c.So(got, c.ShouldBeFalse)
	})

	c.Convey("不存在的域名测试", t, func() {
		url := "https://nonexistentdomain12345.com"
		got := Get(url)
		c.So(got, c.ShouldBeFalse)
	})

	c.Convey("HTTP 404测试", t, func() {
		url := "https://httpbin.org/status/404"
		got := Get(url)
		c.So(got, c.ShouldBeFalse) // 404状态码应该返回false
	})

	c.Convey("HTTP 200测试", t, func() {
		url := "https://httpbin.org/status/200"
		got := Get(url)
		c.So(got, c.ShouldBeTrue) // 200状态码应该返回true
	})
}

// func TestGet(t *testing.T) {
// 	c.Convey("基础用例", t, func() {

// 		got := Get(s, sep)
// 		// 断言
// 		c.So(got, c.ShouldResemble, expect)
// 	})
// }
