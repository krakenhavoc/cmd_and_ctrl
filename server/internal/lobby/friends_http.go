package lobby

// friends_http.go is ADR 0213 slice 3 (#3313): friends, personal blocks
// and the two social settings, over internal/social.
//
//	GET    /me/friends                          friends, requests in, requests out
//	POST   /me/friends/requests                 ask: {user_id} or {game_id, player_id}
//	POST   /me/friends/requests/{user_id}/accept
//	POST   /me/friends/requests/{user_id}/ignore
//	DELETE /me/friends/requests/{user_id}       cancel your own request
//	DELETE /me/friends/{user_id}                unfriend, silently
//	GET    /me/blocks                           the people you blocked
//	GET    /me/blocks/{user_id}                 whether you blocked one
//	PUT    /me/blocks/{user_id}                 block
//	DELETE /me/blocks/{user_id}                 unblock
//	GET    /me/social                           your two settings
//	PUT    /me/social                           change one or both
//
// Same caller rule as the rest of /me/*: a signed-in person. A guest seat
// and the admin token are credentials, not people, and get 403. With no
// database (c.Social nil) every route answers 503.
//
// WHO YOU CAN ASK (§2): someone you have shared a table with
// (Lobby.IsTablemate, the whole relation, so the Played with list's cap
// of 50 does not hide anyone you can ask). There is no search by name.
// From a seat the client sends {game_id, player_id}, because the game
// view carries no account ids; the server resolves the seat, and the
// answer does not echo the account id either.
//
// REFUSALS THAT DO NOT TELL (§3): a request to someone who blocked you
// and a request to someone who takes none answer the same 403. A request
// to someone you blocked says so, because you are the one who can lift
// it.

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// FriendEntry is one person in GET /me/friends or GET /me/blocks. Since
// is when the friendship, the request or the block was made, in Unix
// milliseconds.
type FriendEntry struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Since       int64  `json:"since"`
}

// friendsResponse is the body of GET /me/friends. Friends are in name
// order; requests newest first. Outgoing includes a request its receiver
// ignored: the sender is not told (§2).
type friendsResponse struct {
	Friends  []FriendEntry `json:"friends"`
	Incoming []FriendEntry `json:"incoming"`
	Outgoing []FriendEntry `json:"outgoing"`
	// MaxFriends and MaxOutgoing are the limits, for the page's copy.
	MaxFriends  int `json:"max_friends"`
	MaxOutgoing int `json:"max_outgoing"`
}

// friendRequestBody is POST /me/friends/requests: user_id, or game_id
// and player_id for a seat at a table you sit at.
type friendRequestBody struct {
	UserID   string `json:"user_id,omitempty"`
	GameID   string `json:"game_id,omitempty"`
	PlayerID string `json:"player_id,omitempty"`
}

// friendRequestResponse says what asking did. Outcome is social.Outcome:
// "sent", "pending" (already was), "friends" (they had asked you, so you
// are friends now) or "already_friends". UserID echoes the body's
// user_id, and is absent for a seat.
type friendRequestResponse struct {
	Outcome     string `json:"outcome"`
	UserID      string `json:"user_id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// blocksResponse is the body of GET /me/blocks, most recent first.
type blocksResponse struct {
	Blocks []FriendEntry `json:"blocks"`
}

// blockResponse is GET /me/blocks/{user_id}.
type blockResponse struct {
	Blocked bool `json:"blocked"`
}

// socialResponse is GET and PUT /me/social.
type socialResponse struct {
	// Presence is "Who can see what I'm doing on CMD & CTRL": "friends"
	// (default), "played_with" or "nobody". Stored now; slice 4 reads it.
	Presence string `json:"presence"`
	// Requests is "Who can send me friend requests": "played_with"
	// (default) or "nobody".
	Requests string `json:"requests"`
}

type putSocialBody struct {
	Presence *string `json:"presence"`
	Requests *string `json:"requests"`
}

// The words of the refusals.
const (
	errFriendsOff        = "friends are not available on this server"
	errNotPlayedWith     = "no such person; you can send a friend request to someone you have played with"
	errRequestRefused    = "can't send a friend request to that person"
	errNoFriendRequest   = "no such friend request"
	errNoSuchPerson      = "no such person"
	errSeatNotAPerson    = "that seat isn't signed in, so there is no account to add"
	errNotAtThatTable    = "you can add a friend from a seat at a table you sit at"
	errInvalidUserID     = "invalid user_id"
	errFriendListFull    = "your friends list is full; remove someone first"
	errTheirListFull     = "that person's friends list is full"
	errTooManyPending    = "you have too many friend requests waiting for an answer; cancel one first"
	errTooManyThisHour   = "you've sent a lot of friend requests in the last hour; try again later"
	errCannotTargetSelf  = "that is you"
	errSocialPrefInvalid = "presence is friends, played_with or nobody; requests is played_with or nobody"
)

// socialCaller is the signed-in person behind a /me/friends, /me/blocks
// or /me/social call, once the feature is known to be on.
func socialCaller(c Config, r *http.Request) (uuid.UUID, error) {
	p, err := signedInUser(r)
	if err != nil {
		return uuid.Nil, err
	}
	if c.Social == nil {
		return uuid.Nil, httpError(http.StatusServiceUnavailable, errFriendsOff)
	}
	return p.UserID, nil
}

// pathUserID is {user_id} as a uuid; yourself is refused.
func pathUserID(r *http.Request, caller uuid.UUID) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(r.PathValue("user_id")))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, httpError(http.StatusBadRequest, errInvalidUserID)
	}
	if id == caller {
		return uuid.Nil, httpError(http.StatusBadRequest, errCannotTargetSelf)
	}
	return id, nil
}

func entriesOf(ps []social.Person) []FriendEntry {
	out := make([]FriendEntry, 0, len(ps))
	for _, p := range ps {
		out = append(out, FriendEntry{
			UserID: p.UserID, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL, Since: p.At.UnixMilli(),
		})
	}
	return out
}

// socialError maps a social error onto the route's answer.
func socialError(c Config, err error, what string) error {
	switch {
	case errors.Is(err, social.ErrRefused):
		return httpError(http.StatusForbidden, errRequestRefused)
	case errors.Is(err, social.ErrYouBlocked):
		return httpError(http.StatusForbidden, errYouBlocked)
	case errors.Is(err, social.ErrFriendCap):
		return httpError(http.StatusConflict, errFriendListFull)
	case errors.Is(err, social.ErrTheirFriendCap):
		return httpError(http.StatusConflict, errTheirListFull)
	case errors.Is(err, social.ErrTooManyPending):
		return httpError(http.StatusConflict, errTooManyPending)
	case errors.Is(err, social.ErrRateLimited):
		return httpError(http.StatusTooManyRequests, errTooManyThisHour)
	case errors.Is(err, social.ErrNoRequest):
		return httpError(http.StatusNotFound, errNoFriendRequest)
	case errors.Is(err, social.ErrSelf):
		return httpError(http.StatusBadRequest, errCannotTargetSelf)
	}
	c.logger().Error(what+" failed", "err", err)
	return httpError(http.StatusInternalServerError, "something went wrong; try again")
}

// myFriends is GET /me/friends.
func myFriends(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	o, err := c.Social.Overview(r.Context(), me.String())
	if err != nil {
		return socialError(c, err, "listing a user's friends")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, friendsResponse{
		Friends:     entriesOf(o.Friends),
		Incoming:    entriesOf(o.Incoming),
		Outgoing:    entriesOf(o.Outgoing),
		MaxFriends:  social.MaxFriends,
		MaxOutgoing: social.MaxOutgoing,
	})
}

// postFriendRequest is POST /me/friends/requests.
//
// Order of refusals: the caller (403, then 503 with no database), the
// body (400), the seat when one is named (404 for no such table or seat,
// 403 when the caller does not sit there, 422 for a seat with no
// account), the relation (404, the same words for a stranger and an id
// that is nobody's), then social.Request's own answers.
func postFriendRequest(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	var body friendRequestBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.UserID = strings.TrimSpace(body.UserID)
	body.GameID = strings.TrimSpace(body.GameID)
	body.PlayerID = strings.TrimSpace(body.PlayerID)

	var target uuid.UUID
	switch {
	case body.UserID != "" && (body.GameID != "" || body.PlayerID != ""):
		return httpError(http.StatusBadRequest, "name the person with user_id, or with game_id and player_id, not both")
	case body.UserID != "":
		id, err := uuid.Parse(body.UserID)
		if err != nil || id == uuid.Nil {
			return httpError(http.StatusBadRequest, errInvalidUserID)
		}
		target = id
	case body.GameID != "" && body.PlayerID != "":
		id, err := c.seatAccount(me, body.GameID, body.PlayerID)
		if err != nil {
			return err
		}
		target = id
	default:
		return httpError(http.StatusBadRequest, "name the person: user_id, or game_id and player_id")
	}
	if target == me {
		return httpError(http.StatusBadRequest, errCannotTargetSelf)
	}

	mate, err := c.Lobby.IsTablemate(me, target)
	if err != nil {
		return socialError(c, err, "checking a friend request's tablemate")
	}
	if !mate {
		return httpError(http.StatusNotFound, errNotPlayedWith)
	}
	outcome, err := c.Social.Request(r.Context(), me.String(), target.String())
	if err != nil {
		return socialError(c, err, "sending a friend request")
	}
	out := friendRequestResponse{Outcome: string(outcome), DisplayName: c.displayNameOf(r.Context(), target)}
	if body.UserID != "" {
		out.UserID = target.String()
	}
	return writeJSON(w, http.StatusOK, out)
}

// seatAccount resolves {game_id, player_id} to the seat's account, for a
// caller who sits at that table.
func (c Config) seatAccount(me uuid.UUID, gameID, playerID string) (uuid.UUID, error) {
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return uuid.Nil, httpError(http.StatusBadRequest, "invalid game_id")
	}
	pid, err := uuid.Parse(playerID)
	if err != nil {
		return uuid.Nil, httpError(http.StatusBadRequest, "invalid player_id")
	}
	meta, err := c.Lobby.Get(gid)
	if err != nil {
		return uuid.Nil, httpError(http.StatusNotFound, "no such table")
	}
	if _, seated := seatOfUser(meta.Players, me); !seated {
		return uuid.Nil, httpError(http.StatusForbidden, errNotAtThatTable)
	}
	for _, s := range meta.Players {
		if s.PlayerID != pid {
			continue
		}
		if s.IsBot || s.IsAgent || s.UserID == "" {
			return uuid.Nil, httpError(http.StatusUnprocessableEntity, errSeatNotAPerson)
		}
		id, err := uuid.Parse(s.UserID)
		if err != nil {
			return uuid.Nil, httpError(http.StatusUnprocessableEntity, errSeatNotAPerson)
		}
		return id, nil
	}
	return uuid.Nil, httpError(http.StatusNotFound, "no such seat")
}

// displayNameOf is the person's current display name, or "".
func (c Config) displayNameOf(ctx context.Context, id uuid.UUID) string {
	u, err := c.userStore().Get(ctx, id)
	if err != nil {
		return ""
	}
	return u.DisplayName
}

// acceptFriendRequest is POST /me/friends/requests/{user_id}/accept.
func acceptFriendRequest(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	from, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if err := c.Social.Accept(r.Context(), me.String(), from.String()); err != nil {
		return socialError(c, err, "accepting a friend request")
	}
	return writeJSON(w, http.StatusOK, friendRequestResponse{
		Outcome: string(social.OutcomeFriends), UserID: from.String(), DisplayName: c.displayNameOf(r.Context(), from),
	})
}

// ignoreFriendRequest is POST /me/friends/requests/{user_id}/ignore.
func ignoreFriendRequest(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	from, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if err := c.Social.Ignore(r.Context(), me.String(), from.String()); err != nil {
		return socialError(c, err, "ignoring a friend request")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// cancelFriendRequest is DELETE /me/friends/requests/{user_id}: withdraw
// your own request to that person.
func cancelFriendRequest(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	to, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if err := c.Social.Cancel(r.Context(), me.String(), to.String()); err != nil {
		return socialError(c, err, "cancelling a friend request")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// removeFriend is DELETE /me/friends/{user_id}. Silent, from either
// side, and idempotent.
func removeFriend(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	other, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if err := c.Social.Unfriend(r.Context(), me.String(), other.String()); err != nil {
		return socialError(c, err, "removing a friend")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// myBlocks is GET /me/blocks.
func myBlocks(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	bl, err := c.Social.Blocks(r.Context(), me.String())
	if err != nil {
		return socialError(c, err, "listing a user's blocks")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, blocksResponse{Blocks: entriesOf(bl)})
}

// myBlock is GET /me/blocks/{user_id}: whether you blocked them. It
// never says whether they blocked you.
func myBlock(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	target, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	blocked, err := c.Social.IsBlocking(r.Context(), me.String(), target.String())
	if err != nil {
		return socialError(c, err, "reading a block")
	}
	return writeJSON(w, http.StatusOK, blockResponse{Blocked: blocked})
}

// putBlock is PUT /me/blocks/{user_id}: block someone, which ends any
// friendship and request between you (ADR 0213 §3). The person must be
// an account; nobody is told.
func putBlock(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	target, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if _, err := c.userStore().Get(r.Context(), target); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return httpError(http.StatusNotFound, errNoSuchPerson)
		}
		return socialError(c, err, "reading a block's target")
	}
	if err := c.Social.Block(r.Context(), me.String(), target.String()); err != nil {
		return socialError(c, err, "blocking someone")
	}
	return writeJSON(w, http.StatusOK, blockResponse{Blocked: true})
}

// deleteBlock is DELETE /me/blocks/{user_id}: unblock. It brings back
// nothing the block ended.
func deleteBlock(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	target, err := pathUserID(r, me)
	if err != nil {
		return err
	}
	if err := c.Social.Unblock(r.Context(), me.String(), target.String()); err != nil {
		return socialError(c, err, "unblocking someone")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// mySocial is GET /me/social.
func mySocial(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	p, err := c.Social.Prefs(r.Context(), me.String())
	if err != nil {
		return socialError(c, err, "reading social settings")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, socialResponse{Presence: string(p.Presence), Requests: string(p.Requests)})
}

// putMySocial is PUT /me/social. A setting left out keeps its value; at
// least one must be named.
func putMySocial(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := socialCaller(c, r)
	if err != nil {
		return err
	}
	var body putSocialBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Presence == nil && body.Requests == nil {
		return httpError(http.StatusBadRequest, "name a setting: presence or requests")
	}
	var (
		presence *social.Presence
		requests *social.Requests
	)
	if body.Presence != nil {
		v := social.Presence(*body.Presence)
		if !v.Valid() {
			return httpError(http.StatusBadRequest, errSocialPrefInvalid)
		}
		presence = &v
	}
	if body.Requests != nil {
		v := social.Requests(*body.Requests)
		if !v.Valid() {
			return httpError(http.StatusBadRequest, errSocialPrefInvalid)
		}
		requests = &v
	}
	p, err := c.Social.SetPrefs(r.Context(), me.String(), presence, requests)
	if err != nil {
		return socialError(c, err, "saving social settings")
	}
	return writeJSON(w, http.StatusOK, socialResponse{Presence: string(p.Presence), Requests: string(p.Requests)})
}
