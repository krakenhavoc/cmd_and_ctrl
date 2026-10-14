package lobby

// modelkeys_http.go is ADR 0182's /me/model-key: a signed-in person's
// own model API key for the AI seats they add, over internal/modelkeys.
//
// The key goes in on PUT and never comes out again: every answer is a
// modelkeys.View (provider, model, the last four characters, dates and
// the last failure), and nothing here logs a request body.
//
// The feature needs the database and CMDCTRL_IDENTITY_KEY (the sealing
// key, ADR 0051). Missing either, every route answers 503 with
// "available": false and the client hides the Settings tab.

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/modelkeys"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// maxModelKeyRequestBytes caps a PUT body: a key of at most 512 bytes,
// a model id, an address and two numbers.
const maxModelKeyRequestBytes = 4096

// modelKeyResponse is the body of GET, PUT and DELETE. A 503 carries
// Error and "available": false and nothing else that means anything.
type modelKeyResponse struct {
	Error string `json:"error,omitempty"`
	// Available is whether this server offers own keys at all.
	Available bool `json:"available"`
	// Key is the saved key's View, or null when none is saved.
	Key *modelkeys.View `json:"key"`
	// Providers is the picker: the listed providers and "custom".
	Providers []modelkeys.Provider `json:"providers,omitempty"`
	// Limits are the bounds and defaults the form offers.
	Limits *modelKeyLimits `json:"limits,omitempty"`
}

type modelKeyLimits struct {
	MinGamesPerDay      int `json:"min_games_per_day"`
	MaxGamesPerDay      int `json:"max_games_per_day"`
	DefaultGamesPerDay  int `json:"default_games_per_day"`
	MinCallsPerGame     int `json:"min_calls_per_game"`
	MaxCallsPerGame     int `json:"max_calls_per_game"`
	DefaultCallsPerGame int `json:"default_calls_per_game"`
}

func modelKeyResponseOf(v *modelkeys.View) modelKeyResponse {
	return modelKeyResponse{
		Available: true,
		Key:       v,
		Providers: modelkeys.Providers(),
		Limits: &modelKeyLimits{
			MinGamesPerDay: modelkeys.MinGamesPerDay, MaxGamesPerDay: modelkeys.MaxGamesPerDay,
			DefaultGamesPerDay: modelkeys.DefaultGamesPerDay,
			MinCallsPerGame:    modelkeys.MinCallsPerGame, MaxCallsPerGame: modelkeys.MaxCallsPerGame,
			DefaultCallsPerGame: modelkeys.DefaultCallsPerGame,
		},
	}
}

// putModelKeyRequest is PUT /me/model-key. Key may be left out to keep
// the saved key while changing the model or the limits.
type putModelKeyRequest struct {
	Provider     string `json:"provider"`
	BaseURL      string `json:"base_url,omitempty"`
	Model        string `json:"model"`
	Key          string `json:"key,omitempty"`
	GamesPerDay  int    `json:"games_per_day"`
	CallsPerGame int    `json:"calls_per_game"`
}

// writeModelKeysOff is the 503 every route answers while the feature is
// off.
func writeModelKeysOff(w http.ResponseWriter) error {
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusServiceUnavailable, modelKeyResponse{
		Error:     "your own model key is not available on this server",
		Available: false,
	})
}

// modelKeyError maps a Service error onto the answer. A refusal of what
// was typed is the person's to fix (422, its own sentence); anything
// else is logged without detail and answered without it.
func (c Config) modelKeyError(err error, what string) error {
	var inv *modelkeys.InvalidError
	if errors.As(err, &inv) {
		return httpError(http.StatusUnprocessableEntity, inv.Msg)
	}
	c.logger().Error(what+" failed", "err", err)
	return httpError(http.StatusInternalServerError, "could not "+what+"; try again")
}

// myModelKey is GET /me/model-key.
func myModelKey(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if c.ModelKeys == nil {
		return writeModelKeysOff(w)
	}
	v, ok, err := c.ModelKeys.Get(r.Context(), p.UserID)
	if err != nil {
		return c.modelKeyError(err, "load your model key")
	}
	w.Header().Set("Cache-Control", "no-store")
	if !ok {
		return writeJSON(w, http.StatusOK, modelKeyResponseOf(nil))
	}
	return writeJSON(w, http.StatusOK, modelKeyResponseOf(&v))
}

// putMyModelKey is PUT /me/model-key: add, replace, or change the model
// and limits of the saved key. Per-person bucket, spent only by a write
// that decoded.
func putMyModelKey(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		if c.ModelKeys == nil {
			return writeModelKeysOff(w)
		}
		var req putModelKeyRequest
		r.Body = http.MaxBytesReader(w, r.Body, maxModelKeyRequestBytes)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			// Never echo the decoder's error: it can quote the body,
			// and the body holds the key.
			return httpError(http.StatusBadRequest, "invalid body")
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "1")
			return httpError(http.StatusTooManyRequests, "too many changes; try again in a second")
		}
		v, err := c.ModelKeys.Save(r.Context(), p.UserID, modelkeys.Input{
			Provider: req.Provider, BaseURL: req.BaseURL, Model: req.Model, Key: req.Key,
			GamesPerDay: req.GamesPerDay, CallsPerGame: req.CallsPerGame,
		})
		req.Key = ""
		if err != nil {
			return c.modelKeyError(err, "save your model key")
		}
		w.Header().Set("Cache-Control", "no-store")
		return writeJSON(w, http.StatusOK, modelKeyResponseOf(&v))
	}
}

// deleteMyModelKey is DELETE /me/model-key. Removing a key that is not
// there answers the same as removing one that is: either way there is
// none now.
func deleteMyModelKey(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		if c.ModelKeys == nil {
			return writeModelKeysOff(w)
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "1")
			return httpError(http.StatusTooManyRequests, "too many changes; try again in a second")
		}
		if _, err := c.ModelKeys.Remove(r.Context(), p.UserID); err != nil {
			return c.modelKeyError(err, "remove your model key")
		}
		w.Header().Set("Cache-Control", "no-store")
		return writeJSON(w, http.StatusOK, modelKeyResponseOf(nil))
	}
}

// ownKeySeat is an accepted "run on my key" request: whose key, and the
// model the seat asks for ("" for the key's own).
type ownKeySeat struct {
	owner uuid.UUID
	model string
}

// ownKeyForSeat checks a POST /games/{id}/seats/bot that asks for the
// caller's own key. The owner is always the CALLER's user id, taken from
// their session: there is no field to name anyone else's key, which is
// what keeps a key to its owner's own seats (ADR 0182 §5).
func ownKeyForSeat(c Config, r *http.Request, tier, modelID string) (ownKeySeat, error) {
	if c.ModelKeys == nil {
		return ownKeySeat{}, httpError(http.StatusServiceUnavailable, "your own model key is not available on this server")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID == uuid.Nil {
		return ownKeySeat{}, httpError(http.StatusForbidden, "sign in with Discord to run a bot on your own model key")
	}
	if !aiseat.Tier(tier).NeedsModel() || !declaredTier(tier) {
		return ownKeySeat{}, httpError(http.StatusUnprocessableEntity, "only the assisted and strong bots use a model; pick one of them to run it on your key")
	}
	if modelID != "" {
		m, err := modelkeys.ValidateModel(modelID)
		if err != nil {
			return ownKeySeat{}, c.modelKeyError(err, "check the model")
		}
		modelID = m
	}
	if _, has := c.ModelKeys.HasKey(r.Context(), p.UserID); !has {
		return ownKeySeat{}, httpError(http.StatusUnprocessableEntity, "add your model key in Settings › AI opponents first")
	}
	left, err := c.ModelKeys.GamesLeftToday(r.Context(), p.UserID)
	if err != nil {
		return ownKeySeat{}, c.modelKeyError(err, "check your model key's limits")
	}
	if !left {
		return ownKeySeat{}, httpError(http.StatusUnprocessableEntity, "your model key has played all the games you allowed it today (Settings › AI opponents)")
	}
	return ownKeySeat{owner: p.UserID, model: modelID}, nil
}

// declaredTier reports whether tier is one this build declares.
func declaredTier(tier string) bool {
	_, ok := aiseat.LookupTier(tier)
	return ok
}

// ownKeyOption is GET /bot/options' own_key: what the picker's "run on
// my key" choice shows. Present only for a signed-in person with a key
// on a server that offers the feature.
type ownKeyOption struct {
	ProviderLabel string `json:"provider_label"`
	Model         string `json:"model"`
	// GamesLeft is false when today's games-a-day limit is used up; the
	// picker then says so instead of offering the choice.
	GamesLeft bool `json:"games_left"`
}

// ownKeyOptionFor reads the caller's key for the picker.
func ownKeyOptionFor(c Config, r *http.Request) *ownKeyOption {
	if c.ModelKeys == nil {
		return nil
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID == uuid.Nil {
		return nil
	}
	v, has := c.ModelKeys.HasKey(r.Context(), p.UserID)
	if !has {
		return nil
	}
	return &ownKeyOption{ProviderLabel: v.ProviderLabel, Model: v.Model, GamesLeft: v.GamesToday < v.GamesPerDay}
}

// testMyModelKey is POST /me/model-key/test: one small call to the
// provider with the saved key, through the same guard a seat's calls
// take (ADR 0182 §3). The answer is a fixed sentence, never what the
// provider or the network said. Slow bucket: it is a real request.
func testMyModelKey(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		if c.ModelKeys == nil {
			return writeModelKeysOff(w)
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "10")
			return httpError(http.StatusTooManyRequests, "you tested the key a moment ago; try again in a few seconds")
		}
		res, err := c.ModelKeys.Test(r.Context(), p.UserID)
		if errors.Is(err, modelkeys.ErrNoKey) {
			return httpError(http.StatusNotFound, "save a key first")
		}
		if err != nil {
			return c.modelKeyError(err, "test your model key")
		}
		w.Header().Set("Cache-Control", "no-store")
		return writeJSON(w, http.StatusOK, res)
	}
}
