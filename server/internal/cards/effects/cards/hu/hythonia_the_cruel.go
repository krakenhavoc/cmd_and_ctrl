package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hythonia the Cruel — Legendary Creature — Gorgon {4}{B}{B}, 4/6 (Theros):
//
//	"Deathtouch
//	 {6}{B}{B}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When Hythonia becomes monstrous, destroy all non-Gorgon creatures."
//
// Monstrosity is the shared activated ability (Arbor Colossus' shape) and the
// trigger WhenBecomesMonstrous (CR 701.37b). The destruction is every creature
// on the battlefield that is not a Gorgon as the trigger resolves, Hythonia
// herself (a Gorgon) and any other Gorgon surviving.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48e3d832-2fe9-4ffd-8258-75451100dcfc",
		Name:            "Hythonia the Cruel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{B}{B}"), 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesMonstrous("Hythonia the Cruel — destroy all non-Gorgon creatures",
				effects.Do(effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Gorgon")))})),
		},
	})
}
