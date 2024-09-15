# YASS-Backend
**Yet Another Stream Splitter (Generally for EMBY)**

![Main Branch Build CI](https://github.com/FacMata/YASS-Backend/actions/workflows/build.yml/badge.svg?branch=main)

## 这是什么

### YASS

一个基于 [MisakaFxxk](https://github.com/MisakaFxxk) 的 [Go_stream](https://github.com/MisakaFxxk/Go_stream) 项目改进而来的，EMBY 视频流分离推送解决方案的程序组。



在 [MisakaFxxk](https://github.com/MisakaFxxk) 没有更新的前提下，它与 YASS-Frontend 可以作为原程序的后继者。



### YASS Backend

YASS 项目的后端程序。其完成的工作是从本地目录找到前端请求的实际文件，以视频流的形式传递给客户端。



本程序在 [Go_stream](https://github.com/MisakaFxxk/Go_stream) 的基础上利用更多 Go 的特性，比如分片传输、并发缓存池、连接回收等，播放效率相比原版有一定提升。



## 如何配置

#### 1. 下载最新 Release

使用 `unzip` 解压到你的运行目录下。

#### 2. 配置 `config.yaml`

```yaml
# 播放端配置
Remote:
  apikey: "your-api-key"

# 目录头配置
Mount: 
  dir: "/mnt"

# 服务器配置
Server:
  port: "12180"
```

此处 **目前** 可参考 [Go_stream](https://github.com/MisakaFxxk/Go_stream) 的相关配置进行操作。

#### 3. 运行程序

```shell
# sudo chmod +x <filename>
# ./<filename> config.yaml
```

持久化运行推荐使用 `SystemD System Service` 或者 `screen` 。



## 寻求交流

Email: [contact@facmata.net](mailto://contact@facmata.net)

Telegram Group: [YASS Talking](https://t.me/YASS_Talking)
