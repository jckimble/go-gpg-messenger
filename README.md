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
 - Add Database Message Storage
 - Add status messages to websocket
 - Add Credentials to Database
 - Add MultiChat
 - Finish Shell Client
 - Add Configuration from files and commandline
 - Add optional dns server
 - Maybe add optional smtp proxy to bridge gap between users and the rest of the world
 - Create React Native App

### License
MIT
