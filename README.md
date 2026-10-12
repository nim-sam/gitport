# GitPort

GitPort is a LAN self-hosted Git server, ideal for team collaboration over LAN or homelabbing.

Point it at any Git repository and it serves it over SSH, with a built-in TUI dashboard for monitoring server activities. There is no separate web service, database server or DNS setup needed to run.

## Features

|Feature|Description|
| :--- | :--- |
| Git over SSH | Clone, fetch and push through the same SSH endpoint as the dashboard. |
| SSH TUI dashboard | Connection with `ssh` to manage user access, browse commits and monitor server logs. |
| Zero-config Discovery | The server annoucnes itself over mDNS, so users on the LAN can reach it. |
| Non-invasive | Pushes are directed to a bare repository, leaving your working tree untouched. |


## Server over SSH TUI

Connecting to the server via the SSH session opens the TUI dashboard:

```bash
ssh <name>.local

# or, on a non-default port
ssh -p 2222 <name>.local
```

## Installation

Requirements: [Go](https://go.dev/dl/), `git` and `make`.

```bash
git clone https://github.com/nim-sam/gitport.git
cd gitport
make install
```

## Usage

Run these from the root of the Git repository you want to serve.

```bash
# Create new Git repo (or use an already initialized one)
mkdir my-repo
cd my-repo
git init -q

# Initialize your GitPort server
# 
# Users will be prompted to enter display name, host, port and root admin credentials.
gitport init
```

<p align="center">
  <img src="img/gitport_init_demo.gif" width="700">
</p>

```
# Start your GitPort server
gitport start
```

<p align="center">
  <img src="img/gitport_start_demo.gif" width="700">
</p>

Once the server is running, the repo can be cloned with the address shown on startup:

```bash
git clone ssh://git@<name>.local:2222/<repo>.git
# on the default port 22
git clone git@<name>.local:<repo>.git
```

## Notes

- Add `.gitport/` to your `.gitignore`. It contains the server's private host key and user database.
- mDNS discovery works on the local network only. If `<name>.local` doesn't resolve, connect by IP address.