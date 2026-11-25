package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// God-Eternal Rhonas — Legendary Creature — Zombie God {3}{G}{G}, 5/5 (War of the Spark, #5520):
//
//	"Deathtouch
//	 When God-Eternal Rhonas enters, double the power of each other creature
//	 you control until end of turn. Those creatures gain vigilance until end
//	 of turn.
//	 When God-Eternal Rhonas dies or is put into exile from the battlefield,
//	 you may put it into its owner's library third from the top."
//
// Thrakkus's doubling over the other creatures you control as the trigger
// resolves, every power read before any is applied (CR 701.10b), and the same
// set gains vigilance. The return is God-Eternal Oketra's: one ability, two
// conditions, a "may", the card slid under the top two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc4f2cdb-966c-4576-8b37-b3d138f5710e",
		Name:            "God-Eternal Rhonas",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("God-Eternal Rhonas — double the power of each other creature you control; those creatures gain vigilance",
				effects.S82OldB171RhonasDoubles),
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B22SelfDiedOrWasExiledFromBattlefield(ev, source)
			}, "God-Eternal Rhonas — put it into its owner's library third from the top", effects.TuckSelfThirdFromTop), "God-Eternal Rhonas — put it into its owner's library third from the top?"),
		},
	})
}
