package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dire Wolf Prowler — Creature — Wolf {2}{G}, 2/2 (Adventures in the
// Forgotten Realms):
//
//	"{1}{G}: This creature gets +2/+2 and gains haste until end of turn.
//	 Activate only once each turn."
//
// Both halves are the one resolution, pinned to the Prowler as it was
// activated (a new object, CR 400.7, gets nothing). "Once each turn" gates
// the offer at the announce (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 and gains haste until end of turn. Activate only once each turn."
	pump := thisGetsUntilEndOfTurn(2, 2, "Dire Wolf Prowler — +2/+2")
	haste := thisGainsKeywordUntilEndOfTurn("haste", "Dire Wolf Prowler — haste")
	Register(Spec{
		OracleID:     "d558b465-75c6-450e-b1f2-5843ad67de71",
		Name:         "Dire Wolf Prowler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{1}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := pump(g, item); err != nil {
					return err
				}
				return haste(g, item)
			},
		}},
	})
}
