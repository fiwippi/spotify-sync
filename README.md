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
```console
$ ./spotify_sync --help
Usage:
   [command]

Available Commands:
  client      Runs the client
  help        Help about any command
  server      Runs the server
  view        Views the server database

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```
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

## Server Setup
# This should be the domain which the server operates on, if needed then
# port numbers should also be specified, e.g. if you used port forwarding
# without port 80. 
DOMAIN=localhost:8096
# Whether the redirect url should be http or https
USE_SSL=true
# The port for the server to run on locally
PORT=8096
# Run the server in Gin debug or release mode
SERVER_MODE=debug
# What level to log at 
SERVER_LOG_LEVEL=trace
# The "Admin Key" should be kept solely by the server owner, this is used
# to edit user data and create and delete users, accesible at /admin
ADMIN_KEY=abcdefghijklmnopqrstuvwxyz123456
```
**Additionally**, inside the spotify developer portal for your application, you should add your domain route followed
by `/spotify-callback` as a valid callback URL, for example: `localhost:8096/spotify-callback`. This is used by the
server to create a spotify client which can control the user playback.

To see the data stored within the database run `spotify_sync view`. The db file can only be used by one program at once so the server should not run at the same time. Alternatively the server provides the `/admin` route to access the admin functionality and the ability to view the database.

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

## Docker
- If running the provided docker image, the port will always be 8096.


## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
`MIT`