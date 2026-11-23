package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harrier Strix — Creature — Bird {U}, 1/1 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 When this creature enters, tap target permanent.
//	 {2}{U}: Draw a card, then discard a card."
//
// The enters trigger is Skateboard's: the target is any permanent, an
// opponent's or your own, chosen as the trigger goes on the stack (CR
// 603.3d) and tapped if it is still legal (tapTheFirstLegalTarget). The
// loot is Agna Qel'a's without the tap, so it can be activated any
// number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb154ebb-d02f-405a-a557-3d5ec7648c42",
		Name:            "Harrier Strix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Harrier Strix — tap target permanent", effects.TapTheFirstLegalTarget),
				effects.TargetPermanent("target permanent")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
		}},
	})
}
