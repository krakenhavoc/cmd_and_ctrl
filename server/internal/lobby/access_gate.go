package lobby

// access_gate.go is ADR 0156 §1: sign-up gated by a role in the C2
// Discord server, checked at sign-in (discordCallback) and again when a
// signed-in session is renewed (renewSession), plus GET /community
// (§6), which tells the client the invite link and whether the gate is
// on.
//
// The gate is on when CMDCTRL_DISCORD_COMMUNITY_GUILD_ID and
// CMDCTRL_DISCORD_PLAYER_ROLE_ID are both set (Config.Discord.Gate).
// Off, nothing here runs and sign-in is what it always was: no member
// call, no write to the two access columns, renewal unchanged.

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// gateCallTimeout bounds each Discord round trip the gate makes, so a
// hung Discord holds a sign-in redirect or a renewal for seconds, not
// for as long as the browser waits.
const gateCallTimeout = 10 * time.Second

// The gate page's routes (ADR 0156 §5). The client owns the page; these
// are where the callback sends a person it turned away.
const (
	requestAccessPath      = "/#/request-access"
	requestAccessRetryPath = "/#/request-access?retry=1"
)

// AccessStore is what the gate needs from the user database beyond
// users.Store. *users.SQLStore implements it; with no database there is
// none, and no signed-in session exists to re-check.
type AccessStore interface {
	DiscordSubject(ctx context.Context, id uuid.UUID) (string, error)
	RefreshToken(ctx context.Context, subject string) (token string, ok bool, err error)
	SetRefreshToken(ctx context.Context, subject, token string) error
	RecordAccessCheck(ctx context.Context, id uuid.UUID, result discord.Access, at time.Time) error
}

var _ AccessStore = (*users.SQLStore)(nil)

// accessStore is the user store as an AccessStore, when it is one.
func (c Config) accessStore() (AccessStore, bool) {
	s, ok := c.userStore().(AccessStore)
	return s, ok
}

// discordClient is the client every Discord call goes through.
func (c Config) discordClient() *http.Client {
	if c.DiscordHTTPClient != nil {
		return c.DiscordHTTPClient
	}
	return http.DefaultClient
}

// gateApplies reports whether discordID has to pass the gate: the gate
// is on and the account is not an exempt admin.
func (c Config) gateApplies(discordID string) bool {
	return c.Discord.Gate.Enabled() && !gateExempt(c.Admins, discordID)
}

// checkMember is one member check with the gate's timeout.
func (c Config) checkMember(ctx context.Context, accessToken string) (discord.Access, error) {
	ctx, cancel := context.WithTimeout(ctx, gateCallTimeout)
	defer cancel()
	return c.Discord.CheckMember(ctx, c.discordClient(), accessToken)
}

// recordAccess stores a check's answer on the user. A failure is
// logged, never fatal: the answer only feeds the admin view.
func (c Config) recordAccess(ctx context.Context, userID uuid.UUID, result discord.Access) {
	store, ok := c.accessStore()
	if !ok || userID == uuid.Nil {
		return
	}
	if err := store.RecordAccessCheck(ctx, userID, result, c.now()); err != nil {
		c.logger().Warn("recording the access check failed", "user_id", userID, "result", result, "err", err)
	}
}

// signInGate is the callback's check (ADR 0156 §1 item 4). It returns
// admitted == true when sign-in may go on as it always did: the gate is
// off, the account is exempt, or Discord says the person holds the
// role (checked is then true, and the caller records it once the users
// row exists). Otherwise it has already answered the request: a
// redirect to the gate page, or, for an invite link, a guest seat.
//
// Nothing here writes about the person. A refused sign-in leaves no
// users row behind and touches none that exists (ADR 0149).
func signInGate(c Config, w http.ResponseWriter, r *http.Request, entry discord.StateEntry,
	user discord.User, accessToken string) (admitted, checked bool, err error) {
	if !c.gateApplies(user.ID) {
		return true, false, nil
	}
	access, cerr := c.checkMember(r.Context(), accessToken)
	if cerr != nil {
		// Fails closed: the person is here to try again (item 4).
		c.logger().Warn("Discord did not answer the sign-up gate's role check; sign-in refused, the person is offered a retry",
			"err", cerr)
		http.Redirect(w, r, requestAccessRetryPath, http.StatusFound)
		return false, false, nil
	}
	if access == discord.AccessAdmitted {
		return true, true, nil
	}
	inviteFlow := !entry.Unbound() && !entry.Link()
	c.logger().Info("sign-in refused by the sign-up gate", "result", access, "invite_flow", inviteFlow)
	if inviteFlow {
		// The host's link vouches for this one table (owner answer 2).
		return false, true, seatRefusedAsGuest(c, w, r, entry, user.DisplayName())
	}
	// The login page, and linking Discord to a seat already held: both
	// would make an account, so both go to the gate page. A link leaves
	// the seat the guest seat it was.
	http.Redirect(w, r, requestAccessPath, http.StatusFound)
	return false, true, nil
}

// seatRefusedAsGuest is the invite flow's refusal: the person takes the
// seat as a guest named from their Discord display name, with no user
// and no Discord identity, exactly as if they had typed that name on
// the join page. The session is a guest's, CMDCTRL_SESSION_TTL long.
func seatRefusedAsGuest(c Config, w http.ResponseWriter, r *http.Request, entry discord.StateEntry, name string) error {
	if c.Lobby.ConnectionsFull() {
		back := url.Values{}
		back.Set("t", entry.InviteToken)
		back.Set("error", capacity.CodeServerFull)
		http.Redirect(w, r, "/#/games/"+entry.GameID.String()+"/join?"+back.Encode(), http.StatusFound)
		return nil
	}
	meta, playerID, err := c.Lobby.JoinAs(entry.GameID, entry.InviteToken, name, DiscordIdentity{}, uuid.Nil)
	if err != nil {
		return err
	}
	p := auth.Principal{
		Role:     auth.RolePlayer,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     name,
	}
	tok, issued, err := issueFor(r.Context(), c, p, nil)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	frag := url.Values{}
	frag.Set("token", tok)
	frag.Set("game", meta.ID.String())
	frag.Set("player_id", playerID.String())
	frag.Set("expires_at", issued.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"))
	frag.Set("name", name)
	http.Redirect(w, r, "/#/oauth-complete?"+frag.Encode(), http.StatusFound)
	return nil
}

// renewalVerdict is what the re-check at renewal decided.
type renewalVerdict int

const (
	// renewalGo: renew as before (gate off, exempt, admitted, or an
	// outage).
	renewalGo renewalVerdict = iota
	// renewalHold: do not renew; the session runs to its own expiry and
	// the next sign-in is the check (no refresh token, or Discord
	// rejected it).
	renewalHold
	// renewalRefused: the person no longer holds the role. Every session
	// they hold is withdrawn.
	renewalRefused
)

// renewalGate is ADR 0156 §1 item 5: before a half-spent signed-in
// session is re-issued, the stored refresh token is exchanged for an
// access token (the rotated refresh token is stored in its place) and
// the role is asked about again.
func (c Config) renewalGate(ctx context.Context, p auth.Principal) renewalVerdict {
	if !c.gateApplies(p.DiscordID) {
		return renewalGo
	}
	log := c.logger().With("user_id", p.UserID)
	store, ok := c.accessStore()
	if !ok {
		log.Info("session not renewed: the sign-up gate is on and there is no user database to re-check against")
		return renewalHold
	}
	subject, err := store.DiscordSubject(ctx, p.UserID)
	if err != nil {
		log.Info("session not renewed: no Discord identity to re-check", "err", err)
		return renewalHold
	}
	rt, has, err := store.RefreshToken(ctx, subject)
	if err != nil {
		log.Warn("session not renewed: the stored refresh token cannot be opened", "var", users.IdentityKeyEnv, "err", err)
		return renewalHold
	}
	if !has {
		log.Info("session not renewed: no refresh token is stored, so the sign-up gate cannot be re-checked; the next sign-in will")
		return renewalHold
	}

	cctx, cancel := context.WithTimeout(ctx, gateCallTimeout)
	tr, err := c.Discord.ExchangeRefreshToken(cctx, c.discordClient(), rt)
	cancel()
	switch {
	case errors.Is(err, discord.ErrRefreshRejected):
		log.Info("session not renewed: Discord rejected the refresh token; the next sign-in will re-check", "err", err)
		return renewalHold
	case err != nil:
		log.Warn("Discord did not answer the sign-up gate's re-check; renewing anyway", "err", err)
		return renewalGo
	}
	if err := store.SetRefreshToken(ctx, subject, tr.RefreshToken); err != nil {
		log.Error("storing the rotated refresh token failed; the next re-check may find it rejected", "err", err)
	}

	access, err := c.checkMember(ctx, tr.AccessToken)
	if err != nil {
		log.Warn("Discord did not answer the sign-up gate's re-check; renewing anyway", "err", err)
		return renewalGo
	}
	c.recordAccess(ctx, p.UserID, access)
	if access != discord.AccessAdmitted {
		log.Info("session not renewed: the sign-up gate refused the re-check", "result", access)
		return renewalRefused
	}
	return renewalGo
}

// communityResponse is GET /community (ADR 0156 §6).
type communityResponse struct {
	// InviteURL is CMDCTRL_DISCORD_INVITE_URL, or "" when unset, which
	// hides every Discord placement.
	InviteURL string `json:"invite_url"`
	// Gate is whether sign-up needs the admitting role (Elder Dragon
	// Highlander on the C2 server).
	Gate bool `json:"gate"`
}

// community is GET /community: public, no session, cacheable for five
// minutes. It names no guild, role or person.
func community(c Config, w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Cache-Control", "public, max-age=300")
	return writeJSON(w, http.StatusOK, communityResponse{
		InviteURL: c.CommunityInviteURL,
		Gate:      c.Discord.Gate.Enabled(),
	})
}
