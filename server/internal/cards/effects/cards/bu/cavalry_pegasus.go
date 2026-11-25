package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavalry Pegasus — Creature — Pegasus {1}{W}, 1/1 (Commander 2020, S82 OLD,
// #5513):
//
//	"Flying
//	 Whenever this creature attacks, each attacking Human gains flying until
//	 end of turn."
//
// The set of attacking Humans is read as the trigger resolves and locked then
// (CR 611.2c): every attacking Human, whoever controls it, which is the
// Pegasus's own team in a normal attack. The grant is a layer 6 keyword until
// end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f6fd5d9-d123-4986-a18b-206cc2dfb3f1",
		Name:            "Cavalry Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Cavalry Pegasus — each attacking Human gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil,
						effects.And(effects.AttackingCreature(), effects.Subtype("Human")),
						"Cavalry Pegasus — each attacking Human gains flying", game.AddKeywordsMod("flying"))
				}),
		},
	})
}
