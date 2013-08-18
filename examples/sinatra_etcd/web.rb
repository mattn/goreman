require 'sinatra'

get '/' do
    "DB #{ENV["DB"]}"
end

get '/env/' do
    ENV
end
