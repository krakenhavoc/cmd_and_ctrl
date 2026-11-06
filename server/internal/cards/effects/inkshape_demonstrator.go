package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inkshape Demonstrator — Creature — Elephant Cleric {3}{W}, 3/4 (Secrets
// of Strixhaven, #4456):
//
//	"Ward {2}
//	 Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, this creature gets +1/+0 and gains lifelink until end of turn."
//
// Ward is the ward helper (CR 702.21); Repartee is Informed Inkwright's
// trigger (informedInkwrightCast). The pump and lifelink belong to this
// permanent and are skipped if it has left or come back as a new object
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4993651f-f161-4007-b485-a0c66a02f91b",
		Name:         "Inkshape Demonstrator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Inkshape Demonstrator — ward {2}"),
			On(game.EventCast, informedInkwrightCast,
				"Inkshape Demonstrator — Repartee: +1/+0 and lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return socB27ThisGetsAndGains(NewContext(g, item), 1, "lifelink", "Inkshape Demonstrator")
				}),
		},
	})
}
