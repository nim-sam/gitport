# GitPort

GitPort is a self-hosted Git server designed for local area networks (LANs), making it easy to share and collaborate on Git repositories without relying on external hosting services.

Point GitPort at a Git repository to serve it over SSH. It includes a built-in terminal UI (TUI) dashboard for managing access, browsing commits, and monitoring server activity.

**No separate web server, database, or DNS configuration required.** GitPort uses mDNS to make your server discoverable on the local network.

## Features

| Feature | Description |
| --- | --- |
| **Git over SSH** | Clone, fetch, pull, and push repositories over SSH using standard Git tools. |
| **SSH TUI dashboard** | Connect over SSH to manage user access, browse commits, and monitor server logs. |
| **Automatic LAN discovery** | Announces the server over mDNS, allowing other devices on the local network to discover it using a `.local` hostname. |
| **Non-invasive hosting** | Serves a bare repository, keeping your working tree untouched by remote pushes. |
| **Self-contained setup** | No separate web service, database server, or DNS configuration needed. |

## Installation

### Requirements

- [Go](<https://go.dev/dl/>)
- `git`
- `make`

### Build and install

```
git clone https://github.com/nim-sam/gitport.git
cd gitport
make install
```

## Getting Started

GitPort is initialized from the root of the Git repository you want to serve. You can create a new repository or use an existing one.

### 1\. Initialize a repository

For a new repository:

```
mkdir my-repo
cd my-repo
git init -q
```

If you already have a Git repository, simply navigate to its root directory.

### 2\. Initialize GitPort

Run the initialization command:

```
gitport init
```

You'll be prompted to configure the server, including its display name, host, port, and root administrator credentials.

<p align="center"> <img src="img/gitport_init_demo.gif" width="700" alt="GitPort initialization walkthrough"> </p>

### 3\. Start the server

From the repository root, run:

```
gitport start
```

GitPort will start serving the repository over SSH and announce its address on startup.

<p align="center"> <img src="img/gitport_start_demo.gif" width="700" alt="Starting the GitPort server"> </p>

### 4\. Clone the repository

Use the SSH address displayed at startup to access the repository from another machine on your LAN.

For a server running on port `2222`:

```
git clone ssh://git@<name>.local:2222/<repo>.git
```

For a server running on the default SSH port (`22`):

```
git clone git@<name>.local:<repo>.git
```

Replace `<name>` with your GitPort server's hostname and `<repo>` with the repository name.

Once cloned, use Git as usual to fetch, pull, and push changes.

## SSH Dashboard

GitPort includes a terminal UI dashboard accessible through SSH. Connect to your server to manage users, browse repository commits, and monitor server activity.

For a server using the default SSH port:

```
ssh <name>.local
```

For a server running on port `2222`:

```
ssh -p 2222 <name>.local
```

<p align="center"> <img src="img/gitport_ssh_demo.gif" width="1000" alt="GitPort SSH terminal dashboard"> </p>

## Notes

- **Protect server data:** Add `.gitport/` to your `.gitignore`. This directory contains GitPort's private SSH host key and user database. Do not commit or share these files.
- **LAN-only discovery:** mDNS works on the local network. If `<name>.local` does not resolve, connect using the server's IP address instead.