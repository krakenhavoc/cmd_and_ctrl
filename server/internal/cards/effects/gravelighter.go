package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravelighter — Creature — Spirit {2}{B}, 2/2 (Kamigawa: Neon Dynasty,
// S82):
//
//	"Flying
//	 When this creature enters, draw a card if a creature died this turn.
//	 Otherwise, each player sacrifices a creature of their choice."
//
// Blacksnag Buzzard's "a creature died this turn" tally, read as the trigger
// resolves (it is no intervening "if"), then either the draw or Abyssal
// Gorestalker's each-player sacrifice, of one creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "45ef5ea9-70a6-4999-8778-6a11cb97d28e",
		Name:            "Gravelighter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Gravelighter — draw a card if a creature died this turn; otherwise each player sacrifices a creature",
				func(g *game.Game, item *game.StackItem) error {
					if b11CreaturesDiedThisTurn(g) > 0 {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					}
					return jmpB1EachPlayerSacrificesN(1, "Gravelighter — choose a creature to sacrifice")(g, item)
				})),
		},
	})
}
