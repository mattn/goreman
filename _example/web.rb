require "webrick"

port = Integer(ENV.fetch("PORT", "4567"))

server = WEBrick::HTTPServer.new(
  Port: port,
  BindAddress: "0.0.0.0",
  AccessLog: [],
  Logger: WEBrick::Log.new($stderr, WEBrick::Log::FATAL)
)

server.mount_proc "/" do |_req, res|
  res.status = 200
  res["Content-Type"] = "text/plain; charset=utf-8"
  res.body = "hello #{ENV.fetch("AUTHOR", "")}"
end

trap("INT") { server.shutdown }
trap("TERM") { server.shutdown }

server.start
