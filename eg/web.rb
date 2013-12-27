require 'sinatra'

set :port, ENV["PORT"]||'4567'

get '/' do
    "hello #{ENV["AUTHOR"]}"
end
