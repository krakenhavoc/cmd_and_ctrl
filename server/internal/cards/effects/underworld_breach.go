package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Underworld Breach — Enchantment {1}{R}:
//
//	"Each nonland card in your graveyard has escape. The escape cost is
//	 equal to the card's mana cost plus exile three other cards from
//	 your graveyard.
//	 At the beginning of the end step, sacrifice this enchantment."
//
// The STANDING half of ADR 0066, and the reason a standing permission
// is derived from the battlefield rather than stored on the player.
// Breach's first line is not an event — it is true continuously, of
// whatever is in the graveyard at the moment you ask (CR 702.138), so
// a card milled with Breach on the table has escape immediately and a
// card in the graveyard loses it the instant Breach leaves. Two
// Breaches compose and one of them being sacrificed does not revoke
// the other's permission, for free, because nothing is written down.
//
// Compare The Grim Captain's Locker, which grants the same keyword
// with the same cost shape: that one is an ABILITY that resolves, so
// CR 611.2c locks its set. Same model, opposite scope, and the
// difference is a Spec slot versus a call at resolution.
//
// "The escape cost is equal to the card's mana cost plus exile three
// other cards" — so the permission names no Cost and the synthesised
// offer takes the covered card's printed mana cost, with the three
// exiles as the same AlternativeCost component the printed escape
// keyword uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27e0948b-9916-473b-8d8c-a51bdfbc7457",
		Name:         "Underworld Breach",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			Filter:                  game.PermissionFilter{NonLandOnly: true},
			AltCostKey:              "escape",
			ExileOtherFromGraveyard: 3,
			Label:                   "Escape — its mana cost, Exile three other cards from your graveyard",
		}},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Underworld Breach — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
