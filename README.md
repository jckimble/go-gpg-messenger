# go-gpg-messenger
go-gpg-messenger is an attempt at a fully distributed encrypted message system
### Installation
Install the dependencies and devDependencies and start the server.
```sh
$ export GOPATH=go-gpg-messenger #change to the path
$ cd go-gpg-messenger
$ make depends server
$ ./server
```

### Todos (What needs to be done before I consider this a 1.0 release)
 - Write Tests
 - Add status messages to websocket
 - Read Notifications
 - Add MultiChat
 - Finish Shell Client
 - Create React Native App

### Random Thoughts (Up for discussion)
 - optional smtp proxy to bridge gap between users and the rest of the world
 - Epithermal Messages (Only Stored in Database Until Downloaded)
 - Auto-Delete Timer for Messages

### License
MIT
