package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apocalypse Chime — Artifact {2} (Homelands, S82 OLD b827, #5635):
//
//	"{2}, {T}, Sacrifice this artifact: Destroy all nontoken permanents with a
//	 name originally printed in the Homelands expansion. They can't be
//	 regenerated."
//
// City in a Bottle's "name originally printed in" test, over CR 206.3c's list
// of Homelands names (game.HomelandsNames; the Chime is on it too, but it has
// already been sacrificed as a cost). A token is not a card with a printed
// name (CR 111.4), so only nontoken permanents are destroyed, whoever controls
// them, and regeneration can't save them (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "995e2412-3f69-4aaf-a5d6-ae0ecbb95061",
		Name:         "Apocalypse Chime",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this artifact: Destroy all nontoken permanents with a name originally printed in the Homelands expansion. They can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{
				Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true,
			}},
			Cost: effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyAllMatching{
					Match: effects.And(
						effects.Not(effects.IsTokenPredicate()),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.IsHomelandsName(c.Name) },
					),
					CantBeRegenerated: true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
