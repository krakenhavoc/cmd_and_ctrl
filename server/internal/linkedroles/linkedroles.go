// Package linkedroles is ADR 0163: CMD & CTRL stats on a person's
// Discord profile through Discord's Linked Roles.
//
// A person links from Discord (a server role's Links requirement, or
// User Settings → Connections), which sends them to GET
// /auth/discord/linked-role (lobby). The callback hands this package the
// grant: Link stores its refresh token (sealed, in
// identities.role_connection_token) and pushes the person's numbers.
// After that the numbers are pushed again whenever a game they sat in
// ends (Enqueue, from the lobby), from one worker goroutine, so the end of
// a game never waits on Discord. Account deletion asks Discord to empty
// the connection (Forget).
//
// The numbers are recomputed from the database on every push, never
// incremented, so a push that failed is corrected by the next one. There
// is no periodic sweep: the counts change only when a game ends, and
// `joined` is a date Discord turns into "days since" on its own.
//
// Discord's rate limits: each push is two calls (a token refresh and the
// PUT). Pushes go one at a time with a short gap between them, a person
// already waiting is not queued twice, a 429 waits Discord's retry_after
// (capped) once, and a 5xx or a timeout is logged at WARN and left for
// the next game.
//
// Nothing here grants anything in CMD & CTRL (ADR 0163 §1): Linked Roles
// shows numbers, and Discord's server settings decide what they earn.
package linkedroles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// The numbers. Constants, not settings.
const (
	// callTimeout bounds one push from the worker or a deletion: the
	// refresh, the PUT and at most one rate-limit wait.
	callTimeout = 45 * time.Second
	// maxRetryWait caps the one wait a 429 earns in the worker.
	maxRetryWait = 30 * time.Second
	// linkRetryWait caps it while a person waits on the link redirect.
	linkRetryWait = 2 * time.Second
	// pushGap is the pause between two people's pushes, so a table of
	// four ending at once is eight calls spread over a second, not a
	// burst.
	pushGap = 250 * time.Millisecond
	// maxPending is how many people can wait for a push. Beyond it a
	// push is dropped with a warning; the person's next game corrects it.
	maxPending = 1024
)

var (
	// ErrScopeMissing: the grant came back without role_connections.write
	// (the person unticked it on Discord's screen).
	ErrScopeMissing = errors.New("linkedroles: the grant does not include " + discord.RoleConnectionScope)
	// ErrNoRefreshToken: Discord sent no refresh token, so the link could
	// not be kept for later pushes.
	ErrNoRefreshToken = errors.New("linkedroles: the grant came without a refresh token")
	// ErrNotFound: no such account, or it was deleted.
	ErrNotFound = errors.New("linkedroles: account not found")
)

// Discord is the Discord half: the real one is NewDiscord, and tests use
// a fake over an httptest.Server or in memory.
type Discord interface {
	// Register sends the application's metadata fields.
	Register(ctx context.Context) error
	// Refresh trades a role-connection refresh token for a new access
	// token and, usually, a rotated refresh token. A token Discord
	// refuses wraps discord.ErrRefreshRejected.
	Refresh(ctx context.Context, refreshToken string) (discord.TokenResponse, error)
	// Put writes a person's connection with their access token. A 429 is
	// a *discord.RateLimitError.
	Put(ctx context.Context, accessToken string, rc discord.RoleConnection) error
	// Clear empties a person's connection.
	Clear(ctx context.Context, accessToken string) error
}

// NewDiscord is the real Discord half: the sign-in config's client
// credentials for the refresh, and the Linked Roles config for the rest.
func NewDiscord(oauth discord.Config, lr discord.LinkedRoles, client *http.Client) Discord {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return realDiscord{oauth: oauth, lr: lr, client: client}
}

type realDiscord struct {
	oauth  discord.Config
	lr     discord.LinkedRoles
	client *http.Client
}

func (d realDiscord) Register(ctx context.Context) error {
	return d.lr.RegisterMetadata(ctx, d.client)
}

func (d realDiscord) Refresh(ctx context.Context, rt string) (discord.TokenResponse, error) {
	return d.oauth.ExchangeRefreshToken(ctx, d.client, rt)
}

func (d realDiscord) Put(ctx context.Context, at string, rc discord.RoleConnection) error {
	return d.lr.PutRoleConnection(ctx, d.client, at, rc)
}

func (d realDiscord) Clear(ctx context.Context, at string) error {
	return d.lr.ClearRoleConnection(ctx, d.client, at)
}

// Tokens is what the service needs from the user store. *users.SQLStore
// implements it.
type Tokens interface {
	DiscordSubject(ctx context.Context, id uuid.UUID) (string, error)
	RoleConnectionToken(ctx context.Context, subject string) (token string, ok bool, err error)
	SetRoleConnectionToken(ctx context.Context, subject, token string) error
	ClearRoleConnectionToken(ctx context.Context, subject string) error
	OpenRoleConnectionToken(subject string, sealed []byte) (string, error)
}

var _ Tokens = (*users.SQLStore)(nil)

// Service links people and pushes their numbers. Construct it with New,
// start the worker with Run.
type Service struct {
	api    Discord
	tokens Tokens
	stats  StatsReader
	log    *slog.Logger
	// sleep waits between attempts and between people; a test replaces
	// it. It returns early, with the context's error, when ctx ends.
	sleep func(ctx context.Context, d time.Duration) error

	// tokMu serialises every read-refresh-store of a role-connection
	// token. Discord rotates the refresh token on each use, so two
	// refreshes of the same token at once would have one rejected, and
	// that one would unlink the person.
	tokMu sync.Mutex

	mu      sync.Mutex
	pending []uuid.UUID
	queued  map[uuid.UUID]bool
	wake    chan struct{}
	// work counts queued pushes and running deletions, for Wait.
	work sync.WaitGroup
}

// New builds the service. log may be nil.
func New(api Discord, tokens Tokens, stats StatsReader, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		api: api, tokens: tokens, stats: stats, log: log, sleep: sleepCtx,
		queued: map[uuid.UUID]bool{}, wake: make(chan struct{}, 1),
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Register sends the metadata fields (ADR 0163 §2). main calls it once at
// boot, off the boot path; a failure is logged and the next boot tries
// again.
func (s *Service) Register(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return s.api.Register(ctx)
}

// Link is the callback's half (ADR 0163 §3): store the grant's refresh
// token for the account and push the numbers with its access token. On
// any error nothing is kept, and the person is told to try again.
func (s *Service) Link(ctx context.Context, userID uuid.UUID, subject string, tok discord.TokenResponse) error {
	if !tok.HasScope(discord.RoleConnectionScope) {
		return ErrScopeMissing
	}
	if tok.RefreshToken == "" {
		return ErrNoRefreshToken
	}
	st, err := s.stats.Stats(ctx, userID)
	if err != nil {
		return err
	}
	s.tokMu.Lock()
	err = s.tokens.SetRoleConnectionToken(ctx, subject, tok.RefreshToken)
	s.tokMu.Unlock()
	if err != nil {
		return fmt.Errorf("linkedroles: store the grant: %w", err)
	}
	if err := s.put(ctx, tok.AccessToken, st.connection(), linkRetryWait); err != nil {
		s.tokMu.Lock()
		if cerr := s.tokens.ClearRoleConnectionToken(ctx, subject); cerr != nil {
			s.log.Error("linked roles: forgetting a grant whose first push failed", "user_id", userID, "err", cerr)
		}
		s.tokMu.Unlock()
		return fmt.Errorf("linkedroles: first push: %w", err)
	}
	s.log.Info("linked roles: account linked", "user_id", userID)
	return nil
}

// Enqueue asks for a push for each listed account, which the lobby does
// when a game they sat in ends. It never blocks and never calls Discord:
// an account that is not linked is skipped by the worker, one already
// waiting is not queued twice, and a full queue drops the push.
func (s *Service) Enqueue(ids ...uuid.UUID) {
	s.mu.Lock()
	added := false
	for _, id := range ids {
		if id == uuid.Nil || s.queued[id] {
			continue
		}
		if len(s.pending) >= maxPending {
			s.log.Warn("linked roles: push queue full; dropped, the next game corrects it", "user_id", id)
			continue
		}
		s.queued[id] = true
		s.pending = append(s.pending, id)
		s.work.Add(1)
		added = true
	}
	s.mu.Unlock()
	if added {
		select {
		case s.wake <- struct{}{}:
		default:
		}
	}
}

// Run is the worker: it pushes queued accounts one at a time until ctx
// ends.
func (s *Service) Run(ctx context.Context) {
	for {
		id, ok := s.next()
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-s.wake:
				continue
			}
		}
		pctx, cancel := context.WithTimeout(ctx, callTimeout)
		s.Push(pctx, id)
		cancel()
		s.work.Done()
		if err := s.sleep(ctx, pushGap); err != nil {
			s.drain()
			return
		}
	}
}

func (s *Service) next() (uuid.UUID, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.pending) == 0 {
		return uuid.Nil, false
	}
	id := s.pending[0]
	s.pending = s.pending[1:]
	delete(s.queued, id)
	return id, true
}

// drain releases what is still queued when the worker stops, so Wait
// returns at shutdown.
func (s *Service) drain() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for range s.pending {
		s.work.Done()
	}
	s.pending = nil
	s.queued = map[uuid.UUID]bool{}
}

// Wait blocks until every push queued so far and every deletion started
// so far has finished. For tests and shutdown; the worker must be
// running.
func (s *Service) Wait() { s.work.Wait() }

// Push recomputes and pushes one account's numbers now. Not linked, or
// no account, is nothing to do. Every failure is logged here, at WARN
// for an outage and at INFO when Discord ended the link.
func (s *Service) Push(ctx context.Context, userID uuid.UUID) {
	log := s.log.With("user_id", userID)
	subject, err := s.tokens.DiscordSubject(ctx, userID)
	if err != nil {
		return // no Discord identity: nobody to push for
	}
	access, linked := s.refresh(ctx, log, subject)
	if !linked {
		return
	}
	st, err := s.stats.Stats(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return
	}
	if err != nil {
		log.Warn("linked roles: reading the stats failed; left for the next game", "err", err)
		return
	}
	if err := s.put(ctx, access, st.connection(), maxRetryWait); err != nil {
		log.Warn("linked roles: push failed; left for the next game", "err", err)
	}
}

// refresh trades the stored token for an access token and stores the
// rotated one, under tokMu. linked is false when there is nothing to
// push with: not linked, a token that will not open, an outage, or
// Discord refusing the token, which unlinks the person (ADR 0163 §4).
func (s *Service) refresh(ctx context.Context, log *slog.Logger, subject string) (access string, linked bool) {
	s.tokMu.Lock()
	defer s.tokMu.Unlock()
	rt, ok, err := s.tokens.RoleConnectionToken(ctx, subject)
	if err != nil {
		if !errors.Is(err, users.ErrNotFound) {
			log.Warn("linked roles: the stored grant cannot be opened", "var", users.IdentityKeyEnv, "err", err)
		}
		return "", false
	}
	if !ok {
		return "", false
	}
	tr, err := s.api.Refresh(ctx, rt)
	switch {
	case errors.Is(err, discord.ErrRefreshRejected):
		log.Info("linked roles: Discord ended the link (the connection was removed or the app de-authorized); unlinked")
		if cerr := s.tokens.ClearRoleConnectionToken(ctx, subject); cerr != nil {
			log.Error("linked roles: clearing a rejected grant failed", "err", cerr)
		}
		return "", false
	case err != nil:
		log.Warn("linked roles: Discord did not answer the refresh; left for the next game", "err", err)
		return "", false
	}
	if err := s.tokens.SetRoleConnectionToken(ctx, subject, tr.RefreshToken); err != nil {
		log.Error("linked roles: storing the rotated grant failed; the next push may find it rejected", "err", err)
	}
	return tr.AccessToken, true
}

// put sends one connection, waiting once on a 429 (at most maxWait).
func (s *Service) put(ctx context.Context, access string, rc discord.RoleConnection, maxWait time.Duration) error {
	return s.retryOnce(ctx, maxWait, func() error { return s.api.Put(ctx, access, rc) })
}

func (s *Service) retryOnce(ctx context.Context, maxWait time.Duration, call func() error) error {
	err := call()
	var rl *discord.RateLimitError
	if !errors.As(err, &rl) {
		return err
	}
	if werr := s.sleep(ctx, min(rl.RetryAfter, maxWait)); werr != nil {
		return err
	}
	return call()
}

// Forget is account deletion's step (ADR 0163 §7): empty the person's
// connection on Discord so their profile stops showing the numbers. The
// identity row, and the sealed token with it, is already gone; sealed is
// the token the deletion read inside its transaction. It runs in the
// background and is best effort: a failure is logged, and the deletion
// has already happened.
func (s *Service) Forget(subject string, sealed []byte) {
	if subject == "" || len(sealed) == 0 {
		return
	}
	s.work.Add(1)
	go func() {
		defer s.work.Done()
		ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
		defer cancel()
		rt, err := s.tokens.OpenRoleConnectionToken(subject, sealed)
		if err != nil {
			s.log.Warn("linked roles: a deleted account's grant cannot be opened; its Discord profile may still show the numbers until they unlink", "err", err)
			return
		}
		tr, err := s.api.Refresh(ctx, rt)
		if errors.Is(err, discord.ErrRefreshRejected) {
			return // already unlinked on Discord's side
		}
		if err != nil {
			s.log.Warn("linked roles: Discord did not answer for a deleted account; its profile may still show the numbers until they unlink", "err", err)
			return
		}
		if err := s.retryOnce(ctx, maxRetryWait, func() error { return s.api.Clear(ctx, tr.AccessToken) }); err != nil {
			s.log.Warn("linked roles: emptying a deleted account's connection failed", "err", err)
		}
	}()
}
