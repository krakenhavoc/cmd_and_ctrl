package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ecstatic Awakener // Awoken Demon — Creature — Human Wizard {B}, 1/1 //
// Creature — Demon 4/4 (Midnight Hunt):
//
//	Ecstatic Awakener
//	  "{2}{B}, Sacrifice another creature: Draw a card, then transform this
//	   creature. Activate only once each turn."
//	Awoken Demon
//	  (no abilities)
//
// The sacrifice is a cost (CR 118.3), so it cannot be paid by the Awakener
// itself. "Once each turn" is the engine's per-object activation tally.
// The draw happens first and then the same permanent turns over in place.
//
// No simplification.
func init() {
	const oracle = "26ff3da5-4e02-41d1-a52f-3f8e87b01a77"
	const label = "{2}{B}, Sacrifice another creature: Draw a card, then transform this creature. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Ecstatic Awakener",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.MidB12DrawThenTransformThis,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Awoken Demon",
		Completeness: effects.CompletenessFull,
	})
}
