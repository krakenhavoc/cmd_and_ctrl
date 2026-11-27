package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pangosaur — Creature — Dinosaur {2}{G}{G}, 6/6 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"Whenever a player plays a land, return this creature to its owner's hand."
//
// Horn of Greed's land-PLAY watch (a fetched or ramped land is not played, CR 305.4),
// open to every player and to the Pangosaur's own controller, over Fleeting Effigy's
// return-to-hand: it comes back only while it is still the permanent the ability
// came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b350f23-9150-40c7-8ac7-86adfdb8ba52",
		Name:         "Pangosaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20LandPlayed(ev, g)
			}, "Pangosaur — return this creature to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
		},
	})
}
