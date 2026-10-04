package mcpseat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// api is the lobby's HTTP surface, as far as a guest seat uses it. Every
// call carries the session as `Authorization: Bearer`, never as a query
// parameter (§8), and there is no cookie jar: the seat can never ride a
// signed-in browser session, so it always joins as a guest (decision 2).
type api struct {
	http *http.Client
	// joinBackoff is the 429 retry ladder for the join routes (§8).
	joinBackoff []time.Duration
	sleep       func(context.Context, time.Duration) error
}

func newAPI(c *http.Client) *api {
	if c == nil {
		c = &http.Client{Timeout: 20 * time.Second}
	}
	// A redirect could carry the session somewhere the allowlist never
	// saw. None of these routes redirects, so refuse them all.
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.Jar = nil
	return &api{
		http:        c,
		joinBackoff: []time.Duration{time.Second, 2 * time.Second, 4 * time.Second},
		sleep:       sleepCtx,
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// apiError is a non-2xx answer, with the server's own words.
type apiError struct {
	Status     int
	Message    string
	Violations []string
}

func (e *apiError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if len(e.Violations) > 0 {
		msg += ": " + strings.Join(e.Violations, "; ")
	}
	return fmt.Sprintf("%d %s", e.Status, msg)
}

// maxBody bounds what any of these routes may send back.
const maxBody = 4 << 20

func (a *api) do(ctx context.Context, method, url, token string, body, out any) error {
	var rd io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return decodeAPIError(resp.StatusCode, raw)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func decodeAPIError(status int, raw []byte) *apiError {
	e := &apiError{Status: status}
	var body struct {
		Error      string `json:"error"`
		Message    string `json:"message"`
		Violations []struct {
			Message string `json:"message"`
			Card    string `json:"card"`
		} `json:"violations"`
	}
	if json.Unmarshal(raw, &body) == nil {
		e.Message = body.Error
		if e.Message == "" {
			e.Message = body.Message
		}
		for _, v := range body.Violations {
			s := v.Message
			if v.Card != "" && !strings.Contains(s, v.Card) {
				s = v.Card + ": " + s
			}
			if s != "" {
				e.Violations = append(e.Violations, s)
			}
		}
	} else {
		e.Message = strings.TrimSpace(string(raw))
	}
	if len(e.Message) > 500 {
		e.Message = e.Message[:500]
	}
	return e
}

// sessionAnswer is the shape both join routes and reclaim return.
type sessionAnswer struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	PlayerID  uuid.UUID `json:"player_id"`
	Principal struct {
		GameID   uuid.UUID `json:"game_id"`
		PlayerID uuid.UUID `json:"player_id"`
		Name     string    `json:"name"`
	} `json:"principal"`
	Game *gameMeta `json:"game"`
}

// gameMeta is the part of the lobby's GameMeta a seat reads.
type gameMeta struct {
	ID      uuid.UUID  `json:"id"`
	Name    string     `json:"name"`
	State   string     `json:"state"`
	Players []seatInfo `json:"players"`
}

type seatInfo struct {
	PlayerID     uuid.UUID `json:"player_id"`
	Name         string    `json:"name"`
	Seat         int       `json:"seat"`
	DeckName     string    `json:"deck_name"`
	DeckUploaded bool      `json:"deck_uploaded"`
	IsBot        bool      `json:"is_bot"`
}

func (s *sessionAnswer) playerID() uuid.UUID {
	if s.PlayerID != uuid.Nil {
		return s.PlayerID
	}
	return s.Principal.PlayerID
}

// join claims a seat with an invite token, declaring the seat an agent.
// A 429 (the route shares the server's per-IP bucket with /admin/login)
// is retried after 1, 2 and 4 s, then reported.
func (a *api) join(ctx context.Context, inv invite, name, client string) (*sessionAnswer, error) {
	body := joinBody{InviteToken: inv.Token, Name: name, Agent: agentField{Client: client}}
	url := inv.Origin + "/games/" + inv.GameID.String() + "/join"
	var out sessionAnswer
	err := a.withJoinRetry(ctx, func() error { return a.do(ctx, http.MethodPost, url, "", body, &out) })
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// reclaim redeems an admin's seat-reclaim ticket.
func (a *api) reclaim(ctx context.Context, inv invite) (*sessionAnswer, error) {
	url := inv.Origin + "/games/" + inv.GameID.String() + "/reclaim"
	var out sessionAnswer
	err := a.withJoinRetry(ctx, func() error {
		return a.do(ctx, http.MethodPost, url, "", map[string]string{"ticket": inv.Token}, &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (a *api) withJoinRetry(ctx context.Context, call func() error) error {
	for i := 0; ; i++ {
		err := call()
		var ae *apiError
		if err == nil || !errors.As(err, &ae) || ae.Status != http.StatusTooManyRequests || i >= len(a.joinBackoff) {
			return err
		}
		if serr := a.sleep(ctx, a.joinBackoff[i]); serr != nil {
			return serr
		}
	}
}

// mePrincipal is GET /me's answer, as far as the seat reads it.
type mePrincipal struct {
	Role      string    `json:"role"`
	GameID    uuid.UUID `json:"game_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// me asks whether a session is still good, and as what.
func (a *api) me(ctx context.Context, origin, token string) (*mePrincipal, error) {
	var out mePrincipal
	if err := a.do(ctx, http.MethodGet, origin+"/me", token, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// game reads the lobby's record of a table.
func (a *api) game(ctx context.Context, origin, token string, id uuid.UUID) (*gameMeta, error) {
	var out gameMeta
	if err := a.do(ctx, http.MethodGet, origin+"/games/"+id.String(), token, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeckInput is a deck for set_deck and join: a pre-built deck id from the
// server, or a pasted decklist. Exactly one.
type DeckInput struct {
	ID   string `json:"id,omitempty" jsonschema:"a pre-built deck id from this server"`
	List string `json:"list,omitempty" jsonschema:"a decklist, as text (Commander: / Mainboard: sections, or one card per line)"`
}

// deckAnswer is POST /games/{id}/decks's success body.
type deckAnswer struct {
	DeckName      string   `json:"deck_name"`
	CardCount     int      `json:"card_count"`
	Commanders    []string `json:"commanders"`
	Unimplemented []string `json:"unimplemented"`
	Warnings      []struct {
		Message string `json:"message"`
	} `json:"warnings"`
}

// setDeck installs a deck on the seat.
func (a *api) setDeck(ctx context.Context, origin, token string, gameID, playerID uuid.UUID, d DeckInput) (*deckAnswer, error) {
	body := map[string]any{"player_id": playerID}
	if d.ID != "" {
		body["deck"] = d.ID
	} else {
		body["source"] = d.List
	}
	var out deckAnswer
	if err := a.do(ctx, http.MethodPost, origin+"/games/"+gameID.String()+"/decks", token, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// cardMeta is GET /cards/{id}'s Scryfall metadata, as far as `card` shows it.
type cardMeta struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	ManaCost   string `json:"mana_cost"`
	OracleText string `json:"oracle_text"`
	Power      string `json:"power"`
	Toughness  string `json:"toughness"`
	Loyalty    string `json:"loyalty"`
	Defense    string `json:"defense"`
	CardFaces  []struct {
		Name       string `json:"name"`
		TypeLine   string `json:"type_line"`
		ManaCost   string `json:"mana_cost"`
		OracleText string `json:"oracle_text"`
		Power      string `json:"power"`
		Toughness  string `json:"toughness"`
		Loyalty    string `json:"loyalty"`
	} `json:"card_faces"`
}

func (a *api) card(ctx context.Context, origin, token, scryfallID string) (*cardMeta, error) {
	var out cardMeta
	if err := a.do(ctx, http.MethodGet, origin+"/cards/"+scryfallID, token, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
