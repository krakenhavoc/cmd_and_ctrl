package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vito's Inquisitor — Creature — Vampire Knight {3}{B}, 3/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{B}, Sacrifice another creature or artifact: Put a +1/+1 counter on
//	 this creature. It gains menace until end of turn."
//
// Bartolomé del Presidio's cost with {B} added, both paid as the ability
// is activated (CR 602.2b); "another" is object identity, so the
// Inquisitor can't pay with itself. The counter and the menace (a layer 6
// grant ending in the cleanup step, CR 514.2) go on the object the
// ability came from; one that left and came back is a new object and
// gets neither (CR 400.7).
//
// No simplification.
func init() {
	const label = "Vito's Inquisitor — +1/+1 counter and menace until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "03ee49c9-95a3-44f1-b010-f7c4b2f86a7c",
		Name:         "Vito's Inquisitor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice another creature or artifact: Put a +1/+1 counter on this creature. It gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				Mana: "{B}",
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
