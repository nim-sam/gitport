package server

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	gitwish "github.com/charmbracelet/wish/git"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/logger"
	"github.com/nim-sam/gitport/internal/mdns"
	"github.com/nim-sam/gitport/internal/model"
)

// commitSyncMu serializes commit-table rebuilds. A Server value is copied
// during wiring, so the lock cannot live on the struct; one process-wide lock
// is enough as a push is the only writer.
var commitSyncMu sync.Mutex

// bareRepoName derives the served bare repository's directory name from the
// repository root folder's name. Characters that are unsafe in a single path
// segment are collapsed to hyphens, and the conventional .git suffix is
// guaranteed.
func (s Server) bareRepoName() string {
	var b strings.Builder

	for _, r := range strings.TrimSpace(filepath.Base(s.Root)) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}

	name := strings.Trim(b.String(), "-.")
	if name == "" {
		name = "repo"
	}
	if !strings.HasSuffix(name, ".git") {
		name += ".git"
	}
	return name
}

// bareRepoPath returns the on-disk location of the served bare repository. It
// lives inside .gitport, next to the config and database, so incoming pushes
// never touch the developer's working tree.
func (s Server) bareRepoPath() string {
	return filepath.Join(config.Dir(s.Root), s.bareRepoName())
}

// CloneAddress is the git clone URL clients use to reach the served bare
// repository over SSH. Authentication is by public key, so the user segment is
// only a conventional placeholder; the path names the served repository.
//
// The host is the mDNS alias announced on the LAN, so no DNS setup is needed.
// The conventional scp-like "git@host:path" form cannot carry a port, so it is
// used only on the standard SSH port; otherwise the explicit ssh:// form is
// returned.
func (s Server) CloneAddress() string {
	host := mdns.Label(s.Name) + ".local"

	port := s.Port
	if port == "" {
		port = defaultPort
	}
	if port == "22" {
		return fmt.Sprintf("git clone git@%s:%s", host, s.bareRepoName())
	}
	return fmt.Sprintf("git clone ssh://git@%s:%s/%s", host, port, s.bareRepoName())
}

// accessFor maps a database access level onto the git server's repo access.
// Owners, admins and read-write users may push; read-only users may only
// clone/fetch; anything else is denied. The capability gates live on the access
// level itself so the dashboard permission matrix cannot drift from this.
func accessFor(level model.AccessLevel) gitwish.AccessLevel {
	switch {
	case level.CanPush():
		return gitwish.ReadWriteAccess
	case level.CanClone():
		return gitwish.ReadOnlyAccess
	default:
		return gitwish.NoAccess
	}
}

// gitCommand reports whether the session was opened as a git transport command
// and, if so, which one. Only these commands are served by the git middleware;
// everything else (interactive sessions) falls through to the TUI.
func gitCommand(cmd []string) (string, bool) {
	if len(cmd) != 2 {
		return "", false
	}
	switch cmd[0] {
	case "git-receive-pack", "git-upload-pack", "git-upload-archive":
		return cmd[0], true
	default:
		return "", false
	}
}

// PrepareGit ensures the served bare repository exists, seeds it from the
// working tree the first time, and opens it for the commit history. It must be
// called after repository.Migrate and before InitCommits.
func (s *Server) PrepareGit() error {
	if err := gitwish.EnsureRepo(config.Dir(s.Root), s.bareRepoName()); err != nil {
		return err
	}

	if err := s.seedBareRepo(); err != nil {
		return err
	}

	repo, err := git.PlainOpen(s.bareRepoPath())
	if err != nil {
		return err
	}
	s.Repo = repo
	return nil
}

// seedBareRepo imports the working repository's branches and tags into the bare
// repo the first time it is served, so the server starts with the developer's
// existing history. Once the bare repo holds refs it is left untouched, since
// it is then the canonical copy clients push to.
func (s Server) seedBareRepo() error {
	bare, err := git.PlainOpen(s.bareRepoPath())
	if err != nil {
		return err
	}
	if hasRefs(bare) {
		return nil
	}

	work, err := git.PlainOpen(s.Root)
	if err != nil || !hasRefs(work) {
		return nil
	}

	cmd := exec.Command("git", "-C", s.Root, "push", s.bareRepoPath(),
		"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("could not seed %s: %w: %s", s.bareRepoPath(), err, out)
	}

	return setHeadToWorkingBranch(s.bareRepoPath(), s.Root)
}

// hasRefs reports whether the repository holds at least one concrete (hash)
// reference. The symbolic HEAD that go-git stores on init is ignored, so a
// freshly created bare repo counts as empty.
func hasRefs(repo *git.Repository) bool {
	refs, err := repo.References()
	if err != nil {
		return false
	}
	defer refs.Close()
	for {
		ref, err := refs.Next()
		if err != nil {
			return false
		}
		if ref.Name() == plumbing.HEAD {
			continue
		}
		if ref.Type() == plumbing.HashReference {
			return true
		}
	}
}

// setHeadToWorkingBranch points the bare repo's HEAD at whatever branch the
// working repo currently has checked out, so a fresh clone checks out the same
// branch rather than a non-existent default.
func setHeadToWorkingBranch(barePath, workPath string) error {
	work, err := git.PlainOpen(workPath)
	if err != nil {
		return err
	}
	head, err := work.Head()
	if err != nil {
		return nil
	}
	bare, err := git.PlainOpen(barePath)
	if err != nil {
		return err
	}
	return bare.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, head.Name()))
}

// ensureBareHead points HEAD at the first branch when it is unborn, which is
// the case after the first push into a freshly created bare repo (init sets
// HEAD to a default branch name that may not match what was pushed).
func (s Server) ensureBareHead() {
	bare, err := git.PlainOpen(s.bareRepoPath())
	if err != nil {
		return
	}
	if _, err := bare.Head(); err == nil {
		return
	}
	refs, err := bare.References()
	if err != nil {
		return
	}
	defer refs.Close()
	for {
		ref, err := refs.Next()
		if err != nil {
			return
		}
		if ref.Name().IsBranch() {
			_ = bare.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, ref.Name()))
			return
		}
	}
}

// gitMiddleware serves the canonical repository over the same SSH endpoint as
// the dashboard. Sessions opened with a git transport command are handled here;
// every other session falls through to the next handler (the TUI).
//
// Access is re-resolved from the database on each request using the connecting
// public key, so a role change applies immediately without a reconnect.
func (s Server) gitMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			sub, isGit := gitCommand(sess.Command())
			if !isGit {
				next(sess)
				return
			}

			user, ok := s.userForPublicKey(sess.PublicKey())
			access := gitwish.NoAccess
			if ok {
				access = accessFor(user.AccessLevel)
			}
			if access == gitwish.NoAccess {
				logger.Warn("git access denied", "user", userName(user, ok), "repo", s.Name, "op", sub)
				gitwish.Fatal(sess, gitwish.ErrNotAuthed)
				return
			}
			if sub == "git-receive-pack" && access < gitwish.ReadWriteAccess {
				logger.Warn("git access denied", "user", userName(user, ok), "repo", s.Name, "op", sub)
				gitwish.Fatal(sess, gitwish.ErrNotAuthed)
				return
			}

			if err := runGit(sess, s.bareRepoPath(), sub); err != nil {
				logger.Error("git command failed", "user", userName(user, ok), "repo", s.Name, "op", sub, "err", err)
				gitwish.Fatal(sess, gitwish.ErrSystemMalfunction)
				return
			}

			if sub == "git-receive-pack" {
				logger.Info("git push", "user", userName(user, ok), "repo", s.Name)
				if ok {
					recordPush(user)
				}
				s.afterPush()
			} else {
				// git-upload-pack covers both clone and pull; the wire command
				// does not distinguish them.
				logger.Info("git fetch (clone/pull)", "user", userName(user, ok), "repo", s.Name)
				if ok {
					recordFetch(user)
				}
			}
		}
	}
}

// userName returns the display name of a resolved user, or "unknown" when the
// session's key could not be matched to an account.
func userName(u model.User, ok bool) string {
	if !ok || u.Name == "" {
		return "unknown"
	}
	return u.Name
}

// runGit executes a git server subcommand against dir, wiring the session's
// stdio straight through so the git wire protocol flows over SSH.
func runGit(sess ssh.Session, dir, sub string) error {
	// git-upload-pack -> git upload-pack.
	cmd := exec.CommandContext(sess.Context(), "git", strings.TrimPrefix(sub, "git-"), dir)
	cmd.Stdin = sess
	cmd.Stdout = sess
	cmd.Stderr = sess.Stderr()
	return cmd.Run()
}

// afterPush fixes up the served repo's HEAD and re-syncs the commit history
// after a successful push so the dashboard reflects what was just received.
func (s Server) afterPush() {
	commitSyncMu.Lock()
	defer commitSyncMu.Unlock()

	s.ensureBareHead()
	s.RefreshCommits()
}
