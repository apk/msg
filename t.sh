/usr/bin/curl -d '{"d":'"`date +%s`"',"k":"2"}' http://localhost:3046/msg/in/here/there
/usr/bin/curl -d '{"d":'"`date +%s`"',"k":"1"}' http://localhost:3046/msg/in/here
/usr/bin/curl -d '{"d":'"`date +%s`"'}' http://localhost:3046/msg/in
