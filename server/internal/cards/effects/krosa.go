package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krosa — Plane — Dominaria (Commander Masters, #4870):
//
//	"All creatures get +2/+2.
//	 Whenever chaos ensues, you may add {W}{U}{B}{R}{G}."
//
// A plane (ADR 0216 §6). The static is a layer-7c anthem over every
// creature on the battlefield, whoever controls it. The chaos ability
// (CR 311.7) is a "you may" asked as it resolves (CR 603.5), so declining
// adds nothing; the planar controller adds the mana as it resolves (it is
// not a mana ability, CR 605.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2732b605-7c88-4995-b46e-62dc09e68919",
		Name:         "Krosa",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16Anthem(allCreatures, 2, 2)},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverChaosEnsues("Krosa — you may add {W}{U}{B}{R}{G}",
				Do(AddMana{Produced: "{W}{U}{B}{R}{G}"})), "Krosa — add {W}{U}{B}{R}{G}?"),
		},
	})
}
