package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The draft's HTTP routes and its socket (ADR 0222 §5, #5438;
// docs/lobby.md, docs/protocol.md).
//
// # These routes are public
//
// None sits behind auth.Middleware: a visitor can draft against bots
// with no sign-in (ADR 0222 §9), as they can play a practice game. Who
// the caller is comes from the session they happen to carry.
//
// # The seat token
//
// A draft seat is a session of role auth.RoleDrafter, minted by the
// server's one authenticator. It is returned in the body and is NOT
// set as the session cookie: a signed-in person stays signed in, and
// drafting never replaces the session a table seat lives in. The
// client sends it as a Bearer header, and as ?token= on the socket,
// which a browser cannot give a header. auth.Middleware refuses the
// role on every other route.
//
// They all live under /draft, beside GET /draft/sets, because that is
// the path the production proxy already sends to this server.

// Stable codes of the draft routes' error bodies.
const (
	CodeSetNotDraftable = "set_not_draftable"
	CodeDraftStarted    = "draft_started"
	CodeDraftFull       = "draft_full"
)

// CubeBooster is the booster label of a cube draft: its packs are cut
// from the cube and approximate no retail product.
const CubeBooster = "cube"

// DraftSockets is the part of *ws.Hub the draft socket needs.
type DraftSockets interface {
	ServeFeed(w http.ResponseWriter, r *http.Request, feed ws.Feed, b ws.FeedBinding)
}

type createDraftRequest struct {
	Name string `json:"name"`
	// Set is a set code GET /draft/sets lists as draftable. Cube is one
	// of the caller's own cubes (GET /me/cubes), by id, and needs a
	// sign-in. Exactly one of the two.
	Set  string `json:"set"`
	Cube string `json:"cube,omitempty"`
	// PackSize is the cards in a cube's packs, 9 to 20; zero is 15. A
	// set's packs are its own size and take none.
	PackSize int `json:"pack_size,omitempty"`
	// Mode is "pick_one" (the default) or "pick_two".
	Mode string `json:"mode"`
	// Seats is 2 to 8; zero is 8.
	Seats int `json:"seats"`
	// PickSeconds is the pick timer, 15 to 300; zero is 60.
	PickSeconds int `json:"pick_seconds"`
	// BuildSeconds is the deckbuilding timer, 120 to 1800; zero is 900.
	BuildSeconds int `json:"build_seconds"`
	// BuildUntimed builds with no timer (#5700): the rounds begin when
	// the last person has submitted. It excludes build_seconds.
	BuildUntimed bool `json:"build_untimed,omitempty"`
	// BestOf is the length of the draft's matches: 1 (the default) or
	// 3.
	BestOf int `json:"best_of"`
}

type joinDraftRequest struct {
	Invite string `json:"invite"`
	Name   string `json:"name"`
}

// draftSessionResponse answers a create and a join: the seat's token
// and the draft as that seat sees it.
type draftSessionResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	Draft     draft.View `json:"draft"`
}

type draftResponse struct {
	Draft draft.View `json:"draft"`
}

// registerDraftRoutes mounts the draft's routes on the lobby's mux.
// join is the bucket of the other routes that take an invite; createIP
// is the operator's per-address creation bucket (ADR 0148).
func registerDraftRoutes(mux interface {
	Handle(pattern string, handler http.Handler)
}, c Config, newLimiter func(rate, burst float64) *ratelimit.Limiter, join, createIP *ratelimit.Limiter) {
	// A visitor's own bucket, as the guest practice table has: 3 per 10
	// minutes per address. A signed-in person is held to one unfinished
	// draft instead, and to the per-address bucket with everyone else.
	visitor := newLimiter(3.0/600, 3)
	mux.Handle("POST /draft", createIPGate(createIP, handlerFunc(c, createDraft(visitor))))
	mux.Handle("GET /draft/{id}", join.Middleware(handlerFunc(c, previewDraft)))
	mux.Handle("POST /draft/{id}/join", join.Middleware(handlerFunc(c, joinDraft)))
	mux.Handle("POST /draft/{id}/start", handlerFunc(c, startDraft))
	mux.Handle("POST /draft/{id}/leave", handlerFunc(c, leaveDraft))
	mux.Handle("GET /draft/{id}/ws", handlerFunc(c, draftSocket))
	// The seat at the table of the caller's match (#5440).
	mux.Handle("POST /draft/{id}/match", handlerFunc(c, seatDraftMatch))
}

// draftCredential reads a draft seat's token: the Bearer header, then
// ?token=. Not the cookie: the cookie is the caller's sign-in or table
// seat, and a draft token is never put there.
func draftCredential(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if v, ok := strings.CutPrefix(h, "Bearer "); ok && v != "" {
			return v
		}
	}
	return r.URL.Query().Get("token")
}

// draftSeat validates the request's seat token for the draft in the
// path: the session must be a draft seat, of this draft, that a person
// still holds.
func draftSeat(c Config, r *http.Request) (auth.Principal, *draft.Room, error) {
	id, err := draftIDFromPath(r)
	if err != nil {
		return auth.Principal{}, nil, err
	}
	cred := draftCredential(r)
	if cred == "" {
		return auth.Principal{}, nil, httpError(http.StatusUnauthorized, "authentication required")
	}
	p, err := c.Auth.Validate(r.Context(), cred)
	if err != nil {
		return auth.Principal{}, nil, httpError(http.StatusUnauthorized, "session expired or invalid")
	}
	if p.Role != auth.RoleDrafter || p.PlayerID == uuid.Nil {
		return auth.Principal{}, nil, httpError(http.StatusForbidden, "this session is not a draft seat")
	}
	if p.DraftID != id {
		return auth.Principal{}, nil, httpError(http.StatusForbidden, "session is not for this draft")
	}
	room := c.Lobby.DraftRoom(id)
	if room == nil {
		return auth.Principal{}, nil, ErrDraftNotFound
	}
	if !room.HasSeat(p.PlayerID) {
		// The seat left the draft; its token is spent.
		return auth.Principal{}, nil, httpError(http.StatusForbidden, "this seat has left the draft")
	}
	return p, room, nil
}

func draftIDFromPath(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// Not a draft id: the same answer as a draft that is not there.
		return uuid.Nil, ErrDraftNotFound
	}
	return id, nil
}

// draftCaller is who is asking on a route that needs no session: a
// signed-in person, or a visitor.
func draftCaller(c Config, r *http.Request) (identity DiscordIdentity, source *auth.Principal) {
	return signedInIdentity(c, r)
}

// issueDraftSeat mints the seat's session and builds the answer.
func issueDraftSeat(c Config, r *http.Request, draftID, seatID uuid.UUID, name string, identity DiscordIdentity, source *auth.Principal) (draftSessionResponse, error) {
	p := auth.Principal{Role: auth.RoleDrafter, DraftID: draftID, PlayerID: seatID, Name: name}
	if source != nil {
		p.UserID = source.UserID
	}
	p = withIdentity(p, identity)
	tok, issued, err := issueFor(r.Context(), c, p, source)
	if err != nil {
		return draftSessionResponse{}, err
	}
	room := c.Lobby.DraftRoom(draftID)
	if room == nil {
		return draftSessionResponse{}, ErrDraftNotFound
	}
	view, ok := room.View(seatID)
	if !ok {
		return draftSessionResponse{}, ErrDraftNotFound
	}
	return draftSessionResponse{Token: tok, ExpiresAt: issued.ExpiresAt, Draft: view}, nil
}

func writeCoded(w http.ResponseWriter, status int, code, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(map[string]string{"error": msg, "code": code})
}

// writeDraftError answers a draft engine refusal that has its own
// status; ok is false for an error the lobby's own mapper should take.
func writeDraftError(w http.ResponseWriter, err error) (handled bool, werr error) {
	switch {
	case errors.Is(err, draft.ErrStarted):
		return true, writeCoded(w, http.StatusConflict, CodeDraftStarted, "the draft has already started")
	case errors.Is(err, draft.ErrFull):
		return true, writeCoded(w, http.StatusConflict, CodeDraftFull, "every seat in this draft is taken")
	case errors.Is(err, draft.ErrNotHost):
		return true, httpError(http.StatusForbidden, "only the host can start the draft")
	case errors.Is(err, draft.ErrNoSuchSeat):
		return true, httpError(http.StatusForbidden, "this seat has left the draft")
	case errors.Is(err, draft.ErrBadConfig):
		return true, httpError(http.StatusBadRequest, strings.TrimPrefix(err.Error(), "draft: "))
	}
	return false, nil
}

// createDraft handles POST /draft.
func createDraft(visitor *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		var body createDraftRequest
		if err := decodeJSON(w, r, &body); err != nil {
			return err
		}
		mode := draft.Mode(strings.ToLower(strings.TrimSpace(body.Mode)))
		if mode == "" {
			mode = draft.PickOne
		}
		if !mode.Valid() {
			return httpError(http.StatusBadRequest, fmt.Sprintf("unknown mode %q (want %q or %q)", body.Mode, draft.PickOne, draft.PickTwo))
		}
		seats := body.Seats
		if seats == 0 {
			seats = draft.DefaultSeats
		}
		if seats < draft.MinSeats || seats > draft.MaxSeats {
			return httpError(http.StatusBadRequest, fmt.Sprintf("seats is %d (want %d to %d)", body.Seats, draft.MinSeats, draft.MaxSeats))
		}
		timeout := draft.DefaultPickTimeout
		if body.PickSeconds != 0 {
			timeout = time.Duration(body.PickSeconds) * time.Second
			if timeout < draft.MinPickTimeout || timeout > draft.MaxPickTimeout {
				return httpError(http.StatusBadRequest, fmt.Sprintf("pick_seconds is %d (want %d to %d)", body.PickSeconds,
					int(draft.MinPickTimeout/time.Second), int(draft.MaxPickTimeout/time.Second)))
			}
		}

		build := time.Duration(0)
		if body.BuildUntimed && body.BuildSeconds != 0 {
			return httpError(http.StatusBadRequest, "send build_seconds or build_untimed, not both")
		}
		if body.BuildSeconds != 0 {
			build = time.Duration(body.BuildSeconds) * time.Second
			if build < draft.MinBuildTimeout || build > draft.MaxBuildTimeout {
				return httpError(http.StatusBadRequest, fmt.Sprintf("build_seconds is %d (want %d to %d)", body.BuildSeconds,
					int(draft.MinBuildTimeout/time.Second), int(draft.MaxBuildTimeout/time.Second)))
			}
		}
		if body.BestOf != 0 && body.BestOf != 1 && body.BestOf != 3 {
			return httpError(http.StatusBadRequest, fmt.Sprintf("best_of is %d (want 1 or 3)", body.BestOf))
		}

		identity, source := draftCaller(c, r)
		signedIn := source != nil && identity.Populated()
		spec := DraftSpec{
			Name:         body.Name,
			Mode:         mode,
			Seats:        seats,
			PickTimeout:  timeout,
			BuildTimeout: build,
			BuildUntimed: body.BuildUntimed,
			BestOf:       body.BestOf,
			HostName:     "Player",
			Solo:         true,
		}
		set := strings.ToLower(strings.TrimSpace(body.Set))
		cubeID := strings.TrimSpace(body.Cube)
		switch {
		case set != "" && cubeID != "":
			return httpError(http.StatusBadRequest, "send set or cube, not both")
		case cubeID != "":
			// A cube (ADR 0222 §8): the caller's own saved list, so it
			// needs their account. A visitor drafts a set.
			if !signedIn {
				return writeCoded(w, http.StatusForbidden, CodeCubeNeedsSignIn,
					"Sign in with Discord to draft a cube. You can draft a set without signing in.")
			}
			cube, handled, err := cubeForDraft(c, w, r, source.UserID, cubeID, seats, body.PackSize)
			if handled || err != nil {
				return err
			}
			spec.SetName, spec.Booster, spec.Cube, spec.Source = cube.Name, CubeBooster, true, cube.Source
		case set == "":
			return httpError(http.StatusBadRequest, "set is required: a set code from GET /draft/sets, or cube: one of your cubes")
		default:
			if body.PackSize != 0 {
				return httpError(http.StatusBadRequest, "pack_size is for a cube: a set's packs have their own size")
			}
			// The set: only one the draftable check passes (ADR 0222
			// §4), by the same list GET /draft/sets serves. A draft
			// never hands someone a card that does not work.
			pool, status, ok := c.DraftSets.DraftablePool(set)
			if !ok {
				return writeCoded(w, http.StatusUnprocessableEntity, CodeSetNotDraftable,
					fmt.Sprintf("%q cannot be drafted yet: every card its packs can contain has to be fully built first", body.Set))
			}
			spec.Set, spec.SetName, spec.Booster = status.Code, status.Name, string(status.Booster)
			spec.Source = draft.NewPoolSource(c.Cards, pool)
		}
		if signedIn {
			// A signed-in person: an invite link, and others may join.
			spec.Solo = false
			spec.HostName = identity.DisplayName()
			spec.HostUser = source.UserID
		} else {
			source = nil
			identity = DiscordIdentity{}
			if !visitor.AllowRequest(r) {
				w.Header().Set("Retry-After", "60")
				return ErrCreateIPRateLimited
			}
		}
		// The connection cap, before a draft is made that its socket
		// would be refused at (ADR 0148), as on a join.
		if c.Lobby.ConnectionsFull() {
			return capacity.ErrServerFull
		}
		draftID, seatID, err := c.Lobby.CreateDraft(spec)
		if err != nil {
			if handled, werr := writeDraftError(w, err); handled {
				return werr
			}
			return err
		}
		out, err := issueDraftSeat(c, r, draftID, seatID, spec.HostName, identity, source)
		if err != nil {
			c.Lobby.closeDraft(draftID)
			return err
		}
		if c.Log != nil {
			c.Log.Info("draft created", "draft_id", draftID.String(), "set", spec.Set, "cube", spec.Cube, "mode", string(mode), "seats", seats, "solo", spec.Solo)
		}
		return writeJSON(w, http.StatusCreated, out)
	}
}

// previewDraft handles GET /draft/{id}?invite=.
func previewDraft(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := draftIDFromPath(r)
	if err != nil {
		return err
	}
	var seat uuid.UUID
	if cred := draftCredential(r); cred != "" {
		if p, verr := c.Auth.Validate(r.Context(), cred); verr == nil && p.Role == auth.RoleDrafter && p.DraftID == id {
			seat = p.PlayerID
		}
	}
	preview, err := c.Lobby.PreviewDraft(id, r.URL.Query().Get("invite"), seat)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, preview)
}

// joinDraft handles POST /draft/{id}/join.
func joinDraft(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := draftIDFromPath(r)
	if err != nil {
		return err
	}
	var body joinDraftRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if c.Lobby.ConnectionsFull() {
		return capacity.ErrServerFull
	}
	identity, source := draftCaller(c, r)
	name := body.Name
	var userID uuid.UUID
	if source != nil && identity.Populated() {
		name = identity.DisplayName()
		userID = source.UserID
	} else {
		source, identity = nil, DiscordIdentity{}
	}
	seatID, err := c.Lobby.JoinDraft(id, body.Invite, name, userID)
	if err != nil {
		if handled, werr := writeDraftError(w, err); handled {
			return werr
		}
		return err
	}
	if strings.TrimSpace(name) == "" {
		name = "Player"
	}
	out, err := issueDraftSeat(c, r, id, seatID, name, identity, source)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, out)
}

// startDraft handles POST /draft/{id}/start.
func startDraft(c Config, w http.ResponseWriter, r *http.Request) error {
	p, room, err := draftSeat(c, r)
	if err != nil {
		return err
	}
	if err := room.Start(p.PlayerID); err != nil {
		if handled, werr := writeDraftError(w, err); handled {
			return werr
		}
		return err
	}
	view, ok := room.View(p.PlayerID)
	if !ok {
		return ErrDraftNotFound
	}
	return writeJSON(w, http.StatusOK, draftResponse{Draft: view})
}

// leaveDraft handles POST /draft/{id}/leave.
func leaveDraft(c Config, w http.ResponseWriter, r *http.Request) error {
	p, _, err := draftSeat(c, r)
	if err != nil {
		return err
	}
	if err := c.Lobby.LeaveDraft(p.DraftID, p.PlayerID); err != nil {
		if handled, werr := writeDraftError(w, err); handled {
			return werr
		}
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// draftSocket handles GET /draft/{id}/ws: it authorizes the seat and
// hands the socket to the hub as a feed.
func draftSocket(c Config, w http.ResponseWriter, r *http.Request) error {
	p, room, err := draftSeat(c, r)
	if err != nil {
		return err
	}
	if c.DraftHub == nil {
		return httpError(http.StatusServiceUnavailable, "draft sockets are not available on this server")
	}
	c.DraftHub.ServeFeed(w, r, &draftFeed{room: room, seat: p.PlayerID}, ws.FeedBinding{
		UserID:   p.UserID,
		IssuedAt: p.IssuedAt,
		Seated:   true,
		Label:    "draft " + p.DraftID.String(),
	})
	return nil
}

// draftFeed is one seat's end of a draft, as a ws.Feed.
type draftFeed struct {
	room *draft.Room
	seat uuid.UUID
}

// draftSub adapts the socket to draft.Subscriber.
type draftSub struct{ out ws.FeedSink }

func (s draftSub) Send(v draft.View)   { s.out.SendFrame(protocol.KindDraft, v) }
func (s draftSub) Close(reason string) { s.out.Close(reason) }

func (f *draftFeed) Attach(out ws.FeedSink) (func(), error) {
	return f.room.Attach(f.seat, draftSub{out})
}

// Handle takes the frames a draft seat sends: its pick, and the deck it
// built.
func (f *draftFeed) Handle(frame protocol.Frame) *protocol.ErrorPayload {
	if frame.Kind == protocol.KindDraftDeck {
		return f.handleDeck(frame)
	}
	if frame.Kind != protocol.KindDraftPick {
		return &protocol.ErrorPayload{Code: protocol.CodeBadRequest, Message: "unknown or unsupported kind"}
	}
	var body protocol.DraftPickPayload
	if err := json.Unmarshal(frame.Payload, &body); err != nil {
		return &protocol.ErrorPayload{Code: protocol.CodeBadRequest, Message: "draft_pick payload is not valid"}
	}
	if len(body.Cards) == 0 || len(body.Cards) > draft.PickTwo.PerPick() {
		return &protocol.ErrorPayload{Code: protocol.CodeDraftPickRefused, Message: "name the card or cards you are taking"}
	}
	if err := f.room.Pick(f.seat, body.Cards); err != nil {
		return &protocol.ErrorPayload{Code: protocol.CodeDraftPickRefused, Message: draftPickMessage(err)}
	}
	return nil
}

// draftPickMessage is the sentence a player sees for a refused pick.
func draftPickMessage(err error) string {
	switch {
	case errors.Is(err, draft.ErrAlreadyPick):
		return "You have already picked from this pack."
	case errors.Is(err, draft.ErrNotPicking):
		return "No pick is being made right now."
	case errors.Is(err, draft.ErrBadPick):
		return "That pick does not match the pack you are holding."
	}
	return "That pick could not be made."
}

// maxDraftDeckCards bounds a draft_deck frame: no pool is this large.
const maxDraftDeckCards = 200

// handleDeck takes a seat's deck for the rounds.
func (f *draftFeed) handleDeck(frame protocol.Frame) *protocol.ErrorPayload {
	var body protocol.DraftDeckPayload
	if err := json.Unmarshal(frame.Payload, &body); err != nil {
		return &protocol.ErrorPayload{Code: protocol.CodeBadRequest, Message: "draft_deck payload is not valid"}
	}
	if len(body.Main) > maxDraftDeckCards || len(body.Basics) > len(draft.BasicLandNames) {
		return &protocol.ErrorPayload{Code: protocol.CodeDraftDeckRefused, Message: "That deck could not be read."}
	}
	if err := f.room.SubmitDeck(f.seat, draft.LimitedDeck{Main: body.Main, Basics: body.Basics}); err != nil {
		return &protocol.ErrorPayload{Code: protocol.CodeDraftDeckRefused, Message: draftDeckMessage(err)}
	}
	return nil
}

// draftDeckMessage is the sentence a player sees for a refused deck.
func draftDeckMessage(err error) string {
	switch {
	case errors.Is(err, draft.ErrDeckIllegal):
		msg := strings.TrimPrefix(err.Error(), draft.ErrDeckIllegal.Error()+": ")
		if msg == "" {
			return "That deck is not legal."
		}
		return strings.ToUpper(msg[:1]) + msg[1:] + "."
	case errors.Is(err, draft.ErrNotBuilding):
		return "Decks are not being built right now."
	}
	return "That deck could not be submitted."
}

// seatDraftMatch handles POST /draft/{id}/match: it seats the caller at
// the table of their match in the round in play. The table seat's
// session is minted from the draft seat's and set as the session
// cookie, which is what the game page and its socket read, exactly as
// a join's is. A signed-in person stays signed in: the seat session
// carries their account, as every table seat's does.
func seatDraftMatch(c Config, w http.ResponseWriter, r *http.Request) error {
	p, _, err := draftSeat(c, r)
	if err != nil {
		return err
	}
	meta, playerID, err := c.Lobby.DraftMatchSeat(p.DraftID, p.PlayerID)
	if err != nil {
		if errors.Is(err, ErrDraftNoMatch) {
			return httpError(http.StatusConflict, "you have no match to play right now")
		}
		return err
	}
	name := p.Name
	for _, s := range meta.Players {
		if s.PlayerID == playerID {
			name = s.Name
		}
	}
	seat := withIdentity(auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   p.UserID,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     name,
	}, principalIdentity(p))
	tok, issued, err := issueFor(r.Context(), c, seat, &p)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  playerID,
	})
}
