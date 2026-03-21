#!/bin/bash
# 测试 Web API 功能

set -e

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Agent-Town Web API ==="
echo ""

# 1. 测试世界信息
echo "1. Testing /world/info..."
curl -s $BASE_URL/world/info | jq .
echo ""

# 2. 测试命令列表
echo "2. Testing /commands..."
curl -s $BASE_URL/commands | jq .
echo ""

# 3. 测试注册
echo "3. Testing /auth/register..."
# 生成 ed25519 密钥对
PUBLIC_KEY=$(openssl rand -base64 32)
REGISTER_RESP=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"public_key\": \"$PUBLIC_KEY\", \"name\": \"WebTestAgent\"")
echo $REGISTER_RESP | jq .
AGENT_ID=$(echo $REGISTER_RESP | jq -r '.agent_id // empty')
echo "Agent ID: $AGENT_ID"
echo ""

# 4. 测试挑战
echo "4. Testing /auth/challenge..."
CHALLENGE_RESP=$(curl -s -X POST $BASE_URL/auth/challenge \
  -H "Content-Type: application/json" \
  -d "{\"public_key\": \"$PUBLIC_KEY\"")
echo $CHALLENGE_RESP | jq .
CHALLENGE_ID=$(echo $CHALLENGE_RESP | jq -r '.challenge_id // empty')
echo "Challenge ID: $CHALLENGE_ID"
echo ""

# 注意：由于需要 ed25519 签名，这里只测试 API 结构
# 完整认证需要使用 CLI 或正确的密钥签名

echo "=== Web API Test Complete ==="
echo ""
echo "注意: 完整认证需要 ed25519 签名，建议使用 CLI 工具:"
echo "  ./bin/cli agent create --name TestAgent"
echo "  ./bin/cli --agent TestAgent --server http://localhost:8080 status"
