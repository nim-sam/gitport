# GitPort

GitPort is a lightweight CLI tool for self-hosting LAN-accessible Git repositories, enabling fast collaboration. Whether you want to host your repository on a server, or on your local machine, GitPort streamlines the nitty-gritty server making overhead into a few simple commands.

## Installation

Installing GitPort is as easy as 123. With these commands, you'll be able to build the executable and add it to your file path.

``` bash
git clone
cd gitport
sudo make install
```

And that's it! You can now run GitPort from any directory on your system.

## Usage

``` bash
# Initialize a new Git repository
mkdir new-repo
cd new-repo
git init
# Start a GitPort server
gitport start <port>
```

When running `gitport start <port>` for the first time on a given repo, you will be prompted with a few options to configure Git and SSH-level access permissions and will be automatically be saved.

Then, your current repository will be cloned into a *bare repository* in your system's `$CONFIG_DIR/gitport/` directory. This *bare repo* will now act as the server-side Git repository. Now, any changes comitted to your initial repository will point to that *bare repo*, as long as the server is running.

A `.../.gitport` folder will also be generated inside the *bare repo* to store all server-side data such as user permissions.

Future calls of `gitport start <port>` will allow you to pick up right where you left off without additional fuss.

## SSH TUI

The server can be monitored during its uptime with the help of a Terminal User Interface (TUI). Accessing this TUI doesn't require any additional installation, as it can be accessed over SSH. Only users with `admin` permissions can access the TUI. 

``` bash
# To access the TUI, simply run this command on any host on the local network. The TUI should appear in full screen if you have the correct credentials and access level.
ssh -p <port> <server_ip_addr>
```

On the TUI, server configurations such as user permissions (admin, ...) and repository-level edit access (read, write, ...) can be modified. *Note that a server reboot will be necessary for those changes to apply*. Additionally, the TUI displays the commit history and the respective diff's for each commit. Server-level logs can also be accessed directly on the TUI. Both the commit history and server-side logs can be filtered via a fuzzy finder.

## Honorable Mentions

During the development process, we were able to fully collaborate on this project with the help of GitPort upon the successful implementation of our first minimal viable product.
