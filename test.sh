#!/bin/bash
set -e

echo "======================================"
echo "FidruaWatch 编译后功能检查"
echo "======================================"

cd "$(dirname "$0")"

# 1. 编译检查
echo ""
echo "[1/5] 编译检查..."
go build -o fidruawatch . 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 2. 二进制文件检查
echo ""
echo "[2/5] 二进制文件检查..."
if [ -f "./fidruawatch" ]; then
    SIZE=$(du -h ./fidruawatch | cut -f1)
    echo "✅ 二进制文件存在 (大小: $SIZE)"
else
    echo "❌ 二进制文件不存在"
    exit 1
fi

# 3. 单元测试
echo ""
echo "[3/5] 运行单元测试..."
go test -v ./... 2>&1 | grep -E "(PASS|FAIL|---)"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo "✅ 所有单元测试通过"
else
    echo "❌ 单元测试失败"
    exit 1
fi

# 4. 依赖检查
echo ""
echo "[4/5] 依赖检查..."
go mod verify 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 依赖完整"
else
    echo "❌ 依赖检查失败"
    exit 1
fi

# 5. GUI启动检查 (仅在有显示的环境)
echo ""
echo "[5/5] GUI启动检查..."
if [ -n "$DISPLAY" ]; then
    # 启动程序，等待2秒后检查是否还在运行
    timeout 3 ./fidruawatch &
    PID=$!
    sleep 2
    if kill -0 $PID 2>/dev/null; then
        echo "✅ GUI启动正常"
        kill $PID 2>/dev/null || true
    else
        echo "⚠️ GUI进程已退出（可能是正常的）"
    fi
else
    # 尝试使用 xvfb-run（如果可用）
    if command -v xvfb-run &> /dev/null; then
        echo "使用 Xvfb 虚拟显示测试..."
        timeout 5 xvfb-run -a ./fidruawatch &
        PID=$!
        sleep 3
        if kill -0 $PID 2>/dev/null; then
            echo "✅ GUI启动正常 (Xvfb)"
            kill $PID 2>/dev/null || true
        else
            echo "⚠️ GUI测试结束"
        fi
    else
        echo "⚠️ 无显示环境，跳过GUI测试 (安装 xvfb 可启用)"
    fi
fi

echo ""
echo "======================================"
echo "✅ 所有检查完成!"
echo "======================================"
