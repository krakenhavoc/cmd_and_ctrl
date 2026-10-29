package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirkwood Meditator — Creature — Elf Druid {2}{U}, 2/4 (The Hobbit, HOB):
//
//	"Landfall — Whenever a land you control enters, you may have this
//	 creature's base power and toughness become 4/2 until end of turn."
//
// A "you may" landfall trigger (the question is asked as it resolves,
// CR 603.5). The change is a layer 7b set of the BASE power and toughness
// (CR 613.4b), so +1/+1 counters and pumps still add to the 4/2, and it
// applies only while this is still the permanent that triggered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "399cd159-7376-421c-b001-c26f1cbadced",
		Name:         "Mirkwood Meditator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Landfall("Mirkwood Meditator — you may have its base power and toughness become 4/2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
						"Mirkwood Meditator — base power and toughness become 4/2", game.SetBasePTMods(4, 2)...)
				}),
				"Have Mirkwood Meditator's base power and toughness become 4/2 until end of turn?"),
		},
	})
}
