package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drover Grizzly — Creature — Bear Mount {2}{G}:
//
//	"Whenever this creature attacks while saddled, creatures you
//	 control gain trample until end of turn.
//	 Saddle 1"
//
// The set of creatures is fixed when the trigger resolves (CR 611.2c),
// so a creature that arrives later in the turn does not gain trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "290fe09e-04ab-48d1-a60e-db47ea8c4d00",
		Name:         "Drover Grizzly",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(1)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Drover Grizzly — creatures you control gain trample", func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"trample"},
					Label:    "Drover Grizzly — trample until end of turn",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
