package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rank and File — Creature — Zombie {2}{B}{B} 3/3:
//
//	"When this creature enters, green creatures get -1/-1 until end of turn."
//
// The creatures affected are the green creatures on the battlefield as it
// resolves, whoever controls them (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0b8ab40-3c27-48bd-ae19-fb1a155a0522",
		Name:         "Rank and File",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rank and File — green creatures get -1/-1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.OfColor("G")), Power: -1, Toughness: -1,
					Label: "Rank and File — green creatures get -1/-1 until end of turn",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
