package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avatar Kyoshi — Legendary Creature — Avatar, 5/4 (colorless):
//
//	"Lands you control have trample and hexproof.
//	 {T}: Add X mana of any one color, where X is the greatest power
//	 among creatures you control."
//
// The BACK face of The Legend of Kyoshi ("<oracle_id>#1", ADR 0034),
// reached only through chapter III's exile-and-return
// (the_legend_of_kyoshi.go).
//
// The mana ability is Selvala, Heart of the Wilds' clause with "any
// ONE color" instead of "any combination" — ProducedOneColor rather
// than ProducedAnyCombinationOfColors — and no {G} in the cost, so a
// bare {T} with five power on the board adds five mana of ONE colour,
// the controller's pick, rather than five independent picks.
//
// "Lands you control have trample and hexproof" is a plain Layer 6
// keyword grant (b16GrantKeywords) over the controller's lands.
// Trample has nothing to apply to unless a land is also a creature;
// hexproof matters the moment anything targets one, and the targeting
// gate honours a hexproof grant from any source, keyword-badge or
// static (the protection-style keyword gate landed with #77 — no
// caveat needed here, unlike an older grant declared before it did).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     theLegendOfKyoshiOracleID + "#1",
		Name:         "Avatar Kyoshi",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			}, "trample", "hexproof"),
		},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(func(g *game.Game, controller, _ uuid.UUID) int {
				return b42GreatestPowerControlledBy(g, controller)
			}),
			Label: "{T}: Add X mana of any one color, where X is the greatest power among creatures you control",
		}},
	})
}
