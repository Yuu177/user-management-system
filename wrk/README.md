wrk 测试 lua 脚本

wrk -t5 -c200 -d15s -T10s -s test.lua http://localhost:1088/login
1088 为 http server 的端口