package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angrath's Marauders — 4/4 Creature — Human Pirate for {5}{R}{R}:
//
//	"If a source you control would deal damage to a permanent or
//	player, it deals double that damage to that permanent or player
//	instead."
//
// The first damage-modifying replacement in the catalog — S17 built
// the pipeline and Fog only ever cancelled events; this one rewrites
// the amount. Everything routed through the damage pipeline doubles:
// combat damage, Reckless Fireweaver pings, a cracked Bombardment.
//
// CR 615.9 / 616: with a second doubler (Solphim) the affected
// player orders them, and doubling twice is x4 either way, so the
// order prompt is invisible here — unlike the Doubling Season /
// Hardened Scales pair, where order changes the result.
func init() {
	Register(Spec{
		OracleID: "2d4976d4-649c-4d42-ac5a-ada4b46a480c",
		Name:     "Angrath's Marauders",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
					return false
				}
				return damageSourceControlledBy(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
		}},
	})
}

// damageSourceControlledBy reports whether the damage event's source
// is a permanent (or spell) controlled by the given player. A source
// that has already left the battlefield — a Bolt in the graveyard by
// the time damage resolves — is looked up wherever it now is, which
// keeps its controller readable.
func damageSourceControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageSource == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	return ok && c.Controller == controller
}
