package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildsear, Scouring Maw — Legendary Creature — Elemental Wolf {3}{R}{G},
// 6/6 (Bloomburrow Commander):
//
//	"Trample
//	 Enchantment spells you cast from your hand have cascade."
//
// The granted cascade is GrantsCascade (Quandrix, the Proof's shape): a
// battlefield trigger on every cast of its controller that queues the same
// cascade trigger for an enchantment spell. "From your hand" is the spell's
// own cast zone (CastFromOwnersHand), so a cast from exile, a graveyard or
// the command zone never cascades.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ceac4c6-e62e-4445-852c-b2c621e4a83b",
		Name:            "Wildsear, Scouring Maw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			GrantsCascade("Wildsear, Scouring Maw", func(spell game.Card, source *game.Card, g *game.Game) bool {
				return spell.IsEnchantment() &&
					spell.Owner == source.Controller &&
					CastFromOwnersHand()(g, source.Controller, spell)
			}),
		},
	})
}
