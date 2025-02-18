## What is this?
superluminal (lowercase) is a lightweight terminal content streaming application that enables
real-time sharing of a pseudoterminal (PTY) session with multiple clients. The server
manages the PTY, reading from its master side and efficiently forwarding all on-screen
content to connected clients over a custom TCP-based protocol using Protocol Buffers,
effectively streaming the terminal output with minimal overhead

superluminal includes a simple, minimal terminal user interface built with Bubble Tea,
allowing hosts to manage sessions seamlessly. Hosts can approve connections, remove users,
terminate sessions, and pause streaming at their discretion—all without compromising
session integrity.

## Setup
In order to get superluminal running you need the following;
* [go](https://go.dev/dl/)
* [git](https://git-scm.com/downloads)
* an [ngrok](https://ngrok.com/downloads/) account if you plan on sharing across the internet.
You can forego this if you have another way of hosting the session your self so long as you
can expose the port 42024 across the internet.

Run the following commands to clone and set up superluminal for you system.

``` sh
git clone https://github.com/willofdaedalus/superluminal
cd superluminal/
go build
```

### Starting and sharing your session as a host
To start a session after running the setup commands, do `./superluminal -s` to start a server.
If you didn't run the build command, it's `go run . -s` in the project directory to start a
server. This puts the passphrase at the top which you can share with anyone who wants to join
your session.

To make the session available over the internet, make sure you have `ngrok` installed (see above)
and follow the instructions to authenticate yourself. Once that's done run the following command
to expose superluminal's port to the internet; `ngrok tcp 127.0.0.1:42024`. This will launch
the ngrok interface in your terminal or you can check your browser on `localhost:4040` for
a URL that clients can join via. Share this URL and password with any clients you want
to join your session.

## Why is this?
superluminal is both a challenge and a learning experience for me—an opportunity to
explore the unconventional interfaces provided by the Linux operating system. I chose
Go for its simplicity and built-in concurrency, ensuring that I spend more time
understanding network programming rather than wrestling with syntax.

As I’ve worked on this, I’ve realized some potential use cases:
* a teaching aid for new programmers working in the terminal in situations on low-end
and bandwidth scenarios
* a tool for pair programming and remote collaboration

The project is still a work in progress, and I’ll be updating this README as it evolves.
- [x] Client Authentication
- [x] Pty Pipeline to Stream Applications to Clients
- [ ] Bubble Tea frontend
