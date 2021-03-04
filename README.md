# Spotify-Sync
## Overview [![GoDoc](https://godoc.org/github.com/fiwippi/spotify-sync?status.svg)](https://godoc.org/github.com/fiwippi/spotify-sync)
A terminal interface server and client combo which syncs the spotify playback of multiple clients towards a singular host

Note: This only works for **Spotify Premium** users due to API limitations

## Install
```
go get github.com/fiwippi/spotify-sync
```
Then `cd` into the directory and `make`

## Usage
### Server
The server performs the syncing operations on the client's connected to it. Once clients are connected with the server 
they can create their own spotify sessions which other users can join. Some `env` variables must be specified 
for the server to run:
```dotenv
## Spotify Setup, applications created at https://developer.spotify.com/
# The ID of the spotify application
SPOTIFY_ID=abcdefghijklmnopqrstuvwxyz123456
# The secret of the spotify application
SPOTIFY_SECRET=abcdefghijklmnopqrstuvwxyz123456
# How often in seconds to send requests to the spotify api to sync clients, 
# a lower value increases the risk of rate limiting being applied 
SYNC_REFRESH=10

## Server Keys
# The "Server Key" is given to users so that they can create their own 
# account
SERVER_KEY=abcdefghijklmnopqrstuvwxyz123456
# The "Admin Key" should be kept solely by the server owner, this is used
# to edit user data and delete users
ADMIN_KEY=abcdefghijklmnopqrstuvwxyz123456

## Server Setup
# This should be the domain which the server operates on, if needed then
# port numbers should also be specified, e.g. if you used port forwarding
# without port 80. 
DOMAIN=localhost:8096
# The port for the server to run on locally
PORT=8096
```
**Additionally**, inside the spotify developer portal for your application, you should add your domain route followed
by `/spotify-callback` as a valid callback URL, for example: `localhost:8096/spotify-callback`. This is used by the
server to create a spotify client which can control the user playback.

To see the data stored within the database the `spotify_sync_view` binary should be used, it should be in the same 
directory as the created `spotify.db` file. The db file can only be used by one program at once so the server should not run at the same time. 

### Clients
Clients can perform certain operations by typing in the chat box provided after they connect to the server,
the argument fields for commands are separated by a comma:
```dotenv
CREATE = Create a session
JOIN = Join a session someone has created e.g. "join,username"
EXIT/QUIT = Disconnect from the server
DISCONNECT = Leave the session
ID = Displays the ID of the current session
MSG = Send a message to other users in the same session e.g. "msg,change the song?""`
```
The client also provided functionality to connect with the server and create, update or delete user accounts. 
This is authenticated with the Server and Admin keys where the Server Key can only authenticate the creation of
accounts whereas the Admin Key can authenticate creation, deletion or updating. 

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

MIT