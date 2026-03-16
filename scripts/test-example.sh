#!/usr/bin/env bash
# 功能测试示例：注册 → 登录 → 创建职位 → 看推荐 → 投递 → 再看推荐（排除已投递）
# 使用前确保已启动所有服务：./scripts/start-all.sh
# 每次运行使用新邮箱，保证能看出「投递前推荐多、投递后推荐少」的差异。

set -e
BASE="${BASE_URL:-http://localhost:8080}"
API="$BASE/api/v1"
DEMO_EMAIL="demo+$(date +%s)@internhub.test"
DEMO_PASS="password123"

# 若系统有 jq 则用于格式化输出和提取 token
if command -v jq >/dev/null 2>&1; then
  JQ="jq"
else
  JQ=""
fi

echo "========== 1. 健康检查 =========="
curl -s "$BASE/health" | ${JQ:-cat}
echo ""

echo "========== 2. 注册用户（本次邮箱: $DEMO_EMAIL） =========="
REGISTER_RES=$(curl -s -X POST "$API/users/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$DEMO_EMAIL\",\"password\":\"$DEMO_PASS\",\"name\":\"Demo User\"}")
echo "$REGISTER_RES" | ${JQ:-cat} 2>/dev/null || echo "$REGISTER_RES"
echo ""

echo "========== 3. 登录获取 Token =========="
LOGIN_RES=$(curl -s -X POST "$API/users/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$DEMO_EMAIL\",\"password\":\"$DEMO_PASS\"}")
if [ -n "$JQ" ]; then
  TOKEN=$(echo "$LOGIN_RES" | jq -r '.access_token // empty')
  echo "$LOGIN_RES" | jq .
else
  TOKEN=$(echo "$LOGIN_RES" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
  echo "$LOGIN_RES"
fi
if [ -z "$TOKEN" ]; then
  echo "登录失败，请检查服务与账号密码"
  exit 1
fi
echo "Token 已获取（前 20 字符）: ${TOKEN:0:20}..."
echo ""

echo "========== 4. 创建多个职位 =========="
for payload in \
  '{"title":"Go 后端实习","company":"字节跳动","link":"https://jobs.bytedance.com"}' \
  '{"title":"前端开发实习","company":"腾讯","link":"https://careers.tencent.com"}' \
  '{"title":"算法实习","company":"阿里云","link":"https://talent.alibaba.com"}' \
  '{"title":"运维开发实习","company":"美团","link":"https://zhaopin.meituan.com"}'; do
  curl -s -X POST "$API/jobs" -H "Content-Type: application/json" -d "$payload" | ${JQ:-cat} 2>/dev/null || true
done
echo ""

echo "========== 5. 获取职位列表 =========="
curl -s "$API/jobs" | ${JQ:-cat}
echo ""

echo "========== 6. 岗位推荐（尚未投递，应包含当前全部职位） =========="
REC6=$(curl -s -H "Authorization: Bearer $TOKEN" "$API/recommendations")
echo "$REC6" | ${JQ:-cat}
if [ -n "$JQ" ]; then COUNT6=$(echo "$REC6" | jq '.list | length'); else COUNT6=""; fi
echo "(本步推荐条数: ${COUNT6:-—})"
echo ""

echo "========== 7. 投递职位 1 和 2 =========="
curl -s -X POST "$API/applications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"job_id":1}' | ${JQ:-cat}
curl -s -X POST "$API/applications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"job_id":2}' | ${JQ:-cat}
echo ""

echo "========== 8. 我的投递列表 =========="
curl -s -H "Authorization: Bearer $TOKEN" "$API/applications/me" | ${JQ:-cat}
echo ""

echo "========== 9. 岗位推荐（已投递 1、2，应比步骤 6 少 2 条） =========="
REC9=$(curl -s -H "Authorization: Bearer $TOKEN" "$API/recommendations")
echo "$REC9" | ${JQ:-cat}
if [ -n "$JQ" ]; then COUNT9=$(echo "$REC9" | jq '.list | length'); else COUNT9=""; fi
echo "(本步推荐条数: ${COUNT9:-—}，应比步骤 6 少 2)"
echo ""

echo "========== 10. 更新个人资料（昵称） =========="
curl -s -X PATCH "$API/users/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nickname":"小Demo"}' | ${JQ:-cat}
echo ""

echo "========== 11. 查看当前用户资料 =========="
curl -s -H "Authorization: Bearer $TOKEN" "$API/users/me" | ${JQ:-cat}
echo ""

echo "========== 测试完成 =========="
echo "每次运行使用新账号，步骤 6 为「未投递」时的推荐，步骤 9 为「已投递 1、2」后的推荐，条数应少 2。"
echo "若配置了 OPENAI_API_KEY，推荐接口会返回 reason 与 summary；未配置则按时间倒序。"
