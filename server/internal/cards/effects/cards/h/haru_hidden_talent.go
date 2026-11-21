package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haru, Hidden Talent — Legendary Creature — Human Peasant Ally {1}{G},
// 1/1 (Avatar: The Last Airbender, #4459):
//
//	"Whenever another Ally you control enters, earthbend 1."
//
// Avatar Enthusiasts' trigger condition — any other permanent with the
// Ally type entering under Haru's controller's control (a token, a
// changeling), read through the layers, the type a word of the card's
// text (CR 612.2) — with the earthbend keyword action as its effect, its
// own target land chosen as the trigger goes on the stack. Two Allies
// entering together are two triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7488ce54-37c6-475c-b327-59f52f2b1424",
		Name:         "Haru, Hidden Talent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Haru, Hidden Talent — earthbend 1", effects.AvB14Earthbend(effects.EarthbendCount(1))), effects.EarthbendTargets()),
		},
	})
}
