#!/bin/bash


# 检查第一个参数是否是 is "release"
if [ "$1" == "release" ] || [ "$1" == "r" ]; then
    MODE_FLAG="-X main.mod=release"
else
    MODE_FLAG="-X main.mod=debug"
fi



REPO_NAME=$(basename $(git rev-parse --show-toplevel))

# 获取最新提交的哈希值
LATEST_COMMIT_HASH=$(git rev-parse HEAD)

# 获取最新提交的标签
LATEST_COMMIT_TAG=$(git tag --contains $LATEST_COMMIT_HASH)

# 获取最近的v开头的标签
LATEST_V_TAG=$(git describe --tags --abbrev=0 `git rev-list --tags --max-count=1` 2>/dev/null)

if [ -z "$LATEST_COMMIT_TAG" ]; then
    # 当最新的提交没有tag时
    if [[ "$LATEST_V_TAG" == v* ]]; then
        # 如果最近的标签以v开头，则取最近的v开头的tag然后在Tag后面加上-dev-最新的提交的短哈希
        VERSION="$LATEST_V_TAG-dev-$(git rev-parse --short $LATEST_COMMIT_HASH)"
    else
        # 如果没有找到v开头的标签，设置默认值
        VERSION="1.0.0"
    fi
else
    # 当最新的提交存在tag时
    if [[ "$LATEST_COMMIT_TAG" == v* ]]; then
        # 并且v开头时，则取Tag的内容为version
        VERSION=$LATEST_COMMIT_TAG
    else
        # 如果标签不是以v开头，设置默认值
        VERSION="1.0.0"
    fi
fi


# 其他变量
GIT_HASH=$(git rev-parse HEAD)
BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S)
LD_FLAGS="-X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitHash=$GIT_HASH $MODE_FLAG"

# 打印信息
echo "=============================================="
echo "            BUILD INFORMATION"
echo "=============================================="
echo ""
echo "REPO_NAME:   $REPO_NAME"
echo "VERSION:     $VERSION"
echo "BUILD_DATE:  $BUILD_DATE"
echo "GIT_HASH:    $GIT_HASH"
echo "MODE_FLAG:   $MODE_FLAG"
echo ""
echo "=============================================="
echo "            BUILD PROCESS"
echo "=============================================="


# 编译流程
echo "Running 'go mod tidy'"
go mod tidy

# Build ARM Linux version
echo "Building $REPO_NAME-linux-arm"
GOOS=linux GOARCH=arm go build -ldflags "$LD_FLAGS" -o "output/$REPO_NAME-linux-arm" ./...

# Build AMD64 Linux version
echo "Building $REPO_NAME-linux-amd64"
GOOS=linux GOARCH=amd64 go build -ldflags "$LD_FLAGS" -o "output/$REPO_NAME-linux-amd64" ./...

# Build Windows AMD64 version with .exe extension
echo "Building $REPO_NAME-windows-amd64.exe"
GOOS=windows GOARCH=amd64 go build -ldflags "$LD_FLAGS" -o "output/$REPO_NAME-windows-amd64.exe" ./...

# 进入output目录
cd "output"

# 将 .env 和 auth.yaml 打包到输出产物
echo "=============================================="
echo "            PACKAGING"
echo "=============================================="
echo ""
echo "Packaging $REPO_NAME-linux-arm.tar.gz"
tar -czf "$REPO_NAME-linux-arm.tar.gz" "$REPO_NAME-linux-arm" ../.env ../auth.yaml

echo "Packaging $REPO_NAME-linux-amd64.tar.gz"
tar -czf "$REPO_NAME-linux-amd64.tar.gz" "$REPO_NAME-linux-amd64" ../.env ../auth.yaml

echo "Packaging $REPO_NAME-windows-amd64.tar.gz"
tar -czf "$REPO_NAME-windows-amd64.tar.gz" "$REPO_NAME-windows-amd64.exe" ../.env ../auth.yaml

# 清理临时文件
echo "=============================================="
echo "            CLEAN UP"
echo "=============================================="
echo ""
echo "Removing temporary files"
#rm "$REPO_NAME-linux-arm"
#rm "$REPO_NAME-linux-amd64"
#rm "$REPO_NAME-windows-amd64.exe"

echo "Packaging complete!"
