// this file outlines the "custom" protocol used by the server to authenticate
// new clients that connect and *try* to protect against malicious actors
// obviously it's very basic but should serve as minimal security for the servers
// running sessions
//
// Process
//   - once a client connects successfully, it awaits a message from the server that
//     indicates it's a "superluminal" server and not just any server. upon receive by
//     client, we await their information such as the passphrase and other exposed
//     client properties for verification
//   - server checks the client's pass and prompts the client up to three times before
//     before it rejects the client if they fail the passphrase verification. if they
//     pass verification, the server adds the new client to a map of already existing
//     clients

package server
