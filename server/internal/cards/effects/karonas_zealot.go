package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karona's Zealot — Creature — Human Cleric {4}{W}, 2/5:
//
//	"Morph {3}{W}{W}
//	 When this creature is turned face up, all damage that would be dealt
//	 to it this turn is dealt to target creature instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn from the
// Zealot to the target. The ruling: a target that has left, or is no
// longer a creature, redirects nothing (CR 614.9), and later damage that
// turn stays on the Zealot.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "dff66bcf-e126-49d2-b67e-ea3b0a38d390",
		Name:             "Karona's Zealot",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisIsTurnedFaceUp("Karona's Zealot — damage to it this turn is dealt to target creature instead",
				func(g *game.Game, item *game.StackItem) error {
					return RedirectDamage{Protect: ShieldThis, To: RedirectToClause(0)}.Apply(NewContext(g, item))
				}), TargetCreature("target creature")),
		},
	})
}
