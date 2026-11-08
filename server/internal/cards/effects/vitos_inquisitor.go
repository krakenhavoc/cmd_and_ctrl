package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "03ee49c9-95a3-44f1-b010-f7c4b2f86a7c",
		Name:         "Vito's Inquisitor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice another creature or artifact: Put a +1/+1 counter on this creature. It gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				Mana: "{B}",
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
