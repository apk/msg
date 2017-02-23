This is a message server. You can put messages into it, and retrieve them.

Messages exist under a path (which is an array of strings), and have a timestamp.
The message itself is an arbitrary json object.

The timestamps are given by the server itself, and also serve as an index to
the stored messages.

Putting messages is done via HTTP POST or PUT requests, retrieving messages
happens via HTTP GET. Messages can also be obtained via websockets where
you subscribe to a set of paths, and the server pushes to you any new
messages that are put to these paths.

Messages can be stored (and usually are), and can be retrieved later.
There is also a method to obtain only the latest message for a subtree
of paths, and to DELETE the notion of the current message for a given
path. If messages are logged this deletion is logged and retrievable
as well.

Paths can be provided via messages on the websocket interface as well,
and set to be automatically deleted when the websocket connection closes.
