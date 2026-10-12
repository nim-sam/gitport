// Package mdns announces <repo>.local on the LAN so `ssh <repo>.local` works
// without any DNS setup.
package mdns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"

	"github.com/nim-sam/gitport/internal/logger"
	"github.com/nim-sam/gitport/internal/util"
)

const (
	// keepaliveInterval re-announces the service even while no client is
	// connecting, so the A record never outlives its 120s TTL.
	keepaliveInterval = 90 * time.Second

	// minRefresh coalesces refresh bursts (for example several sessions
	// opening at once) so the announcement is not torn down and rebuilt
	// needlessly.
	minRefresh = 10 * time.Second
)

var invalidLabel = regexp.MustCompile(`[^a-z0-9-]+`)

// Label turns a repo name into the mDNS host label:
// "My Repo_1" -> "my-repo-1". The result is a valid DNS label
// (max 63 chars, no leading/trailing hyphen).
func Label(repo string) string {
	l := invalidLabel.ReplaceAllString(strings.ToLower(repo), "-")
	l = strings.Trim(l, "-")
	if l == "" {
		l = "repo"
	}
	if max := 63; len(l) > max {
		l = strings.TrimRight(l[:max], "-")
	}
	return l
}

// register publishes the hostname's A record and an _ssh._tcp service entry
// (visible in Bonjour/Avahi browsers).
func register(repo, label, ip string, port int) (*zeroconf.Server, error) {
	return zeroconf.RegisterProxy(label, "_ssh._tcp", "local.", port, label,
		[]string{ip}, []string{"app=gitport", "repo=" + repo}, nil)
}

// Announce publishes <repo>.local on the LAN and returns the hostname clients
// should use, a refresh func that re-announces it, and a cleanup func that
// withdraws it.
//
// zeroconf never answers host lookups and never re-announces on its own, and
// its A records carry a 120s TTL. The service is therefore re-announced on a
// keepalive timer (and, through refresh, whenever a client connects) so the
// alias stays resolvable.
func Announce(repo, port, ip string) (host string, refresh func(), cleanup func(), err error) {
	noop := func() {}

	p, err := strconv.Atoi(port)
	if err != nil {
		return "", noop, noop, fmt.Errorf("bad port %q: %w", port, err)
	}
	if ip == "" {
		ip = util.GetLocalIP()
	}

	label := Label(repo)

	cur, err := register(repo, label, ip, p)
	if err != nil {
		return "", noop, noop, err
	}

	var (
		mu   sync.Mutex
		last = time.Now()
	)

	refresh = func() {
		mu.Lock()
		defer mu.Unlock()

		if time.Since(last) < minRefresh {
			return
		}

		cur.Shutdown()

		next, err := register(repo, label, ip, p)
		if err != nil {
			logger.Warn("mDNS refresh failed", "name", label+".local", "err", err)
			return
		}
		cur = next
		last = time.Now()
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				refresh()
			}
		}
	}()

	logger.Info("mDNS announced", "name", label+".local", "ip", ip)

	cleanup = func() {
		close(done)
		wg.Wait()
		mu.Lock()
		defer mu.Unlock()
		cur.Shutdown()
	}

	return label + ".local", refresh, cleanup, nil
}
