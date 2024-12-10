## What is this?
superluminal (working title) tries to be a simple terminal streaming application
that starts a server which clients can connect to and creates a pseudoterminal(pty)
which the host can read from and write to. The server then reads from the master side
of this pty and forwards all content efficiently to all connected clients over custom
protocol on tcp effectively "streaming" your terminal contents.

## Why is this?
This is a challenge and a learning experience for myself on how to interact with the
unusual interfaces exposed to us by the Linux operating system. I chose Go specifically
for concurrency and simplicity to grok so that I don't end up losing a fight against
the language syntax while figuring out network programming the first time.  

The use cases for this as I've realised could be
* a teaching aid for new programmers in terminal
* peer programming
  
Currently it's not ready yet but I'll be updating this README from time to time as
this project progresses.

- [x] Client Authentication
- [x] Pty Pipeline to Stream Applications to Clients
- [ ] Bubble Tea frontend
