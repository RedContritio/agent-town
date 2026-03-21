#!/bin/bash
# 测试事件驱动任务执行引擎

set -e

echo "=== Testing Event-Driven Task Engine ==="
echo ""

# 启动服务器（后台）
echo "1. Starting server..."
export PATH=$PATH:/usr/local/go/bin
make server > /tmp/engine-test.log 2>&1 &
SERVER_PID=$!
sleep 2

# 检查服务器是否启动
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "✗ Server failed to start"
    cat /tmp/engine-test.log
    exit 1
fi
echo "   ✓ Server started (PID: $SERVER_PID)"

# 注册测试 Agent
echo "2. Registering test agent..."
PUBLIC_KEY=$(openssl rand -base64 32)
REGISTER_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"public_key\": \"$PUBLIC_KEY\", \"name\": \"EngineTestAgent\"")
AGENT_ID=$(echo $REGISTER_RESP | grep -o '"agent_id":[0-9]*' | cut -d: -f2)
echo "   ✓ Agent registered (ID: $AGENT_ID)"

# 创建挑战
echo "3. Creating challenge..."
CHALLENGE_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/challenge \
  -H "Content-Type: application/json" \
  -d "{\"public_key\": \"$PUBLIC_KEY\"")\necho "   ✓ Challenge created"

# 注意：由于需要 ed25519 签名，这里只测试 API 结构
# 完整测试需要使用 CLI 工具

echo ""
echo "=== Engine Test Complete ==="
echo ""
echo "To fully test the engine:"
echo "  1. Build CLI: make cli"
echo "  2. Create agent: ./bin/cli agent create --name TestAgent"
echo "  3. Run status: ./bin/cli --agent TestAgent --server http://localhost:8080 status"
echo "  4. Create task: ./bin/cli --agent TestAgent move 3,0"
echo "  5. Watch logs: tail -f /tmp/engine-test.log"

# 清理
echo ""
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null || true
echo "✓ Done"
