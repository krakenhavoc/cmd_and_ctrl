package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frontline Strategist — Creature — Human Soldier {W}, 1/1:
//
//	"Morph {W} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)
//	 When this creature is turned face up, prevent all combat damage
//	 non-Soldier creatures would deal this turn."
//
// Morph is the engine's (CR 702.37). The trigger's shield is #2026's
// negation over a subtype, read as each creature would deal combat
// damage (CR 609.7b); the ruling: "If the creature that is dealing the
// damage is not on the battlefield, use its creature type right before
// it left the battlefield" (CR 608.2h).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6e202114-48f3-46d2-bc54-df766d149d9d",
		Name:         "Frontline Strategist",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Morph("{W}"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Frontline Strategist — prevent all combat damage non-Soldier creatures would deal this turn",
				Do(combatShieldAgainstCreatures(exceptSubtypes("Soldier")))),
		},
	})
}
