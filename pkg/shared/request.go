package ws

// Http Request which the client can send to the server to perform admin duties (account editing)
type Request struct {
	CurrentName string `json:"current_name"` // Name of the user to act upon
	NewName     string `json:"new_name"`     // Name of new user or name to change username to
	NewPassword string `json:"new_password"` // Password of new user or password to change user's pass to
	ServerKey   string `json:"server_key"`   // Server key to authorise basic requests
	AdminKey    string `json:"admin_key"`    // Server key to authorise privileged requests
}

// Http Response which the server sends to the client in response to a request
type Response struct {
	Success bool   `json:"success"` // Whether the request was a success
	Error   string `json:"error"`   // The error msg if the request failed
}