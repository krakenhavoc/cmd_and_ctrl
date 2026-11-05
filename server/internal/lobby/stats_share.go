package lobby

// stats_share.go is ADR 0217 §6 (slice 6 of #4880): a person's share
// link.
//
//   - POST /me/stats/share-link mints an unguessable token (128 bits,
//     base64url) and answers it once. Only its SHA-256 is kept, in
//     stats_prefs.share_link_hash, so the link cannot be shown again:
//     minting again rotates it and the old one stops working.
//   - DELETE /me/stats/share-link revokes it.
//   - GET /stats/shared/{token} is the read-only view behind
//     #/stats/shared/<token>. It needs no sign-in (owner answer 6), so it
//     is rate limited per address like the other unauthenticated
//     routes, and an unknown or revoked token is a 404.
//
// What a shared view carries is the person's display name and their own
// stats: the overview (when their advanced stats are on) and the basic
// record. It never carries another seat, another person's name, a table
// name, a game id or a replay link. The recent-games list of GET
// /me/stats is left out for that reason: it names tables.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// shareTokenBytes is the token's entropy: 128 bits (ADR 0217 §6).
const shareTokenBytes = 16

// shareTokenRe is a token's shape: 16 bytes in unpadded base64url. A
// path that does not match is a 404 without a database read.
var shareTokenRe = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)

// errNoSharedStats is the answer for an unknown, revoked or malformed
// token, and for an account that is gone. One answer for all of them,
// so a probe learns nothing.
var errNoSharedStats = httpError(http.StatusNotFound, "this stats link does not exist or has been revoked")

// shareLinkHash is what stats_prefs keeps for a token.
func shareLinkHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// newShareToken is a fresh token.
func newShareToken() (string, error) {
	b := make([]byte, shareTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// shareLinkResponse is POST /me/stats/share-link's body: the token,
// shown this once, and the client route it opens.
type shareLinkResponse struct {
	Token     string `json:"token"`
	Path      string `json:"path"`
	CreatedAt int64  `json:"created_at"`
}

// postMyStatsShareLink is POST /me/stats/share-link.
func postMyStatsShareLink(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	token, err := newShareToken()
	if err != nil {
		c.logger().Error("minting a stats share link failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not make a link; try again")
	}
	saved, err := c.StatsPrefs.SetShareLink(r.Context(), p.UserID.String(), shareLinkHash(token), c.now())
	if errors.Is(err, gamestats.ErrNoUser) {
		return httpError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		c.logger().Error("saving a stats share link failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not make a link; try again")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusCreated, shareLinkResponse{Token: token, Path: "#/stats/shared/" + token,
		CreatedAt: millisOrZero(saved.ShareLinkAt)})
}

// deleteMyStatsShareLink is DELETE /me/stats/share-link: the link
// answers 404 from now on. Having none is not an error.
func deleteMyStatsShareLink(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	if err := c.StatsPrefs.ClearShareLink(r.Context(), p.UserID.String(), c.now()); err != nil {
		c.logger().Error("revoking a stats share link failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not revoke the link; try again")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// sharedBasic is the basic record a shared view shows (ADR 0165 §7),
// without the recent games, which name tables.
type sharedBasic struct {
	gamestats.Tally
	Unfinished  int                 `json:"unfinished"`
	ByDeck      []gamestats.Group   `json:"by_deck"`
	ByCommander []gamestats.Group   `json:"by_commander"`
	TopCards    []gamestats.TopCard `json:"top_cards"`
}

// sharedStatsResponse is GET /stats/shared/{token}'s body.
type sharedStatsResponse struct {
	DisplayName string `json:"display_name"`
	// Advanced is whether the person's advanced stats are on; Overview
	// is nil while they are off, and the view shows Basic alone.
	Advanced bool                      `json:"advanced"`
	Basic    sharedBasic               `json:"basic"`
	Overview *advancedOverviewResponse `json:"overview"`
}

// sharedStats is GET /stats/shared/{token}.
func sharedStats(c Config, w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	if !shareTokenRe.MatchString(token) || c.StatsPrefs == nil {
		return errNoSharedStats
	}
	makeup, rng, err := advancedFilters(r)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), statsTimeout)
	defer cancel()
	ownerID, err := c.StatsPrefs.ShareLinkOwner(ctx, shareLinkHash(token))
	if errors.Is(err, gamestats.ErrNoShareLink) {
		return errNoSharedStats
	}
	if err != nil {
		return sharedStatsFailed(c, err)
	}
	owner, err := uuid.Parse(ownerID)
	if err != nil || owner == uuid.Nil || c.Users == nil {
		return errNoSharedStats
	}
	u, err := c.Users.Get(ctx, owner)
	if err != nil {
		return errNoSharedStats
	}
	prefs, def, err := c.advancedStatsFor(ctx, owner)
	if err != nil {
		return sharedStatsFailed(c, err)
	}
	basic, err := c.Lobby.UserStats(owner)
	if err != nil {
		return sharedStatsFailed(c, err)
	}
	out := sharedStatsResponse{DisplayName: u.DisplayName, Basic: sharedBasic{Tally: basic.Tally,
		Unfinished: basic.Unfinished, ByDeck: nonNilGroups(basic.ByDeck), ByCommander: nonNilGroups(basic.ByCommander),
		TopCards: basic.TopCards}}
	if out.Basic.TopCards == nil {
		out.Basic.TopCards = []gamestats.TopCard{}
	}
	if prefs.Effective(def.Advanced) && c.AdvancedStats != nil {
		ov, err := c.advancedOverview(ctx, owner, makeup, rng)
		if err != nil {
			return sharedStatsFailed(c, err)
		}
		// A shared view has no deck pages yet, and a deck id is the
		// owner's library key, not something a viewer needs.
		for i := range ov.Decks {
			ov.Decks[i].DeckID = ""
		}
		if ov.MostPlayedDeck != nil {
			ov.MostPlayedDeck.DeckID = ""
		}
		out.Advanced, out.Overview = true, &ov
	}
	w.Header().Set("Cache-Control", "no-store")
	// A link posted in Discord is not something to index.
	w.Header().Set("X-Robots-Tag", "noindex")
	return writeJSON(w, http.StatusOK, out)
}

func nonNilGroups(g []gamestats.Group) []gamestats.Group {
	if g == nil {
		return []gamestats.Group{}
	}
	return g
}

func sharedStatsFailed(c Config, err error) error {
	c.logger().Error("reading shared statistics failed", "err", err)
	return httpError(http.StatusInternalServerError, "could not load these statistics; try again")
}
