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
