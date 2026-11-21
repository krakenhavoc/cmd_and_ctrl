package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightscape Battlemage — Creature — Zombie Wizard {2}{B}, 2/2:
//
//	"Kicker {2}{U} and/or {2}{R} (You may pay an additional {2}{U}
//	 and/or {2}{R} as you cast this spell.)
//	 When this creature enters, if it was kicked with its {2}{U}
//	 kicker, return up to two target nonblack creatures to their
//	 owners' hands.
//	 When this creature enters, if it was kicked with its {2}{R}
//	 kicker, destroy target land."
//
// Thornscape Battlemage's shape (#2153). "Up to two" is a 0-2 target
// clause: with none chosen the trigger never goes on the stack (CR
// 603.3d), and each target that has left is skipped at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e96e68b4-cb32-4b75-a885-e1781f4b74ab",
		Name:          "Nightscape Battlemage",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: effects.Kickers("{2}{U}", "{2}{R}"),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{2}{U}")),
				"Nightscape Battlemage — kicked with {2}{U}, return up to two target nonblack creatures to their owners' hands",
				effects.BounceEachLegalTarget), effects.TargetCreature("up to two target nonblack creatures", effects.NonBlack()).WithCount(0, 2)),
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{2}{R}")),
				"Nightscape Battlemage — kicked with {2}{R}, destroy target land",
				effects.DestroyChosenPermanent), effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
