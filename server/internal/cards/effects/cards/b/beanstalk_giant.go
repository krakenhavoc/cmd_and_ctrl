package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const beanstalkGiantOracleID = "22c75ad5-237c-4dbd-bbc4-3058a3db408f"

// Beanstalk Giant // Fertile Footsteps — Creature — Giant {6}{G}, */* //
// Sorcery — Adventure {2}{G} (Duskmourn Commander, #5368):
//
//	Beanstalk Giant: "Beanstalk Giant's power and toughness are each equal to
//	 the number of lands you control."
//	Fertile Footsteps: "Search your library for a basic land card, put it onto
//	 the battlefield, then shuffle."
//
// The creature face is Ashaya's characteristic-defining ability over lands
// (landsYouControlPT, layer 7a). The adventure half is a second spec under the
// "#1" key: the search is the shared basic-land fetch with the land entering
// untapped, and the adventure's exile-then-cast-the-creature is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     beanstalkGiantOracleID,
		Name:         "Beanstalk Giant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Beanstalk Giant's power and toughness are each equal to the number of lands you control."),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     beanstalkGiantOracleID + "#1",
		Name:         "Fertile Footsteps",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2B15SearchBasicLand("Fertile Footsteps — a basic land card, onto the battlefield", false)(ctx.Game, item)
		},
	})
}
