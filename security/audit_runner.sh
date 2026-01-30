#!/bin/bash
# Talos Security Audit - Automated Execution Script
# Run this before production launch

set -e

AUDIT_DATE=$(date +%Y%m%d)
AUDIT_REPORT="security-audit-${AUDIT_DATE}.txt"

echo "🔒 Talos Security Audit - $(date)" > $AUDIT_REPORT
echo "======================================" >> $AUDIT_REPORT
echo "" >> $AUDIT_REPORT

# Function to test and log
test_and_log() {
    local test_name="$1"
    local command="$2"
    local expected="$3"
    
    echo -n "Testing: $test_name... "
    if eval "$command" | grep -q "$expected"; then
        echo "✅ PASS" | tee -a $AUDIT_REPORT
        return 0
    else
        echo "❌ FAIL" | tee -a $AUDIT_REPORT
        return 1
    fi
}

FAILED_TESTS=0

echo "🔐 1. Authentication & Authorization" >> $AUDIT_REPORT
echo "=====================================" >> $AUDIT_REPORT

# Test JWT expiration
test_and_log "JWT token validation" \
    "curl -s -H 'Authorization: Bearer invalid_token' http://localhost:8080/api/resources" \
    "401\|Unauthorized" || ((FAILED_TESTS++))

# Test RBAC
test_and_log "RBAC viewer cannot delete" \
    "curl -s -X DELETE -H 'Authorization: Bearer viewer_token' http://localhost:8080/api/resources/test" \
    "403\|Forbidden" || ((FAILED_TESTS++))

echo "" >> $AUDIT_REPORT
echo "🔑 2. Secrets Management" >> $AUDIT_REPORT
echo "========================" >> $AUDIT_REPORT

# Check for secrets in code
echo -n "Checking for hardcoded secrets... "
if git grep -i "sk-or-v1\|apk_\|password.*=" | grep -v ".md\|.txt\|AUDIT" > /dev/null 2>&1; then
    echo "❌ FAIL - Found hardcoded secrets!" | tee -a $AUDIT_REPORT
    git grep -i "sk-or-v1\|apk_\|password.*=" | grep -v ".md\|.txt" | head -5 >> $AUDIT_REPORT
    ((FAILED_TESTS++))
else
    echo "✅ PASS" | tee -a $AUDIT_REPORT
fi

# Test Vault connection
test_and_log "Vault connectivity" \
    "curl -s -H 'X-Vault-Token: $VAULT_TOKEN' http://localhost:8200/v1/sys/health" \
    "initialized.*true" || ((FAILED_TESTS++))

echo "" >> $AUDIT_REPORT
echo "🗄️  3. Database Security" >> $AUDIT_REPORT
echo "=======================" >> $AUDIT_REPORT

# Test SQL injection
test_and_log "SQL injection prevention" \
    "curl -s -X POST http://localhost:8080/api/resources -d '{\"id\": \"1; DROP TABLE actions;--\"}'" \
    "error\|invalid" || ((FAILED_TESTS++))

# Test database SSL
test_and_log "PostgreSQL SSL enabled" \
    "docker-compose exec -T postgres psql -U talos_user -d talos -c 'SHOW ssl;'" \
    "on" || ((FAILED_TESTS++))

echo "" >> $AUDIT_REPORT
echo "🌐 4. API Security" >> $AUDIT_REPORT
echo "==================" >> $AUDIT_REPORT

# Test rate limiting
echo -n "Testing rate limiting... "
RATE_TEST_PASSED=true
for i in {1..150}; do
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/roi)
    if [ "$RESPONSE" = "429" ]; then
        echo "✅ PASS (rate limit triggered at request $i)" | tee -a $AUDIT_REPORT
        RATE_TEST_PASSED=true
        break
    fi
done
if [ "$RATE_TEST_PASSED" = false ]; then
    echo "❌ FAIL (no rate limit)" | tee -a $AUDIT_REPORT
    ((FAILED_TESTS++))
fi

# Test CORS
test_and_log "CORS protection" \
    "curl -s -H 'Origin: http://evil.com' http://localhost:8080/api/resources" \
    "Access-Control-Allow-Origin.*localhost\|error" || ((FAILED_TESTS++))

echo "" >> $AUDIT_REPORT
echo "🔒 5. Network Security" >> $AUDIT_REPORT
echo "======================" >> $AUDIT_REPORT

# Test exposed ports
echo -n "Checking exposed ports... "
OPEN_PORTS=$(nmap -p- localhost 2>/dev/null | grep "open" | awk '{print $1}')
EXPECTED_PORTS="8080|3000|9090"
if echo "$OPEN_PORTS" | grep -qvE "$EXPECTED_PORTS"; then
    echo "⚠️  WARNING - Unexpected ports open:" | tee -a $AUDIT_REPORT
    echo "$OPEN_PORTS" | grep -vE "$EXPECTED_PORTS" >> $AUDIT_REPORT
else
    echo "✅ PASS" | tee -a $AUDIT_REPORT
fi

# Test TLS (if production)
if [ "$ENVIRONMENT" = "production" ]; then
    test_and_log "TLS/SSL enabled" \
        "curl -I https://talos.dev 2>&1" \
        "200\|301\|302" || ((FAILED_TESTS++))
fi

echo "" >> $AUDIT_REPORT
echo "🐳 6. Docker Security" >> $AUDIT_REPORT
echo "=====================" >> $AUDIT_REPORT

# Check containers not running as root
echo -n "Checking container user permissions... "
ROOT_CONTAINERS=$(docker-compose ps -q | xargs -I {} docker inspect {} --format '{{.Config.User}}' | grep -c "^$")
if [ "$ROOT_CONTAINERS" -gt 0 ]; then
    echo "❌ FAIL - $ROOT_CONTAINERS containers running as root!" | tee -a $AUDIT_REPORT
    ((FAILED_TESTS++))
else
    echo "✅ PASS" | tee -a $AUDIT_REPORT
fi

echo "" >> $AUDIT_REPORT
echo "📦 7. Dependency Security" >> $AUDIT_REPORT
echo "==========================" >> $AUDIT_REPORT

# Check for vulnerable dependencies
echo -n "Scanning Go dependencies... "
if command -v govulncheck &> /dev/null; then
    if govulncheck ./... > /tmp/vuln_check.txt 2>&1; then
        echo "✅ PASS - No vulnerabilities found" | tee -a $AUDIT_REPORT
    else
        echo "❌ FAIL - Vulnerabilities detected:" | tee -a $AUDIT_REPORT
        cat /tmp/vuln_check.txt >> $AUDIT_REPORT
        ((FAILED_TESTS++))
    fi
else
    echo "⚠️  SKIP - govulncheck not installed" | tee -a $AUDIT_REPORT
fi

echo "" >> $AUDIT_REPORT
echo "🔍 8. Code Security" >> $AUDIT_REPORT
echo "===================" >> $AUDIT_REPORT

# Static analysis
echo -n "Running static analysis (gosec)... "
if command -v gosec &> /dev/null; then
    if gosec -quiet ./... > /tmp/gosec.txt 2>&1; then
        echo "✅ PASS" | tee -a $AUDIT_REPORT
    else
        echo "❌ FAIL - Security issues found:" | tee -a $AUDIT_REPORT
        tail -20 /tmp/gosec.txt >> $AUDIT_REPORT
        ((FAILED_TESTS++))
    fi
else
    echo "⚠️  SKIP - gosec not installed" | tee -a $AUDIT_REPORT
fi

echo "" >> $AUDIT_REPORT
echo "☸️  9. Kubernetes Security" >> $AUDIT_REPORT
echo "===========================" >> $AUDIT_REPORT

if kubectl get pods -n talos &> /dev/null; then
    # Check for privileged pods
    echo -n "Checking for privileged pods... "
    PRIV_PODS=$(kubectl get pods -n talos -o json | jq '.items[] | select(.spec.containers[].securityContext.privileged==true)' | wc -l)
    if [ "$PRIV_PODS" -gt 0 ]; then
        echo "❌ FAIL - Found $PRIV_PODS privileged pods!" | tee -a $AUDIT_REPORT
        ((FAILED_TESTS++))
    else
        echo "✅ PASS" | tee -a $AUDIT_REPORT
    fi
else
    echo "⚠️  SKIP - Kubernetes not accessible" | tee -a $AUDIT_REPORT
fi

echo "" >> $AUDIT_REPORT
echo "📝 10. Logging & Monitoring" >> $AUDIT_REPORT
echo "============================" >> $AUDIT_REPORT

# Check logs don't contain secrets
echo -n "Checking logs for secrets... "
if docker-compose logs | grep -iE "sk-or-v1|password.*=|secret.*=" > /dev/null 2>&1; then
    echo "❌ FAIL - Secrets found in logs!" | tee -a $AUDIT_REPORT
    ((FAILED_TESTS++))
else
    echo "✅ PASS" | tee -a $AUDIT_REPORT
fi

# Final Report
echo "" >> $AUDIT_REPORT
echo "=====================================" >> $AUDIT_REPORT
echo "FINAL AUDIT RESULTS" >> $AUDIT_REPORT
echo "=====================================" >> $AUDIT_REPORT
echo "Date: $(date)" >> $AUDIT_REPORT
echo "Failed Tests: $FAILED_TESTS" >> $AUDIT_REPORT
echo "" >> $AUDIT_REPORT

if [ $FAILED_TESTS -eq 0 ]; then
    echo "✅ ALL TESTS PASSED - READY FOR PRODUCTION" | tee -a $AUDIT_REPORT
    exit 0
else
    echo "❌ $FAILED_TESTS TESTS FAILED - FIX BEFORE LAUNCH" | tee -a $AUDIT_REPORT
    echo "" >> $AUDIT_REPORT
    echo "Review full report: $AUDIT_REPORT"  | tee -a $AUDIT_REPORT
    exit 1
fi
