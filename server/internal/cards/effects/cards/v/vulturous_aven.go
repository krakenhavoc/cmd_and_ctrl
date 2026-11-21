package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulturous Aven — Creature — Bird Shaman {3}{B}, 2/3 (Dragons of Tarkir,
// #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, you draw two cards and you lose 2 life."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff
// is a reflexive trigger raised only when a creature was really
// sacrificed, and a targeted payoff chooses its target then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83c0c648-f000-470d-bcc3-9ec81257d0cd",
		Name:            "Vulturous Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Vulturous Aven", vulturousAvenDrawBody),
		},
	})
}

// vulturousAvenDrawBody is "When this creature exploits a creature, you draw two cards and you lose 2 life." An on-disk key:
// never renamed, never reused.
var vulturousAvenDrawBody = game.ReflexiveBody("vulturous-aven/draw-and-lose", effects.SimpleBody(effects.Do(effects.DrawCards{N: 2}, effects.GainLife{Amount: -2})), nil)
