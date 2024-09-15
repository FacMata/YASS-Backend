package main

import (
        "crypto/md5"
        "encoding/hex"
        "fmt"
        "io"
        "net/http"
        "os"
        "strconv"
        "strings"
        "sync"

        "github.com/gin-gonic/gin"
        "github.com/spf13/viper"
)

// 定义全局的缓冲池
var bufferPool = sync.Pool{
        New: func() interface{} {
                // 每个缓冲区大小为 4MB
                return make([]byte, 4*1024*1024) // 4MB
        },
}

var (
        remote_token string
        mount_dir    string
        port         string
)

// 处理 HTTP Range 请求，支持按需加载
func remote(c *gin.Context) {
        MediaSourceId := c.Query("MediaSourceId")
        dir := c.Query("dir")
        key := c.Query("key")

        // 鉴权逻辑
        raw_string := "dir=" + dir + "&MediaSourceId=" + MediaSourceId + "&remote_token=" + remote_token
        hash_1 := md5.Sum([]byte(raw_string))
        hash := hex.EncodeToString(hash_1[:])
        if key != hash {
                c.AbortWithStatus(403)
                return
        }

        // 文件路径
        local_dir := mount_dir + dir
        file, err := os.Open(local_dir)
        if err != nil {
                c.AbortWithStatus(404)
                return
        }
        defer file.Close()

        // 获取文件信息
        fileInfo, err := file.Stat()
        if err != nil {
                c.AbortWithStatus(500)
                return
        }
        fileSize := fileInfo.Size()

        // 解析 Range 请求头
        rangeHeader := c.GetHeader("Range")
        if rangeHeader == "" {
                // 如果没有 Range 头，返回整个文件
                c.Writer.Header().Set("Content-Type", "video/mp4")
                c.Writer.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
                c.Status(http.StatusOK)

                // 分片传输整个文件
                streamFile(file, c, 0, fileSize-1)
                return
        }

        // Range 请求的格式: bytes=START-END
        ranges := strings.Split(rangeHeader, "=")
        if len(ranges) != 2 || ranges[0] != "bytes" {
                c.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
                return
        }
        rangeParts := strings.Split(ranges[1], "-")

        start, err := strconv.ParseInt(rangeParts[0], 10, 64)
        if err != nil || start < 0 || start >= fileSize {
                c.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
                return
        }

        var end int64
        if rangeParts[1] != "" {
                end, err = strconv.ParseInt(rangeParts[1], 10, 64)
                if err != nil || end >= fileSize || end < start {
                        c.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
                        return
                }
        } else {
                end = fileSize - 1
        }

        // 设置响应头信息
        contentLength := end - start + 1
        c.Writer.Header().Set("Content-Type", "video/mp4")
        c.Writer.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
        c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
        c.Status(http.StatusPartialContent)

        // 分片传输文件的指定范围
        streamFile(file, c, start, end)
}

// 分片传输文件的指定部分
func streamFile(file *os.File, c *gin.Context, start, end int64) {
        // 移动文件指针到指定位置
        _, err := file.Seek(start, 0)
        if err != nil {
                c.AbortWithStatus(500)
                return
        }

        // 使用缓冲区按块读取并传输
        var wg sync.WaitGroup
        wg.Add(1)

        go func() {
                defer wg.Done()
                buffer := bufferPool.Get().([]byte)
                defer bufferPool.Put(buffer)

                totalBytes := end - start + 1
                for totalBytes > 0 {
                        readSize := int64(len(buffer))
                        if totalBytes < readSize {
                                readSize = totalBytes
                        }

                        n, err := file.Read(buffer[:readSize])
                        if err != nil && err != io.EOF {
                                c.AbortWithStatus(500)
                                return
                        }

                        if n == 0 {
                                break
                        }

                        _, writeErr := c.Writer.Write(buffer[:n])
                        if writeErr != nil {
                                // 如果客户端断开连接
                                fmt.Println("Client connection lost:", writeErr)
                                return
                        }

                        c.Writer.Flush() // 刷新数据
                        totalBytes -= int64(n)
                }
        }()

        wg.Wait()
}

// 定义中间件处理跨域请求
func corsMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
                c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
                c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                if c.Request.Method == "OPTIONS" {
                        c.AbortWithStatus(204)
                        return
                }
                c.Next()
        }
}

func main() {
        // 读取配置文件
        args := os.Args[1:]
        if len(args) == 0 {
                fmt.Println("Please provide the configuration file as an argument.")
                return
        }
        configFile := args[0]

        viper.SetConfigType("yaml")
        viper.SetConfigFile(configFile)
        if err := viper.ReadInConfig(); err != nil {
                fmt.Println("Error reading config file:", err)
                return
        }

        // 从配置文件中读取参数
        remote_token = viper.GetString("Remote.apikey")
        mount_dir = viper.GetString("Mount.dir")
        port = viper.GetString("Server.port") // 服务端口

        // 初始化 Gin 引擎
        gin.SetMode(gin.ReleaseMode)
        r := gin.Default()

        // 添加跨域中间件
        r.Use(corsMiddleware())

        // 设置路由和文件流传输处理
        r.GET("/stream", remote)

        // 启动服务，监听指定端口
        r.Run(":" + port)
}