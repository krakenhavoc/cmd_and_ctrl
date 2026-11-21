package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Butcher of Malakir — Creature — Vampire Warrior {5}{B}{B}, 5/4:
//
//	"Flying"
//	"Whenever this creature or another creature you control dies, each
//	 opponent sacrifices a creature of their choice."
//
// Grave Pact on a body, and the "THIS creature or another" wording
// matters: the Butcher's own death triggers it, so removal aimed at
// the Butcher still costs each opponent a creature. Compare Grave
// Pact, which says only "a creature you control" — the enchantment
// can't die to creature removal in the first place, so it never needed
// the clause.
//
// Expensive enough that it's a payoff rather than an engine, but it
// closes games in a way the enchantment can't: it attacks.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a85197ab-dc94-4b72-9716-8dbdbbe90ff8",
		Name:            "Butcher of Malakir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			// "This creature OR another creature you control" —
			// the source's own death counts, and diedCreature has
			// already resolved the card post-move so its
			// controller is still readable.
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Butcher of Malakir — each opponent sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Creature(),
				Label:            "a creature",
			})), game.Purpose{DeathPayoff: true, Answers: game.AnswerRemove}),
		},
	})
}
