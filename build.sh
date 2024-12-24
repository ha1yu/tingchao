#!/bin/sh

echo "开始构建duck-cc-server"
echo "移除dist目录"

rm -rf ./dist
mkdir dist

# linux 打包
echo "开始构建duck-cc-http-server-linux-amd64包"
mkdir -p ./dist/tmp_linux_amd64
cp -r ./conf ./dist/tmp_linux_amd64
cp -r ./ccstatic ./dist/tmp_linux_amd64
cp -r ./webstatic ./dist/tmp_linux_amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w" -o ./dist/tmp_linux_amd64/duck-cc-server-http-linux64 main.go
chmod +x ./dist/tmp_linux_amd64/duck-cc-server-http-linux64
cp ./dist/tmp_linux_amd64/duck-cc-server-http-linux64 ./dist

echo "构建duck-cc-http-server-linux-amd64二进制包成功"

tar zcf ./dist/duck-cc-http-server-linux-amd64.tar.gz -C ./dist/tmp_linux_amd64 .
echo "生成duck-cc-http-server-linux-amd64.tar.gz成功"

rm -rf ./dist/tmp_linux_amd64
echo "移除缓存目录dist/tmp_linux_amd64"


# win打包
echo "开始构建duck-cc-http-server-win-amd64包"
mkdir -p ./dist/tmp_win_amd64
cp -r ./conf ./dist/tmp_win_amd64
cp -r ./ccstatic ./dist/tmp_win_amd64
cp -r ./webstatic ./dist/tmp_win_amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags "-s -w" -o ./dist/tmp_win_amd64/duck-cc-server-http-win64.exe main.go
cp ./dist/tmp_win_amd64/duck-cc-server-http-win64.exe ./dist

echo "构建duck-cc-server-http-win64.exe二进制包成功"

tar zcf ./dist/duck-cc-http-server-win-amd64.tar.gz -C ./dist/tmp_win_amd64 .
echo "生成duck-cc-http-server-win-amd64.tar.gz成功"

rm -rf ./dist/tmp_win_amd64
echo "移除缓存目录dist/tmp_win_amd64"


echo "构建duck-cc-server完成"
echo "Success"


# 不加 -a 参数能极大的提升编译速度
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./dist/duck-cc-server-http-win64.exe main.go
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./dist/duck-cc-server-http-linux64 main.go


