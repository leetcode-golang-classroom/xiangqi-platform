#!/usr/bin/env bash
# 暫時性 workaround：Pants 的 Go 後端會對 Go>=1.20 設定
# GOEXPERIMENT=nocoverageredesign，但 Go>=1.24 已移除該 experiment，
# 導致所有 go 指令報錯：unknown GOEXPERIMENT coverageredesign。
#
# 本腳本把 Pants vendored sdk.py 的版本判斷收斂為「僅 1.20<=Go<1.24 才設定」，
# 讓 Pants 可在 Go 1.24/1.25/1.26 上運作。
#
# 冪等；可重複執行。Pants 版本變更（重新下載 venv）後需再跑一次。
# 追蹤上游修復後即可移除。
set -euo pipefail

found=0
while IFS= read -r f; do
  found=1
  if grep -q 'is_compatible_version("1.20") and not goroot.is_compatible_version("1.24")' "$f"; then
    echo "已套用，略過：$f"
    continue
  fi
  sed -i \
    's/if goroot.is_compatible_version("1.20"):/if goroot.is_compatible_version("1.20") and not goroot.is_compatible_version("1.24"):/' \
    "$f"
  echo "已修補：$f"
done < <(find "${HOME}/.cache/nce" -type f -path "*backend/go/util_rules/sdk.py" 2>/dev/null)

if [ "$found" -eq 0 ]; then
  echo "找不到 Pants 的 sdk.py（請先執行 'pants version' 以下載 venv）" >&2
  exit 1
fi
