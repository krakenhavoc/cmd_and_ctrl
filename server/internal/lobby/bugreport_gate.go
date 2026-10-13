package lobby

// bugreport_gate.go — the throttle in front of GitHub for the in-app
// bug button (ADR 0017 §10, #2979).
//
// Before a public launch the button is a way for any signed-in visitor
// to make the server's token write to the project tracker. The IP
// bucket in front of POST /bugreport stops a stuck retry loop, not a
// determined or merely numerous crowd, so the gate adds:
//
//   - a per-account window (one session's account, however many
//     addresses it comes from),
//   - a global per-day cap on issues opened,
//   - duplicate detection that adds a +1 comment to the issue a recent
//     report already opened instead of opening another,
//   - an admin kill switch that stops every GitHub write at once.
//
// Over a limit, or while paused, a report is NOT refused: it is kept in
// the bug store with the issue it would have opened, and an admin lists
// the queue at GET /admin/bugreport. The reporter gets a success whose
// body says plainly that it was queued, not filed.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
)

const (
	// bugPlayerLabel is on every issue the button files, whatever its
	// kind, so triage can filter the public's reports from the
	// maintainers' own. It must exist on the repo (ADR 0017 §10).
	bugPlayerLabel = "player-report"

	// Defaults for BugLimits. Chosen for a table of friends that has
	// opened to the public: five reports an hour is more than anyone
	// writes by hand, and fifty a day is more than triage reads.
	defaultBugAccountMax    = 5
	defaultBugAccountWindow = time.Hour
	defaultBugDailyCap      = 50
	defaultBugDedupeWindow  = 24 * time.Hour

	// bugRecentMax bounds the recent-filing index. Entries older than
	// the dedupe window are dead weight, but a flood inside the window
	// must not grow the file without bound.
	bugRecentMax = 200

	// bugQueueListMax is how many queued reports GET /admin/bugreport
	// returns, newest first.
	bugQueueListMax = 100
)

// BugLimits tunes the gate. A zero field means its default, so a
// deployment sets nothing and a test sets only what it exercises.
type BugLimits struct {
	// AccountMax reports per account per AccountWindow before the
	// next one is queued.
	AccountMax    int
	AccountWindow time.Duration
	// DailyCap issues opened per UTC day, across every account.
	DailyCap int
	// DedupeWindow is how long a filed issue absorbs duplicates.
	DedupeWindow time.Duration
}

func (l BugLimits) withDefaults() BugLimits {
	if l.AccountMax <= 0 {
		l.AccountMax = defaultBugAccountMax
	}
	if l.AccountWindow <= 0 {
		l.AccountWindow = defaultBugAccountWindow
	}
	if l.DailyCap <= 0 {
		l.DailyCap = defaultBugDailyCap
	}
	if l.DedupeWindow <= 0 {
		l.DedupeWindow = defaultBugDedupeWindow
	}
	return l
}

// BugCommenter is the optional second half of the BugReporter seam:
// adding a comment to an existing issue. *github.Client satisfies it.
// A reporter that does not simply never dedupes — every report opens
// its own issue, as before.
type BugCommenter interface {
	CreateComment(ctx context.Context, number int, body string) (url string, err error)
}

type bugRoute int

const (
	bugRouteFile bugRoute = iota
	bugRouteComment
	bugRouteQueue
)

const (
	bugReasonPaused  = "paused"
	bugReasonAccount = "account limit"
	bugReasonDaily   = "daily cap"
)

type bugDecision struct {
	route  bugRoute
	reason string          // queue only
	issue  bugstore.Filing // comment only
}

// bugGate is the in-process state of the throttle: the persisted
// bugstore.GateState (kill switch, today's count, recent filings) and
// the per-account windows, which are deliberately memory-only — a
// restart forgetting who reported in the last hour costs one extra
// report each, and the IP bucket and the daily cap both stand.
type bugGate struct {
	mu       sync.Mutex
	store    *bugstore.Store
	log      *slog.Logger
	lim      BugLimits
	st       bugstore.GateState
	accounts map[string][]time.Time
}

func newBugGate(c Config) *bugGate {
	g := &bugGate{
		store:    c.BugStore,
		log:      c.Log,
		lim:      c.BugLimits.withDefaults(),
		accounts: map[string][]time.Time{},
	}
	st, err := c.BugStore.LoadGate()
	if err != nil {
		// Unreadable state starts empty rather than refusing to boot:
		// the worst case is a forgotten pause, which an admin sees on
		// GET /admin/bugreport.
		logBugStoreWarning(c, "load throttle state", err)
	}
	g.st = st
	return g
}

// persistLocked writes the state. A failure is logged, not returned:
// the in-memory state still governs this process, and a report must
// never be lost because a counter file would not save.
func (g *bugGate) persistLocked() {
	if err := g.store.SaveGate(g.st); err != nil && g.log != nil {
		g.log.Warn("bugreport: save throttle state failed", "err", err)
	}
}

func (g *bugGate) rollDayLocked(now time.Time) {
	if day := now.UTC().Format("2006-01-02"); g.st.Day != day {
		g.st.Day, g.st.Filed = day, 0
	}
}

// decide picks what happens to a report and reserves what that costs:
// an account slot for a filing or comment, a daily-cap slot for a
// filing. The caller owes a release if the work then fails.
//
// Order is the point. Paused beats everything (the kill switch is
// "no GitHub writes"), then the account window, then the duplicate
// check — before the daily cap, so a +1 still lands after the cap is
// reached — then the cap.
func (g *bugGate) decide(now time.Time, acct string, keys []string, canComment bool) bugDecision {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rollDayLocked(now)

	if g.st.Paused {
		return bugDecision{route: bugRouteQueue, reason: bugReasonPaused}
	}
	cutoff := now.Add(-g.lim.AccountWindow)
	var recent []time.Time
	for _, t := range g.accounts[acct] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= g.lim.AccountMax {
		g.accounts[acct] = recent
		return bugDecision{route: bugRouteQueue, reason: bugReasonAccount}
	}
	if canComment {
		if f, ok := g.matchLocked(now, keys); ok {
			g.accounts[acct] = append(recent, now)
			return bugDecision{route: bugRouteComment, issue: f}
		}
	}
	if g.st.Filed >= g.lim.DailyCap {
		g.accounts[acct] = recent
		return bugDecision{route: bugRouteQueue, reason: bugReasonDaily}
	}
	g.accounts[acct] = append(recent, now)
	g.st.Filed++
	g.persistLocked()
	return bugDecision{route: bugRouteFile}
}

// matchLocked finds a filing inside the dedupe window that shares a
// key with the report. Newest first, so a re-filed issue wins over the
// one it replaced.
func (g *bugGate) matchLocked(now time.Time, keys []string) (bugstore.Filing, bool) {
	if len(keys) == 0 {
		return bugstore.Filing{}, false
	}
	for i := len(g.st.Recent) - 1; i >= 0; i-- {
		f := g.st.Recent[i]
		if now.Sub(f.At) > g.lim.DedupeWindow {
			continue
		}
		for _, k := range keys {
			for _, fk := range f.Keys {
				if k == fk {
					return f, true
				}
			}
		}
	}
	return bugstore.Filing{}, false
}

// release gives back what decide reserved when the GitHub call (or the
// artifact build before it) failed, so an outage does not eat a
// reporter's allowance or the day's cap.
func (g *bugGate) release(now time.Time, acct string, filed bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ts := g.accounts[acct]
	for i := len(ts) - 1; i >= 0; i-- {
		if ts[i].Equal(now) {
			g.accounts[acct] = append(ts[:i:i], ts[i+1:]...)
			break
		}
	}
	if filed && g.st.Filed > 0 {
		g.st.Filed--
		g.persistLocked()
	}
}

// recordFiled adds an opened issue to the dedupe index.
func (g *bugGate) recordFiled(now time.Time, keys []string, number int, url string) {
	if len(keys) == 0 {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.st.Recent = append(g.st.Recent, bugstore.Filing{Keys: keys, Number: number, URL: url, At: now})
	cutoff := now.Add(-g.lim.DedupeWindow)
	kept := g.st.Recent[:0]
	for _, f := range g.st.Recent {
		if f.At.After(cutoff) {
			kept = append(kept, f)
		}
	}
	if len(kept) > bugRecentMax {
		kept = kept[len(kept)-bugRecentMax:]
	}
	g.st.Recent = kept
	g.persistLocked()
}

// forget drops an issue from the index — it was deleted on GitHub, so
// a +1 to it cannot land.
func (g *bugGate) forget(number int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	kept := g.st.Recent[:0]
	for _, f := range g.st.Recent {
		if f.Number != number {
			kept = append(kept, f)
		}
	}
	g.st.Recent = kept
	g.persistLocked()
}

func (g *bugGate) setPaused(now time.Time, paused bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.st.Paused == paused {
		return
	}
	g.st.Paused = paused
	g.st.PausedAt = time.Time{}
	if paused {
		g.st.PausedAt = now.UTC()
	}
	g.persistLocked()
}

// bugAccountKey names the account a report counts against: the signed-
// in user when there is one, else the Discord identity, else the
// session's seat, else the admin login. A guest can mint new seats, so
// that tier is the weak one — the IP bucket and the daily cap are what
// bound a guest.
func bugAccountKey(p auth.Principal) string {
	switch {
	case p.UserID != uuid.Nil:
		return "user:" + p.UserID.String()
	case p.DiscordID != "":
		return "discord:" + p.DiscordID
	case p.PlayerID != uuid.Nil:
		return "player:" + p.PlayerID.String()
	case p.AdminID != uuid.Nil:
		return "admin:" + p.AdminID.String()
	default:
		return "role:" + string(p.Role)
	}
}

var (
	bugNonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	bugUUIDRe  = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	bugDigitRe = regexp.MustCompile(`[0-9]+`)
)

// bugDedupeKeys are the strings a later report is matched on, all
// scoped to the kind (the same words as a bug and as an idea are two
// issues):
//
//   - the normalised title: lower case, runs of anything that is not a
//     letter or digit folded to one space;
//   - the stack: the first client-side error line in the log, with ids
//     and numbers masked so a different game's copy of the same
//     exception still matches. Only the kinds that carry the log have
//     one.
//
// There is no card key: no card reaches a report (the context holds
// game, turn, phase, step, seq and connection only).
func bugDedupeKeys(kind bugKind, title string, log []bugLogEntry) []string {
	var keys []string
	if t := strings.TrimSpace(bugNonWord.ReplaceAllString(strings.ToLower(title), " ")); t != "" {
		keys = append(keys, "t:"+kind.Label+"|"+clipRunes(t, 120))
	}
	if kind.Log {
		for _, e := range log {
			if e.Kind != "error" {
				continue
			}
			sig := bugDigitRe.ReplaceAllString(bugUUIDRe.ReplaceAllString(strings.ToLower(e.Text), "#"), "#")
			sig = strings.Join(strings.FieldsFunc(sig, unicode.IsSpace), " ")
			// A line this short ("error", "failed") would merge
			// unrelated reports.
			if len(sig) >= 16 {
				keys = append(keys, "e:"+kind.Label+"|"+clipRunes(sig, 160))
			}
			break
		}
	}
	return keys
}

func clipRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// renderBugPlusOne is the comment added to an existing issue. Short on
// purpose: the screenshots, log and pins of a duplicate are not kept
// (the issue it joins already has its own), only who saw it again and
// what they said.
func renderBugPlusOne(in bugIssue) string {
	in = redactBugIssue(in)
	var b strings.Builder
	fmt.Fprintf(&b, "+1 — another %s report of this.\n\n", in.Kind.Noun)
	if d := strings.TrimSpace(in.Desc); d != "" {
		for _, line := range strings.Split(clipRunes(d, 1000), "\n") {
			b.WriteString("> " + strings.TrimRight(line, "\r") + "\n")
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "- Reporter: %s (%s)\n", clip(reporterName(in.Principal), bugFieldMax), in.Principal.Role)
	if in.Ctx != nil && in.Ctx.GameID != "" {
		fmt.Fprintf(&b, "- Game: `%s`", clip(in.Ctx.GameID, bugFieldMax))
		if in.Ctx.Turn > 0 {
			fmt.Fprintf(&b, " — turn %d, %s/%s", in.Ctx.Turn, clip(in.Ctx.Phase, bugFieldMax), clip(in.Ctx.Step, bugFieldMax))
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "- Server time: %s\n", in.Now.Format(time.RFC3339))
	return b.String()
}

// upstreamGone reports whether err says the issue no longer exists.
func upstreamGone(err error) bool {
	var sc statusCoder
	if !errors.As(err, &sc) {
		return false
	}
	return sc.StatusCode() == http.StatusNotFound || sc.StatusCode() == http.StatusGone
}

// queueNote is the sentence the reporter reads. Honest about the one
// thing that matters: nothing was filed.
func queueNote(reason string) string {
	switch reason {
	case bugReasonPaused:
		return "Filing is paused right now, so your report was saved for the admin to review. It has not been filed as an issue."
	case bugReasonAccount:
		return "You have sent several reports recently, so this one was saved for the admin to review. It has not been filed as an issue."
	default:
		return "A lot of reports have come in today, so yours was saved for the admin to review. It has not been filed as an issue."
	}
}

// queueUnavailable is what a reporter hears when the report cannot be
// queued either, because there is no store to queue into.
func queueUnavailable(reason string) error {
	if reason == bugReasonPaused {
		return httpError(http.StatusServiceUnavailable, "bug reporting is paused right now; try again later")
	}
	return httpError(http.StatusTooManyRequests, "too many reports right now; try again later")
}

// adminBugReport handles GET /admin/bugreport: the kill switch, the
// day's count against the cap, and the queued reports, newest first.
func (g *bugGate) adminBugReport(c Config, w http.ResponseWriter, _ *http.Request) error {
	now := c.now()
	g.mu.Lock()
	g.rollDayLocked(now)
	body := map[string]any{
		"paused":                 g.st.Paused,
		"filed_today":            g.st.Filed,
		"daily_cap":              g.lim.DailyCap,
		"account_max":            g.lim.AccountMax,
		"account_window_seconds": int(g.lim.AccountWindow.Seconds()),
		"queue_enabled":          g.store.Enabled(),
	}
	if g.st.Paused {
		body["paused_at"] = g.st.PausedAt
	}
	g.mu.Unlock()

	queued := []map[string]any{}
	if g.store.Enabled() {
		ms, err := g.store.Queued(bugQueueListMax)
		if err != nil {
			return httpError(http.StatusInternalServerError, "reading the queue failed")
		}
		for _, m := range ms {
			queued = append(queued, map[string]any{
				"report_id":  m.ID,
				"created_at": m.CreatedAt,
				"reason":     m.Queued.Reason,
				"kind":       m.Queued.Kind,
				"label":      m.Queued.Label,
				"reporter":   m.Queued.Reporter,
				"game_id":    m.GameID,
				"title":      m.Queued.Title,
				"body":       m.Queued.Body,
			})
		}
	}
	body["queued"] = queued
	return writeJSON(w, http.StatusOK, body)
}

// adminSetBugPaused handles PUT /admin/bugreport/paused {"paused": bool}.
func (g *bugGate) adminSetBugPaused(c Config, w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Paused *bool `json:"paused"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	if req.Paused == nil {
		return httpError(http.StatusBadRequest, "paused is required")
	}
	g.setPaused(c.now(), *req.Paused)
	return writeJSON(w, http.StatusOK, map[string]any{"paused": *req.Paused})
}
