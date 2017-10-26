#!/usr/bin/ruby

require 'rubygems'
require 'websocket-client-simple'
require 'json'

require 'pp'

a='ws://localhost:3046/msg/ws'

ws = WebSocket::Client::Simple.connect a

ws.on :message do |msg|
  puts "In: "+msg.data#.inspect[0..60]
  pp JSON.parse(msg.data)
  # proc_msg(msg.data)
end

ws.on :open do
  puts "Opened"
  ws.send JSON.generate( "$" => 42, "!" => "select", "#" => { "n" => 20, "patterns" => [['ad'],['x','a']] } )
  ws.send JSON.generate( "$" => 7, "!" => "post", "#" => { "a" => ["ad","df"], "d" => "The data" } )
  ws.send JSON.generate( "$" => 23, "!" => "wat", "#" => { "id" => 1 } )
  puts "Sent..."
end

ws.on :close do |e|
  puts "Closed"
  p e
  exit 1
end

ws.on :error do |e|
  puts "Error!"
  p e
  exit 1
end

sleep 4

1.upto 10 do |n|
  a=[%w{ ad bd tst x }[rand(4)]]
  1.upto(1+rand(3)) do |x|
    a.push(rand(36**x).to_s(36))
  end
  puts "Sent again"
  ws.send JSON.generate( "$" => rand(15), "!" => "post", "#" => { "a" => a, "d" => "The data #{rand(36**12).to_s(36)}" } ) if rand(15) < 12
  ws.send JSON.generate( "$" => rand(15), "!" => "post", "a" => a, "d" => "The inline #{rand(36**12).to_s(36)}" ) if rand(15) < 12
  sleep 3
end
