wrk.method = "POST"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
wrk.body   = "username=user0&password=user"

function request()
   return wrk.format(wrk.method, "/login", wrk.headers, wrk.body)
end
