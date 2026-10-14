package modelkeys

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// SeatClient is the model client for one bot seat that runs on its
// owner's key (aiseat/tiers.OwnKeyClients). It is cheap and touches
// nothing: the Manager builds policies under its own lock, so the key
// is read, opened and counted against the owner's limits on the seat's
// first call, on the seat's own goroutine.
func (s *Service) SeatClient(seat aiseat.SeatSpec) model.Client {
	return &seatClient{svc: s, owner: seat.KeyOwner, game: seat.GameID, model: seat.KeyModel}
}

// seatClient is a model.Client over the owner's key with ADR 0182's
// rules on top:
//
//   - the key is the owner's, read for this seat only, and re-read when
//     the owner replaces or removes it (Service.gen);
//   - the game counts once against the owner's games-a-day limit, and
//     the seat stops at the owner's calls-per-game limit;
//   - a failure that will not pass (refused, out of quota, no such
//     model) ends the seat's model use for the game: every later call
//     answers model.ErrBudgetExceeded at once, which the policy takes
//     as "play the heuristic, quietly", and the seat posts one chat
//     line saying so (aiseat.Noticer);
//   - no error leaving it carries a provider's or a transport's text.
type seatClient struct {
	svc   *Service
	owner uuid.UUID
	game  uuid.UUID
	model string

	mu    sync.Mutex
	inner model.Client
	// secret is the opened key, kept beside the client that already
	// holds it for one reason: to scrub it out of a reply. A custom
	// endpoint can put anything in its answer, and the policy logs a
	// malformed answer and traces every one.
	secret    string
	modelID   string
	gen       uint64
	loaded    bool
	claimed   bool
	calls     int
	limit     int
	dead      string
	notice    string
	lastTouch time.Time
}

// Name identifies the transport in logs. It names no provider and no
// person.
func (c *seatClient) Name() string { return "own-key" }

// TakeNotice implements aiseat.Noticer (through the tiers wrapper).
func (c *seatClient) TakeNotice() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := c.notice
	c.notice = ""
	return n
}

// stop ends the seat's model use for the game. Callers hold c.mu.
// record is whether the reason is a fault with the key, which the
// owner sees on Settings; a limit reached is not one.
func (c *seatClient) stopLocked(reason string, record bool) {
	if c.dead != "" {
		return
	}
	c.dead = reason
	c.inner, c.secret = nil, ""
	c.notice = "I can't use my owner's model key any more, so I'm playing on CMD & CTRL's built-in heuristic for the rest of this game. " + reason
	c.svc.log.Info("own-key bot seat stopped using its model key; playing the heuristic",
		"game", c.game.String(), "reason", reason)
	if record {
		// Synchronous and short: a goroutine would outlive the seat (and
		// a test's database), and this happens once per seat at most.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := c.svc.store.setError(ctx, c.owner, reason, c.svc.now()); err != nil {
			c.svc.log.Warn("recording a model key failure failed", "err", err)
		}
	}
}

// prepare makes sure the seat holds a working client for the owner's
// current key. Callers hold c.mu.
func (c *seatClient) prepareLocked(ctx context.Context) error {
	if c.owner == uuid.Nil {
		c.stopLocked(ReasonRemoved, false)
		return model.ErrBudgetExceeded
	}
	gen := c.svc.generation(c.owner)
	if c.loaded && gen == c.gen && c.inner != nil {
		return nil
	}
	r, key, err := c.svc.openKey(ctx, c.owner)
	switch {
	case errors.Is(err, ErrNoKey):
		c.stopLocked(ReasonRemoved, false)
		return model.ErrBudgetExceeded
	case errors.Is(err, errUnreadable):
		c.stopLocked(ReasonUnreadable, true)
		return model.ErrBudgetExceeded
	case err != nil:
		// The database, not the key: this window falls back, the next
		// one tries again.
		return model.ErrOutage
	}
	inner, err := c.svc.clientFor(r, key)
	if err != nil {
		c.stopLocked(ReasonUnreadable, true)
		return model.ErrBudgetExceeded
	}
	c.inner, c.secret, c.gen, c.loaded, c.limit = inner, key, gen, true, r.CallsPerGame
	c.modelID = c.model
	if c.modelID == "" {
		c.modelID = r.Model
	}
	return nil
}

// Complete makes one call on the owner's key, or refuses at once.
func (c *seatClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	c.mu.Lock()
	if c.dead != "" {
		c.mu.Unlock()
		return model.Response{}, model.ErrBudgetExceeded
	}
	if err := c.prepareLocked(ctx); err != nil {
		c.mu.Unlock()
		return model.Response{}, err
	}
	if !c.claimed {
		ok, err := c.svc.claimGame(ctx, c.owner, c.game)
		if err != nil {
			c.mu.Unlock()
			return model.Response{}, model.ErrOutage
		}
		if !ok {
			c.stopLocked(ReasonDailyLimit, false)
			c.mu.Unlock()
			return model.Response{}, model.ErrBudgetExceeded
		}
		c.claimed = true
	}
	if c.calls >= c.limit {
		c.stopLocked(ReasonCallLimit, false)
		c.mu.Unlock()
		return model.Response{}, model.ErrBudgetExceeded
	}
	c.calls++
	inner, gen, secret := c.inner, c.gen, c.secret
	// The seat's model, whatever the tier's profiles name, and no
	// effort or thinking field: those are tuned for the server's own
	// models, and a model the person chose may refuse either.
	req.Model, req.Effort, req.Thinking = c.modelID, "", ""
	c.mu.Unlock()

	resp, err := inner.Complete(ctx, req)
	if err == nil {
		c.touch()
		if secret != "" {
			resp.Text = strings.ReplaceAll(resp.Text, secret, "[key]")
			resp.Reasoning = strings.ReplaceAll(resp.Reasoning, secret, "[key]")
		}
		return resp, nil
	}
	fatal, reason := classify(err)
	if fatal {
		c.mu.Lock()
		// A key replaced while the call was out is the new key's to
		// prove; the failure belonged to the old one.
		if c.gen == gen && c.svc.generation(c.owner) == gen {
			c.stopLocked(reason, true)
		} else {
			c.loaded = false
		}
		c.mu.Unlock()
	}
	return model.Response{}, sanitize(err, reason)
}

// touch records last_used_at at most once a minute per seat: a seat
// makes a call most windows, and the date is for the owner, not an
// audit.
func (c *seatClient) touch() {
	now := c.svc.now()
	c.mu.Lock()
	if now.Sub(c.lastTouch) < time.Minute {
		c.mu.Unlock()
		return
	}
	c.lastTouch = now
	c.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = c.svc.store.touchUsed(ctx, c.owner, now)
}
