package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b220_helpers.go — shared pieces for Rivals of Ixalan batch 220
// (S82, #5531).

// S82OldB220NextTurnInstantSorceryBan is "each opponent can't cast instant or
// sorcery spells during that player's next turn" (Sphinx's Decree): every
// opponent still in the game gets a cast ban over instants and sorceries that
// reaches only their next turn (CastBanRule.OnTurnBegun) and ends with it.
func S82OldB220NextTurnInstantSorceryBan(ctx *Context, label string) error {
	g := ctx.Game
	for _, opp := range ctx.Opponents() {
		if err := (RestrictCasting{
			Player: opp,
			Rule: game.CastBanRule{
				Kind:        game.CastBanOutright,
				Filter:      game.PermissionFilter{InstantOrSorceryOnly: true},
				OnTurnBegun: g.NextTurnNumberForEffect(opp),
			},
			Label:    label,
			Duration: g.UntilEndOfYourNextTurnDuration(opp),
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB220DifferentlyNamedLands is the number of differently named lands
// `player` controls, names compared case-insensitively on the battlefield
// object as it stands (CR 201.2).
func S82OldB220DifferentlyNamedLands(g *game.Game, player uuid.UUID) int {
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsLand() {
			continue
		}
		if name := strings.ToLower(strings.TrimSpace(c.Name)); name != "" {
			names[name] = true
		}
	}
	return len(names)
}

// S82OldB220UnattachedHost is the permanent an Equipment became unattached
// from (Grafted Exoskeleton's read, exported for Captain's Hook): the event's
// target when the link broke with the Equipment still on the battlefield, and
// what the departed Equipment was last attached to when it left.
func S82OldB220UnattachedHost(ctx *Context) uuid.UUID { return graftedHost(ctx) }
