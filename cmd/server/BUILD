# 本機 WebSocket 對局伺服器入口（無 build tag，可直接產出 binary）。
#   pants run cmd/server:bin        # 編譯/執行伺服器二進位
#   pants run cmd/server:server     # 直接以原生 go run 啟動（可附 -addr）
go_package()

go_binary(
    name="bin",
)

run_shell_command(
    name="server",
    command="go run ./cmd/server",
    workdir="/",
)
