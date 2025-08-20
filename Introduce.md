# 短链接项目讲解

## 项目概述
这是一个基于 go-zero 微服务框架开发的短链接服务，主要功能是将长链接转换为短链接，并支持短链接的解析重定向。项目采用了现代化的微服务架构，具备高性能、高可用的特点。

## 技术架构

### 技术栈选择
    -框架 ：go-zero（阿里开源的微服务框架）
    -数据库 ：MySQL
    -缓存 ：Redis（分布式缓存 + 布隆过滤器）
    -编码算法：Base62
    -哈希算法 ：MD5

### 项目结构
```bash
shortener/
├── internal/          # 内部业务逻辑
│   ├── handler/       # HTTP处理器
│   ├── logic/         # 业务逻辑层
│   ├── svc/          # 服务上下文
│   └── types/        # 类型定义
├── model/            # 数据模型
├── sequence/         # 发号器实现
├── pkg/              # 工具包
└── etc/              # 配置文件
```

## 核心业务流程与源码解析
1. 长链接转短链接（Convert API）
    API定义shortener.api
    ```go
    type ConvertRequest {
        LongURL string `json:"longUrl" validate:"required"`
    }

    type ConvertResponse {
        ShortUrl string `json:"shortUrl"`
    }

    service shortener-api {
        @handler ConvertHandler
        post /convert (ConvertRequest) returns (ConvertResponse)
    }
    ```

    Handler层实现converthandler.go
    ```go
    func ConvertHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 1. 解析请求参数
            var req types.ConvertRequest
            if err := httpx.Parse(r, &req); err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
                return
            }
            
            // 2. 参数校验
            if err := validator.New().StructCtx(r.Context(), &req); err != nil {
                logx.Errorw("validator check failed", logx.LogField{Key: "err", Value: err.Error()})
                httpx.ErrorCtx(r.Context(), w, err)
                return
            }

            // 3. 执行业务逻辑
            l := logic.NewConvertLogic(r.Context(), svcCtx)
            resp, err := l.Convert(&req)
            if err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
            } else {
                httpx.OkJsonCtx(r.Context(), w, resp)
            }
        }
    }
    ```
    核心业务逻辑convertlogic.go
    ```go
    func (l *ConvertLogic) Convert(req *types.ConvertRequest) (resp *types.ConvertResponse, err error) {
        // 1. 网络连通性检测
        if ok := connect.Get(req.LongURL); !ok {
            return nil, errors.New("无效的链接")
        }

        // 2. 防重复转换：MD5哈希检查
        md5Value := md5.Sum([]byte(req.LongURL))
        u, err := l.svcCtx.ShortUrlModel.FindOneByMd5(l.ctx, sql.NullString{String: md5Value, Valid: true})
        if err != sqlx.ErrNotFound {
            if err == nil {
                return nil, fmt.Errorf("该链接已被转为%s", u.Surl.String)
            }
            return nil, err
        }

        // 3. 防循环转换：检查是否已经是短链接
        basePath, err := urltool.GetBasePath(req.LongURL)
        if err != nil {
            return nil, err
        }
        _, err = l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{String: basePath, Valid: true})
        if err != sqlx.ErrNotFound {
            if err == nil {
                return nil, errors.New("该链接已经是短链了")
            }
            return nil, err
        }

        var short string
        for {
            // 4. 获取唯一序列号
            seq, err := l.svcCtx.Sequence.Next()
            if err != nil {
                return nil, err
            }
            
            // 5. Base62编码生成短链接
            short = base62.Int2String(seq)
            
            // 6. 黑名单过滤
            if _, ok := l.svcCtx.ShortUrlBlackList[short]; !ok {
                break
            }
        }

        // 7. 存储映射关系
        if _, err := l.svcCtx.ShortUrlModel.Insert(l.ctx, &model.ShortUrlMap{
            Surl: sql.NullString{String: short, Valid: true},
            Lurl: sql.NullString{String: req.LongURL, Valid: true},
            Md5:  sql.NullString{String: md5Value, Valid: true},
        }); err != nil {
            return nil, err
        }

        // 8. 更新布隆过滤器
        _ = l.svcCtx.Filter.Add([]byte(short))

        return &types.ConvertResponse{ShortUrl: short}, nil
    }
    ```

2. 短链接解析（Show API）
    Handler层实现showhandler.go
    ```go
    func ShowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 解析参数
            var req types.ShowRequest
            if err := httpx.Parse(r, &req); err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
                return
            }

            // 执行业务逻辑
            l := logic.NewShowLogic(r.Context(), svcCtx)
            resp, err := l.Show(&req)
            if err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
            } else {
                // 关键：HTTP 302重定向
                http.Redirect(w, r, resp.LongUrl, http.StatusFound)
            }
        }
    }
    ```
    业务逻辑实现showlogic.go
    ```go
    func (l *ShowLogic) Show(req *types.ShowRequest) (resp *types.ShowResponse, err error) {
        // 1. 布隆过滤器快速判断
        exist, err := l.svcCtx.Filter.Exists([]byte(req.ShortUrl))
        if err != nil {
            logx.Errorw("Filter.Exists failed", logx.LogField{Value: err.Error(), Key: "err"})
        }

        // 不存在直接返回404
        if !exist {
            return nil, Err404
        }

        // 2. 查询数据库（带缓存）
        u, err := l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{Valid: true, String: req.ShortUrl})
        if err != nil {
            if err == sql.ErrNoRows {
                return nil, Err404
            }
            return nil, err
        }

        return &types.ShowResponse{LongUrl: u.Lurl.String}, nil
    }
    ```

## 核心技术亮点
### 1. MySQL发号器设计
```go
func (m *MySQL) Next() (uint64, error) {
    // 使用REPLACE INTO获取唯一序列号
    stmt, err := m.conn.Prepare(sqlReplaceIntoStub)
    if err != nil {
        return 0, err
    }
    defer stmt.Close()

    rest, err := stmt.Exec()
    if err != nil {
        return 0, err
    }

    // 获取自增ID作为序列号
    lid, err := rest.LastInsertId()
    if err != nil {
        return 0, err
    }
    return uint64(lid), nil
}
```
设计亮点 ：

- 利用MySQL的 REPLACE INTO 和自增ID特性
- 保证分布式环境下序列号的唯一性
- 支持高并发，每个连接的 LAST_INSERT_ID() 互不干扰

### 2. Base62编码算法(base62.go)
```go
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Int2String(num uint64) string {
    if num == 0 {
        return "0"
    }
    
    var result []byte
    for num > 0 {
        result = append([]byte{base62Chars[num%62]}, result...)
        num /= 62
    }
    return string(result)
}
```
优势 ：

- 62进制编码，字符集包含数字+大小写字母
- 生成的短链接更短、更美观
- 避免了特殊字符，URL友好

## 3. 多级缓存架构
布隆过滤器（第一级）
- 作用 ：快速判断短链接是否存在
- 实现 ：基于Redis的布隆过滤器
- 优势 ：20MB内存，支持千万级数据，误判率极低 
- 劣势 ：不支持删除操作，数据过期后无法自动清理

Redis缓存（第二级）
- 作用 ：缓存热点数据，减少数据库压力
- 实现 ：go-zero的 sqlc.CachedConn
- 特性 ：支持SingleFlight防击穿 

MySQL数据库（第三级）
- 作用 ：持久化存储
- 表设计 ：
```sql
CREATE TABLE `short_url_map` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `surl` varchar(10) DEFAULT NULL COMMENT '短链接',
  `lurl` varchar(1024) DEFAULT NULL COMMENT '长链接', 
  `md5` varchar(32) DEFAULT NULL COMMENT '长链接md5',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_surl` (`surl`),
  KEY `idx_md5` (`md5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 性能优化策略
### 1. 防缓存击穿
- 使用go-zero的SingleFlight机制
- 同一时间只有一个请求查询数据库
- 其他请求等待结果，避免数据库压力
### 2. 防恶意请求
- 布隆过滤器快速过滤不存在的短链接
- 减少99%的无效数据库查询
### 3. 黑名单机制
```go
// 避免生成敏感词短链接
if _, ok := l.svcCtx.ShortUrlBlackList[short]; !ok {
    break
}
```

### 4. 连接池优化
- 数据库连接池复用
- Redis连接池管理
