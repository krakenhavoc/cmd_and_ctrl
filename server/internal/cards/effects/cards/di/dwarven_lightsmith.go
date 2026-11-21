package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Lightsmith — Creature — Dwarf Cleric {5}{W}:
//
//	"Assist (Another player can pay up to {5} of this spell's cost.)
//	 When this creature enters, creatures your team controls get +1/+1
//	 until end of turn."
//
// C2 has no teams, so "your team" is just its controller: the ETB is
// the Consul's Lieutenant pump over every creature you control. The
// set is read once, as the trigger resolves (CR 611.2c), and it
// includes the Lightsmith itself, as the printed text does.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb9210eb-f5cb-430d-b893-b1d814117043",
		Name:            "Dwarven Lightsmith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dwarven Lightsmith — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     1,
						Toughness: 1,
						Label:     "Dwarven Lightsmith — +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
