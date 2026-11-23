package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devilish Valet — Creature — Devil Warrior {2}{R}, 1/3 (#5368):
//
//	"Trample, haste
//	 Alliance — Whenever another creature you control enters, double this
//	 creature's power until end of turn."
//
// Alliance is the flavor word for "whenever another creature you control
// enters" (Good-Fortune Unicorn's condition, tokens included). Doubling is
// CR 701.10b: a one-time +X/+0 for the power as the trigger resolves, so a
// second creature entering doubles the doubled power again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1e0dc37-20c9-4236-ae01-e0891ce83afb",
		Name:            "Devilish Valet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B13AnotherCreatureYouControlEntered(ev, source, g)
			}, "Devilish Valet — double this creature's power until end of turn",
				effects.Cmdr2B23DoubleThisPower("Devilish Valet — double power")),
		},
	})
}
