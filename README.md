# Rust backend

This is a modified Golang backend, ported to Rust by [cat](https://matrix.to/#/@cat:plan9.rocks), of a certain website I run.
Published as demonstration of my more recent Golang projects. 

This is suitable as a blog generator, though it was built to support dynamic 
content as well. This project is work in progress.

## Features
- Built for dynamic content and performance
- Utilizes Rust
- No preprocessing or hardcoding pages, everything is searched, built and cached as requested
- Serverside Markdown-to-HTML conversion and syntax highlighting
- Supports templating functions even in Markdown
- Rudimentary user accounts (SQLite, bcrypt), very work in progress though!

![Picture](./screenshot.png?raw=true)
I'm not a designer

## Dependencies
This is intended to be published behind an Nginx reverse proxy!

### Rust
- Cargo will handle it for you!

### Node.js
~~No decent syntax highlighting libraries exist for Golang (at least back during project creation),
so this part is "temporarily" handled by Node.js~~

Fixed by Rust <3
