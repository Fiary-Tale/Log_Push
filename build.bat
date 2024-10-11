@echo off
CHCP 65001
echo "开始编译 Log_Push_mac_inter"

SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o ./mark/Log_Push_mac_amd64 -trimpath -ldflags "-s -w -buildid=" main.go

echo "开始编译 Log_Push_mac_m1"
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=arm64
go build -o ./mark/Log_Push_mac_arm64 -trimpath -ldflags "-s -w -buildid=" main.go

echo "开始编译 Log_Push_Windows_amd64"
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o ./mark/Log_Push_win_amd64.exe -trimpath -ldflags "-s -w -buildid=" main.go

echo "开始编译 Log_Push_linux_amd64"
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o ./mark/Log_Push_linux_amd64 -trimpath -ldflags "-s -w -buildid=" main.go

echo "开始编译 Log_Push_linux_arm64"
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=arm64
go build -o ./mark/Log_Push_linux_arm64 -trimpath -ldflags "-s -w -buildid=" main.go

echo "编译完成!请按任意键退出!"
Pause