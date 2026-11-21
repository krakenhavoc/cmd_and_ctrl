package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4993651f-f161-4007-b485-a0c66a02f91b",
		Name:         "Inkshape Demonstrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Inkshape Demonstrator — ward {2}"),
			effects.On(game.EventCast, effects.InformedInkwrightCast,
				"Inkshape Demonstrator — Repartee: +1/+0 and lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SocB27ThisGetsAndGains(effects.NewContext(g, item), 1, "lifelink", "Inkshape Demonstrator")
				}),
		},
	})
}
