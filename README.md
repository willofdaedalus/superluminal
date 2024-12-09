## What is this?
superluminal (working title) tries to be a simple terminal streaming application
that starts a server which clients can connect to and create a pseudoterminal(pty)
which the host can read from and write to. The output from the master side of this
pty and forwards that over a custom protocol to each connected client effectively
"streaming" your terminal content to other clients.

## Why is this?
This is a challenge and a learning experience for myself on how to interact with the
unusual interfaces exposed to us by the Linux operating system. I chose Go specifically
for concurrency and simplicity to grok so that I don't end up losing a fight against
the language syntax while figuring out network programming the first time.  
  
Currently it's not ready yet but I'll be updating this README from time to time as
this project progresses.
