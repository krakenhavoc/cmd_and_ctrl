package lobby

// bugforum.go: the in-app bug report's mirror in a Discord forum
// channel (ADR 0196, #3327). A report that was filed as a GitHub issue
// also opens a forum post, so a player with no GitHub account can see
// what became of it and answer a question; when the issue closes, the
// post gets a note and is archived.
//
// The SERVER posts, not the gateway bot binary (ADR 0196 §2): it holds
// the report, the bot token and the GitHub client, the bot binary is a
// separate deploy that must never run from develop, and the results feed
// and the #lfg board already post this way. The close needs no public
// webhook (§4): a sweep reads the open posts the bot made and asks
// GitHub whether each one's issue is still open.
//
// Nothing here is on the report's request path. The POST /bugreport
// handler has already answered the reporter by the time a thread is
// attempted, a Discord outage costs a missing thread and a log line, and
// nothing in this file can fail a report.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
)

// The ADR's numbers (§4, §5). Constants, not settings.
const (
	// bugForumMaxAttempts is how many transient failures in a row a new
	// thread tolerates before it is dropped: the results feed's count.
	bugForumMaxAttempts = 3
	// bugForumCallTimeout bounds one Discord call.
	bugForumCallTimeout = 10 * time.Second
	// bugForumMaxRetryWait caps a 429's wait, and bugForumBackoff is the
	// wait after a 5xx or a timeout.
	bugForumMaxRetryWait = 30 * time.Second
	bugForumBackoff      = 2 * time.Second
	// bugForumSweepEvery is how often open threads are checked against
	// their issues, and bugForumFirstSweep how soon after boot the first
	// check runs, so an issue closed while the server was down is caught
	// without waiting a full period.
	bugForumSweepEvery = 5 * time.Minute
	bugForumFirstSweep = 30 * time.Second
	// bugForumMaxChecks caps the GitHub reads one sweep makes. A forum
	// with more open posts than that is checked over several sweeps, in
	// the order Discord lists them, which is not a starvation risk: a
	// closed issue's post leaves the list.
	bugForumMaxChecks = 50
	// bugForumMaxInFlight bounds how many threads are being created at
	// once. The report throttle (bugreport_gate.go) already caps filings
	// per account and per day; this is the second wall, so a burst of
	// slow Discord calls cannot pile up goroutines.
	bugForumMaxInFlight = 4
)

// bugForumAPI is the Discord half the mirror uses: discord.BugForum, or
// a fake in tests.
type bugForumAPI interface {
	CreateThread(ctx context.Context, client *http.Client, t discord.BugThread) (string, error)
	ActiveThreads(ctx context.Context, client *http.Client) ([]discord.ForumThread, error)
	PostNote(ctx context.Context, client *http.Client, threadID, text string) error
	Archive(ctx context.Context, client *http.Client, threadID string) error
}

// BugIssueReader reads an issue's state. *github.Client satisfies it,
// both as the PAT client and as the GitHub App's.
type BugIssueReader interface {
	GetIssue(ctx context.Context, number int) (github.Issue, error)
}

// BugForum mirrors filed reports into a Discord forum. Construct it with
// NewBugForum, put it in Config.BugForum and Start it. A nil *BugForum
// is the feature off: every method is a no-op on it.
type BugForum struct {
	api    bugForumAPI
	issues BugIssueReader
	client *http.Client
	log    *slog.Logger
	// wait sleeps between attempts; a test replaces it.
	wait func(time.Duration)

	slots chan struct{}
	wg    sync.WaitGroup

	// closing remembers what the sweep has already done to a thread:
	// noted (the closing note is posted, the archive is not yet done, so
	// a retry does not post the note twice) or done (so a thread someone
	// reopens to ask a late question is left alone until its issue is
	// reopened and closed again). One small entry per closed report, in
	// memory on purpose: after a restart the worst case is that a thread
	// reopened since is noted and archived once more.
	mu      sync.Mutex
	closing map[string]closeState
}

type closeState int

const (
	closeNoted closeState = iota + 1
	closeDone
)

// NewBugForum builds the mirror over the configured forum and a source of
// issue state.
func NewBugForum(f discord.BugForum, issues BugIssueReader, log *slog.Logger) *BugForum {
	return newBugForum(f, issues, &http.Client{Timeout: bugForumCallTimeout}, log)
}

func newBugForum(api bugForumAPI, issues BugIssueReader, client *http.Client, log *slog.Logger) *BugForum {
	if log == nil {
		log = slog.Default()
	}
	return &BugForum{
		api: api, issues: issues, client: client, log: log,
		wait:    time.Sleep,
		slots:   make(chan struct{}, bugForumMaxInFlight),
		closing: map[string]closeState{},
	}
}

// bugMirror is a filed report, as the mirror needs it. The lobby builds
// it from the same redacted text the issue carries.
type bugMirror struct {
	Number      int
	URL         string
	Title       string
	Kind        string
	Description string
	ReporterID  string
	Images      []string
}

// Mirror opens a thread for a filed report, in the background. It never
// blocks and never fails the caller; a thread that cannot be made is
// logged and forgotten.
func (f *BugForum) Mirror(m bugMirror) {
	if f == nil {
		return
	}
	select {
	case f.slots <- struct{}{}:
	default:
		f.log.Warn("bug forum: too many threads being opened; this report is not mirrored", "issue", m.Number)
		return
	}
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		defer func() { <-f.slots }()
		f.create(m)
	}()
}

// Wait blocks until every thread being opened has finished. A test's
// handle on the background work.
func (f *BugForum) Wait() {
	if f != nil {
		f.wg.Wait()
	}
}

// create makes one thread: up to bugForumMaxAttempts tries while Discord
// is rate-limiting or not answering, and no more for any other error.
func (f *BugForum) create(m bugMirror) {
	t := discord.BugThread{
		IssueNumber: m.Number, IssueURL: m.URL, Title: m.Title, Kind: m.Kind,
		Description: m.Description, ReporterID: m.ReporterID, Images: m.Images,
	}
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), bugForumCallTimeout)
		id, err := f.api.CreateThread(ctx, f.client, t)
		cancel()
		if err == nil {
			f.log.Info("bug forum: thread opened", "issue", m.Number, "thread", id)
			return
		}
		if !discord.Transient(err) || attempt >= bugForumMaxAttempts {
			f.log.Warn("bug forum: thread not opened", "issue", m.Number, "attempt", attempt, "err", err)
			return
		}
		f.wait(retryWait(err))
	}
}

// retryWait is how long to wait before another attempt at a transient
// failure: Discord's own retry_after when it gave one, capped.
func retryWait(err error) time.Duration {
	var rl *discord.RateLimitError
	if errors.As(err, &rl) {
		return min(rl.RetryAfter, bugForumMaxRetryWait)
	}
	return bugForumBackoff
}

// Start runs the close sweep until ctx ends.
func (f *BugForum) Start(ctx context.Context) {
	if f == nil {
		return
	}
	go func() {
		timer := time.NewTimer(bugForumFirstSweep)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				f.Sweep(ctx)
				timer.Reset(bugForumSweepEvery)
			}
		}
	}()
}

// Sweep closes the thread of every report whose issue is closed (or
// gone). It reads the forum's open posts that the bot made, asks GitHub
// about at most bugForumMaxChecks of them, and for each closed issue
// posts one note and archives the thread. A failure is logged and the
// thread is simply looked at again next time.
func (f *BugForum) Sweep(ctx context.Context) {
	if f == nil {
		return
	}
	lctx, cancel := context.WithTimeout(ctx, bugForumCallTimeout)
	threads, err := f.api.ActiveThreads(lctx, f.client)
	cancel()
	if err != nil {
		f.log.Warn("bug forum: could not list open threads", "err", err)
		return
	}
	if len(threads) > bugForumMaxChecks {
		threads = threads[:bugForumMaxChecks]
	}
	for _, th := range threads {
		if ctx.Err() != nil {
			return
		}
		n := discord.IssueNumberFromThreadName(th.Name)
		gone, closed, err := f.issueClosed(ctx, n)
		if err != nil {
			f.log.Warn("bug forum: could not read the issue", "issue", n, "err", err)
			continue
		}
		if !closed {
			// Open, or reopened: a thread closed before may close again.
			f.setState(th.ID, 0)
			continue
		}
		if f.state(th.ID) == closeDone {
			continue
		}
		if err := f.closeThread(ctx, th.ID, n, gone); err != nil {
			f.log.Warn("bug forum: could not close the thread", "issue", n, "thread", th.ID, "err", err)
			if discord.Transient(err) {
				// Discord is rate-limiting this bot: the rest wait for
				// the next sweep rather than make it worse.
				return
			}
		}
	}
}

// issueClosed reports whether the issue is closed, and whether it is
// closed because it no longer exists.
func (f *BugForum) issueClosed(ctx context.Context, number int) (gone, closed bool, err error) {
	rctx, cancel := context.WithTimeout(ctx, bugForumCallTimeout)
	defer cancel()
	iss, err := f.issues.GetIssue(rctx, number)
	switch {
	case err == nil:
		return false, !iss.Open(), nil
	case upstreamGone(err):
		return true, true, nil
	default:
		return false, false, err
	}
}

// closeThread posts the closing note once and archives the thread.
func (f *BugForum) closeThread(ctx context.Context, threadID string, number int, gone bool) error {
	if f.state(threadID) != closeNoted {
		text := fmt.Sprintf("GitHub issue #%d is closed, so this thread is closed too. Thanks for reporting it.", number)
		if gone {
			text = fmt.Sprintf("GitHub issue #%d was removed, so this thread is closed. Thanks for reporting it.", number)
		}
		nctx, cancel := context.WithTimeout(ctx, bugForumCallTimeout)
		err := f.api.PostNote(nctx, f.client, threadID, text)
		cancel()
		if err != nil {
			return err
		}
		f.setState(threadID, closeNoted)
	}
	actx, cancel := context.WithTimeout(ctx, bugForumCallTimeout)
	defer cancel()
	if err := f.api.Archive(actx, f.client, threadID); err != nil {
		return err
	}
	f.setState(threadID, closeDone)
	return nil
}

func (f *BugForum) state(id string) closeState {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closing[id]
}

// setState records a thread's state; 0 forgets it.
func (f *BugForum) setState(id string, st closeState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if st == 0 {
		delete(f.closing, id)
		return
	}
	f.closing[id] = st
}

// bugMirrorable reports whether a report may be mirrored (ADR 0196 §3).
//
// A report filed outside a game (the lobby, the deck pages) may. A report
// filed at a table may only when the host has put that table on the
// public #lfg board: every table is otherwise invite-only, and the
// report text, the screenshots and the reporter's name would be shown to
// a community server's members who were never invited to it. Anything
// that makes the table's publicness unknowable counts as private: a game
// id that does not parse, a table the lobby no longer holds, a practice
// table, no lobby to ask.
func (c Config) bugMirrorable(p auth.Principal, bctx *bugReportContext) bool {
	var ids []uuid.UUID
	if bctx != nil && bctx.GameID != "" {
		id, err := uuid.Parse(bctx.GameID)
		if err != nil {
			return false
		}
		ids = append(ids, id)
	}
	if p.GameID != uuid.Nil {
		ids = append(ids, p.GameID)
	}
	for _, id := range ids {
		if c.Lobby == nil {
			return false
		}
		meta, err := c.Lobby.Get(id)
		if err != nil || meta.Practice || !meta.LFG {
			return false
		}
	}
	return true
}
