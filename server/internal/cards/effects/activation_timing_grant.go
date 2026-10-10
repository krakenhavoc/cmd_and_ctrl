package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_timing_grant.go — #2797: the resolving-spell half of
// activation_timing.go. The derived constructors there
// (ThisSourcesLoyaltyAbilitiesAtInstantSpeed,
// LoyaltyAbilitiesOfYourPlaneswalkersAtInstantSpeed) are statements a
// permanent or an emblem makes for as long as it is there. A spell that
// says "until end of turn, you may activate loyalty abilities of Jace
// planeswalkers you control … any time you could cast an instant" has no
// presence to derive from, so the statement is stored on the player with
// a duration (game/activation_timing_grant.go).

// GrantLoyaltyAbilitiesAtInstantSpeed is "Until end of turn, you may
// activate loyalty abilities of <subtype> planeswalkers you control on
// any player's turn any time you could cast an instant." — Jace's
// Machinations. Subtype is the planeswalker subtype ("Jace"); empty is
// every planeswalker you control.
//
// "On any player's turn" is not modelled separately and does not need
// to be: it is what "any time you could cast an instant" already means,
// spelled out. CR 606.3's once-per-permanent-per-turn limit is
// untouched, so a Jace that already activated this turn stays shut.
//
// The window is this turn by default (a zero Duration is stamped "until
// end of turn" by the one write path); the statement lasts exactly that
// long whether the spell that granted it is still anywhere or not.
type GrantLoyaltyAbilitiesAtInstantSpeed struct {
	// Subtype is the planeswalker subtype the statement is about.
	Subtype string
	// Player is who may activate. uuid.Nil is the controller.
	Player uuid.UUID
	// Label is the clause as printed, for the log.
	Label string
}

// Apply stores the statement. Caller is inside the resolution frame
// (holds g.mu write).
func (e GrantLoyaltyAbilitiesAtInstantSpeed) Apply(ctx *Context) error {
	player := e.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.GrantLoyaltyActivationTimingForEffect(player, game.TimingFlash, e.Subtype, e.Label, ctx.Source(), game.Duration{})
	return nil
}
