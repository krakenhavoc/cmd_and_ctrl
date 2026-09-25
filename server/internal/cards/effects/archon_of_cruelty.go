package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Cruelty — Creature — Archon {6}{B}{B}, 6/6 (EDHREC rank
// 1102):
//
//	"Flying
//	 Whenever this creature enters or attacks, target opponent
//	 sacrifices a creature or planeswalker of their choice, discards
//	 a card, and loses 3 life. You draw a card and gain 3 life."
//
// The reanimator target. "Enters or attacks" is one ability watching
// two event kinds (Sun Titan's shape), targeting an opponent through
// the ordinary pick_target prompt when the trigger goes on the stack.
// The body is b09ArchonOfCrueltyTrigger: the opponent's sacrifice and
// discard are their own choices, made through the sacrifice-choice
// and pending-discard prompts, and the life swing runs at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa1a6646-c1e6-4bff-9092-43ee3e137914",
		Name:            "Archon of Cruelty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Archon of Cruelty — target opponent sacrifices, discards and loses 3; you draw and gain 3",
			Effect:  b09ArchonOfCrueltyTrigger,
		}},
	})
}
