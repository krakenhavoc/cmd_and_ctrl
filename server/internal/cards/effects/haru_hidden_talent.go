package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7488ce54-37c6-475c-b327-59f52f2b1424",
		Name:         "Haru, Hidden Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Haru, Hidden Talent — earthbend 1", avB14Earthbend(EarthbendCount(1))), EarthbendTargets()),
		},
	})
}
