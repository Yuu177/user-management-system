wrk.method = "POST"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"

function request()
   local n = math.random(10000000)
   local body = "username=user"..n.."&password=user"
   wrk.body = body
   return wrk.format(wrk.method, "/login", wrk.headers, wrk.body)
end
