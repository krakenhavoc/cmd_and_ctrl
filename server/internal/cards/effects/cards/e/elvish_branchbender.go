package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Branchbender — Creature — Elf Druid {2}{G}, 2/2 (Lorwyn, S82 OLD
// b547, #5593):
//
//	"{T}: Until end of turn, target Forest becomes an X/X Treefolk creature
//	 in addition to its other types, where X is the number of Elves you
//	 control."
//
// Vastwood Animist's animation: X is counted as the ability resolves and
// fixed then (CR 611.2c), the Branchbender and any changeling counted. The
// Forest keeps its land and Forest types and gains Creature and Treefolk;
// any player's Forest may be the target.
//
// No simplification.
func init() {
	const label = "Elvish Branchbender — target Forest becomes an X/X Treefolk creature until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "6e9fe674-44c8-4ab4-a0ef-ae71d65bbcc0",
		Name:         "Elvish Branchbender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Until end of turn, target Forest becomes an X/X Treefolk creature in addition to its other types, where X is the number of Elves you control.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target Forest", effects.HasSubtype("Forest")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				forest := effects.FirstLegalBattlefieldTarget(ctx)
				if forest == uuid.Nil {
					return nil
				}
				x := effects.S82OldB547CountYouControl(ctx, "Elf")
				return effects.BecomeCreatureUntilEOT{Target: forest, Types: []string{"Creature"}, Subtypes: []string{"Treefolk"}, SetPower: x, SetToughness: x, Label: label}.Apply(ctx)
			},
		}},
	})
}
