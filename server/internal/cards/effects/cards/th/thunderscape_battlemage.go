package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderscape Battlemage — Creature — Human Wizard {2}{R}, 2/2:
//
//	"Kicker {1}{B} and/or {G} (You may pay an additional {1}{B} and/or
//	 {G} as you cast this spell.)
//	 When this creature enters, if it was kicked with its {1}{B}
//	 kicker, target player discards two cards.
//	 When this creature enters, if it was kicked with its {G} kicker,
//	 destroy target enchantment."
//
// Thornscape Battlemage's shape (#2153): each enters ability is linked
// to one kicker cost (CR 702.33f) and reads it off the permanent's
// record as an intervening if (CR 603.4). The discard is the discarding
// player's own choice of cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9aa66e7d-d5e3-4c48-9c68-b61ff2af0b1a",
		Name:          "Thunderscape Battlemage",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: effects.Kickers("{1}{B}", "{G}"),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{1}{B}")),
				"Thunderscape Battlemage — kicked with {1}{B}, target player discards two cards",
				effects.TargetPlayerDiscards(2)), effects.TargetPlayer("target player")),
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{G}")),
				"Thunderscape Battlemage — kicked with {G}, destroy target enchantment",
				effects.DestroyChosenPermanent), effects.TargetPermanent("target enchantment", effects.Enchantment())),
		},
	})
}
