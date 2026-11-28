package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reparations — Enchantment {1}{W}{U} (Mirage, S82 OLD b800, #5633):
//
//	"Whenever an opponent casts a spell that targets you or a creature you
//	 control, you may draw a card."
//
// One trigger on any opponent's cast: the spell's targets are read as it goes
// on the stack (Speedball's cast read), and it fires when one is you or a
// creature you control. Casting a spell is the trigger event, so each such
// spell triggers once however many of your things it targets. "You may" is a
// real choice (CR 603.5), and a draw a limit stops is not offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cb86deb-f4e5-41c5-bfcd-614a24f3499c",
		Name:         "Reparations",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(
				effects.On(game.EventCast, effects.OldB800ReparationsCastTargetsYou,
					"Reparations — you may draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
					}),
				"Reparations — draw a card?", 1),
		},
	})
}
