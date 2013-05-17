require 'sinatra'

get '/' do
    "hello #{ENV["AUTHOR"]}"
end
