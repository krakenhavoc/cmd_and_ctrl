package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abyssal Horror — Creature — Horror {4}{B}{B}, 2/2 (Urza's Saga, #5624):
//
//	"Flying
//	 When this creature enters, target player discards two cards."
//
// Thunderscape Battlemage's kicked discard without the kicker
// (targetPlayerDiscards): the target is chosen as the trigger goes on the
// stack (CR 603.3d), any player, its controller included, and that player
// picks the two cards from their own hand (CR 701.9a). A player with one
// card discards it. Flying is printed card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b60422a-765b-4168-a37b-68109103385a",
		Name:            "Abyssal Horror",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Abyssal Horror — target player discards two cards", effects.TargetPlayerDiscards(2)),
				effects.TargetPlayer("target player")),
		},
	})
}
