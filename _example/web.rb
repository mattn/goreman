require 'sinatra'

set :port, ENV["PORT"]||'4567'

get '/' do
  content_type "text/plain; charset=utf-8"
  "hello #{ENV["AUTHOR"]}"
end
