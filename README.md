# 📁 Upload Util

一个功能完整的文件上传工具，支持多种云存储服务和本地存储。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

## ✨ 特性

- 🚀 **多云存储支持**：阿里云 OSS、腾讯云 COS、华为云 OBS、AWS S3、MinIO
- 📂 **本地存储**：支持本地文件系统存储
- 🔐 **安全访问**：私有存储桶预签名 URL 支持
- 📤 **批量上传**：支持单文件和多文件上传
- 🎯 **文件验证**：文件大小、格式限制
- 🏷️ **灵活命名**：支持多种文件命名策略
- 🌐 **REST API**：完整的 HTTP API 接口
- ⚙️ **配置驱动**：YAML 配置文件，支持多环境

## 🏗️ 项目结构

## 项目结构

```bash
upload-util/
├── cmd/                       # 应用入口
│   ├── main.go
│   └── http.go
├── internal/                  # 私有代码
│   ├── config/                # 配置管理
│   │   └── config.go
│   ├── handler/               # HTTP 处理器
│   │   └── handler.go
│   ├── middleware/            # 中间件
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── recovery.go
│   ├── router/                # 路由配置
│   │   └── router.go
│   └── service/               # 业务逻辑
│       ├── factory.go
│       ├── aliyun-uploader.go
│       ├── huawei-uploader.go
│       ├── local-uploader.go
│       ├── minio-uploader.go
│       ├── qcloud-uploader.go
│       ├── s3-uploader.go
│       ├── tencent-uploader.go
│       └── util.go
├── script/                    # 构建脚本
│   └── build.sh
└── config-example.yaml         # 配置示例
```


## 📦 安装

### 从源码编译
#### 克隆仓库
```shell 
git clone [https://github.com/your-username/upload-util.git](https://github.com/your-username/upload-util.git) cd upload-util
```
#### 编译
```shell
go build -o upload-util ./cmd
```
#### 或使用构建脚本（支持多平台）
```shell
chmod +x script/build.sh ./script/build.sh
```

#### 预编译二进制
从 [Releases](https://github.com/your-username/upload-util/releases) 页面下载对应平台的预编译版本。

## 🚀 快速开始

### 1. 创建配置文件

bash cp config-example.yaml config.yaml


### 2. 编辑配置

```yaml
upload:
  # 上传类型: local, oss, minio
  type: oss
  
  oss:
    provider: aliyun
    aliyun:
      endpoint: oss-cn-hangzhou.aliyuncs.com
      access-key-id: your-access-key
      access-key-secret: your-secret-key
      bucket: your-bucket
      path-prefix: uploads/
      use-ssl: true
```
### 3. 启动服务
``` shell
./upload-util -config=config.yaml
```
服务默认在 :8080 端口启动。

## 📚 API 文档

### 上传单个文件
```shell
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/your/file.jpg"
```
**响应示例：**

```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "url": "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/uploads/uuid.jpg",
    "key": "uploads/uuid.jpg",
    "size": 1024,
    "mime_type": "image/jpeg",
    "filename": "file.jpg"
  }
}
```

### 批量上传

```shell
curl -X POST http://localhost:8080/upload/multiple \
  -F "files=@file1.jpg" \
  -F "files=@file2.png"
```

### 获取访问链接
```shell
curl "http://localhost:8080/url?key=uploads/uuid.jpg"
```
### 删除文件

```shell
curl -X DELETE http://localhost:8080/delete \
  -H "Content-Type: application/json" \
  -d '{"key": "uploads/uuid.jpg"}'
```

### 健康检查

```shell
curl http://localhost:8080/health
```
## ⚙️ 配置说明

### 支持的存储类型
|类型 |   描述 |      配置节点|
| -- | --   | --|
|local| 本地文件系统| upload.local|
|oss |  云对象存储 |  upload.oss|
|minio| MinIO 存储|   upload.minio|

### 支持的云存储提供商
|提供商 | Provider 值| 描述|
| --    | --    |   --  |
|阿里云 OSS|aliyun|阿里云对象存储|
|腾讯云 COS|tencent|腾讯云对象存储|
|华为云 OBS|huawei|华为云对象存储|
|AWS S3|aws|Amazon S3|
|其他 S3 兼容|qcloud|S3 兼容存储|

### 文件命名策略
 * original：保持原文件名
 * uuid：使用 UUID（默认）
 * timestamp：使用时间戳

### 完整配置示例

```yaml
upload:
  type: oss
  
  # 本地存储配置
  local:
    path: ~/uploads
    url-prefix: http://localhost:8080/files
  
  # 云存储配置
  oss:
    provider: aliyun
    aliyun:
      endpoint: oss-cn-hangzhou.aliyuncs.com
      access-key-id: your-key
      access-key-secret: your-secret
      bucket: your-bucket
      path-prefix: uploads/
      use-ssl: true
      # 签名 URL 有效期（秒）
      sign-url-expiry: 3600

# 上传限制
upload-settings:
  max-file-size: 100  # MB
  allowed-extensions:
    - .jpg
    - .jpeg
    - .png
    - .pdf
  filename-strategy: uuid
  keep-original-name: false
```