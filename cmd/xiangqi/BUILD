# Ebiten 渲染層：原始碼以 `//go:build ebiten` 標籤隔離，且嵌入 CJK 字型。
# 預設建構/測試（含無頭 CI、`pants test ::`）不編譯它，故圖形相依不影響 CI。
#
# Pants 的 Go 後端不支援自訂 build tag，故以 shell 目標包一層原生 `go` 指令：
#   pants run cmd/xiangqi:build   # 編譯 GUI 二進位到 ./bin/xiangqi
#   pants run cmd/xiangqi:gui     # 直接執行 GUI（需圖形環境）
go_package()

run_shell_command(
    name="build",
    command="go build -tags ebiten -o bin/xiangqi ./cmd/xiangqi && echo '✔ built bin/xiangqi'",
    workdir="/",
)

run_shell_command(
    name="gui",
    command="go run -tags ebiten ./cmd/xiangqi",
    workdir="/",
)

go_binary(
    name="bin",
)
