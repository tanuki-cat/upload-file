#!/bin/bash

# 构建所有工具的脚本
set -e

APP_NAME="upload-util"
VERSION=${1:-$(date +%Y%m%d%H%M%S)}
OUTPUT_DIR="./bin"
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

LDFLAGS="-s -w -X 'main.Version=${VERSION}' -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuildTime=${BUILD_TIME}'"

declare -a COMMANDS=(
    "cmd/server:upload-server"
    "cmd/upload:upload-cli"
    "cmd/batch:batch-upload"
    "cmd/interactive:upload-interactive"
)

declare -a PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "🚀 开始构建 ${APP_NAME} v${VERSION}"
echo "📝 Git Commit: ${GIT_COMMIT}"
echo "⏰ Build Time: ${BUILD_TIME}"
echo ""

# 清理输出目录
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

for cmd_info in "${COMMANDS[@]}"; do
    IFS=':' read -r source_dir output_name <<< "$cmd_info"

    echo "📦 构建 ${output_name}..."

    for platform in "${PLATFORMS[@]}"; do
        GOOS=${platform%/*}
        GOARCH=${platform#*/}

        output_file="${output_name}_${VERSION}_${GOOS}_${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            output_file="${output_file}.exe"
        fi

        output_path="${OUTPUT_DIR}/${output_file}"

        echo "  📦 ${GOOS}/${GOARCH}..."

        env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
            -ldflags="$LDFLAGS" \
            -o "$output_path" \
            "./$source_dir"
    done
    echo ""
done

echo "🎉 构建完成！"
echo "📁 输出目录: ${OUTPUT_DIR}"
echo ""
echo "📋 构建的工具:"
echo "  🌐 upload-server     - HTTP API 服务器"
echo "  📤 upload-cli        - 单文件上传命令行工具"
echo "  📁 batch-upload      - 批量上传工具"
echo "  💬 upload-interactive - 交互式命令行界面"
echo ""
ls -la "$OUTPUT_DIR"