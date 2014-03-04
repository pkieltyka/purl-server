require 'bundler/setup'
require 'net/http'
require 'uri'
require 'timeout'
require 'openssl'
require 'cgi'
require 'msgpack'
require 'pry'

MAXTIME = 30
PURL_SERVICE = 'http://localhost:9333/fetch'

urls = [
  'https://www.google.ca/',
  'https://www.facebook.com/',
  'https://twitter.com/',
  'http://www.pressly.com/',
  'http://nulayer.com/',
  'http://nulayer.com/',
  'http://nulayer.com/images/logo.png?x=http%3A%2F%2Fblah.com',
  # 'http://localhost:4567/',
  'http://faasdfasfdf23f23f23fwfasdf.com'
]


# GET test
puts "Doing GET test.."
purl_uri = PURL_SERVICE + '?' + urls.map {|u| "url[]=#{CGI.escape(u)}" }.join('&')

uri = URI.parse(purl_uri)
http = Net::HTTP.new(uri.host, uri.port)

puts "GET #{uri}"
resp = http.get(uri.to_s, {})
puts resp.code
obj = MessagePack.unpack(resp.body)
puts obj.inspect

2.times { puts }

# POST test
puts "Doing POST test.."
purl_uri = PURL_SERVICE

uri = URI.parse(purl_uri)
puts "POST #{urls}"

# We can send url[] or url
resp = Net::HTTP.post_form(uri, {'url' => urls})
puts resp.code
puts resp.body
obj = MessagePack.unpack(resp.body)
puts obj.inspect
