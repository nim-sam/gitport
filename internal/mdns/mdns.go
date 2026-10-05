// Package mdns announces the GitPort server on the local network as
// <repo>-gitport.local so `ssh <repo>-gitport.local` works with no DNS setup.
package mdns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/grandcat/zeroconf"

	"github.com/nim-sam/gitport/internal/util"
)

const suffix = "-gitport"

var invalidLabel = regexp.MustCompile(`[^a-z0-9-]+`)

// Label turns a repo name into the mDNS host label:
// "My Repo_1" -> "my-repo-1-gitport". The result is a valid DNS label
// (max 63 chars, no leading/trailing hyphen on the repo part).
func Label(repo string) string {
	l := invalidLabel.ReplaceAllString(strings.ToLower(repo), "-")
	l = strings.Trim(l, "-")
	if l == "" {
		l = "repo"
	}
	if max := 63 - len(suffix); len(l) > max {
		l = strings.TrimRight(l[:max], "-")
	}
	return l + suffix
}

// Announce publishes <repo>-gitport.local on the LAN and returns the
// hostname clients should use plus a cleanup func that withdraws it.
func Announce(repo, port, ip string) (host string, cleanup func(), err error) {
	noop := func() {}

	p, err := strconv.Atoi(port)
	if err != nil {
		return "", noop, fmt.Errorf("bad port %q: %w", port, err)
	}
	if ip == "" {
		ip = util.GetLocalIP()
	}

	label := Label(repo)

	// RegisterProxy publishes the hostname's A record and an _ssh._tcp
	// service entry (visible in Bonjour/Avahi browsers).
	srv, err := zeroconf.RegisterProxy(label, "_ssh._tcp", "local.", p, label,
		[]string{ip}, []string{"app=gitport", "repo=" + repo}, nil)
	if err != nil {
		return "", noop, err
	}

	log.Info("mDNS announced", "name", label+".local", "ip", ip)
	return label + ".local", srv.Shutdown, nil
}