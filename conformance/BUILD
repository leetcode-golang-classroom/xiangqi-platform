# go:embed *.json 需要把 fixtures 宣告為 resources 並讓 go_package 依賴它
resources(
    name="fixtures",
    sources=["*.json"],
)

go_package(
    dependencies=[":fixtures"],
)
