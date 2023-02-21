wrk 测试 lua 脚本

wrk -t5 -c200 -d50s -T20s -s fixedLogin.lua http://localhost:1088/login
1088 为 http server 的端口
