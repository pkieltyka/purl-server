require 'bundler/setup'
require 'sinatra'

get '/' do
  sleep 5
  'Hi hi'
end
