# Ebiten 渲染層：原始碼以 `//go:build ebiten` 標籤隔離，且嵌入 CJK 字型。
# 預設建構/測試（含無頭 CI、`pants test ::`）不編譯它，故圖形相依不影響 CI。
# 因此刻意不放 go_package/go_binary：標籤隔離後 Pants 看不到任何原始碼，
# `pants check ::` 會報 no buildable Go source files。
#
# Pants 的 Go 後端不支援自訂 build tag，故以 shell 目標包一層原生 `go` 指令：
#   pants run cmd/xiangqi:build   # 編譯 GUI 二進位到 ./bin/xiangqi
#   pants run cmd/xiangqi:gui     # 直接執行 GUI（需圖形環境）
#
# gui 指定 EGL_PLATFORM=x11：Wayland 桌面上 GLFW(X11) 的 EGL 會誤選 wayland 平台，
# 在 glfwCreateWindow 內 SIGSEGV（wl_proxy_create_wrapper）。
run_shell_command(
    name="build",
    command="go build -tags ebiten -o bin/xiangqi ./cmd/xiangqi && echo '✔ built bin/xiangqi'",
    workdir="/",
)

run_shell_command(
    name="gui",
    command="EGL_PLATFORM=x11 go run -tags ebiten ./cmd/xiangqi",
    workdir="/",
)
