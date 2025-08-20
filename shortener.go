package main

import (
	"flag"
	"fmt"

	"shortener/internal/config"
	"shortener/internal/handler"
	"shortener/internal/svc"
	"shortener/pkg/base62"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/shortener-api.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)
	fmt.Printf("load conf::%#v\n", c) // 修正：添加缺失的引号和参数

	// 2. 初始化Base62模块
	base62.MustInit(c.BaseString)

	// 3. 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 4. 初始化服务上下文
	ctx := svc.NewServiceContext(c)
	// 5. 注册路由处理器
	handler.RegisterHandlers(server, ctx)

	// 6. 启动服务
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
