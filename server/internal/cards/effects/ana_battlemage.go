package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ana Battlemage — Creature — Human Wizard {2}{G}, 2/2:
//
//	"Kicker {2}{U} and/or {1}{B} (You may pay an additional {2}{U}
//	 and/or {1}{B} as you cast this spell.)
//	 When this creature enters, if it was kicked with its {2}{U}
//	 kicker, target player discards three cards.
//	 When this creature enters, if it was kicked with its {1}{B}
//	 kicker, tap target untapped creature and that creature deals damage
//	 equal to its power to its controller."
//
// Thornscape Battlemage's shape (#2153). The {1}{B} half is two
// instructions on one target: the tap, then the damage, dealt by the
// tapped creature itself (so its lifelink and deathtouch count) to its
// own controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c37537ae-616a-41e4-890c-f4b9475dbf31",
		Name:          "Ana Battlemage",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{2}{U}", "{1}{B}"),
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, AllOf(Self, ThisKickedWith("{2}{U}")),
				"Ana Battlemage — kicked with {2}{U}, target player discards three cards",
				targetPlayerDiscards(3)), TargetPlayer("target player")),
			Targeting(On(game.EventETB, AllOf(Self, ThisKickedWith("{1}{B}")),
				"Ana Battlemage — kicked with {1}{B}, tap target untapped creature, it deals damage equal to its power to its controller",
				tapTargetThenItDealsItsPowerToItsController),
				TargetCreature("target untapped creature", untappedCreature())),
		},
	})
}
