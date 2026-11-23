package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Meanderer — Creature — Bird Fish Illusion {3}{G}{U}{U}, 4/4
// (Reality Fracture, tracker #2795):
//
//	"Flying
//	 This creature has vigilance as long as you control a Jace
//	 planeswalker.
//	 When this creature enters, it fights up to one target creature an
//	 opponent controls."
//
// A Jace token (Empower Jace) is a Jace planeswalker, so it turns the
// vigilance on. The fight is the "up to one" shape (Mawloc): the trigger
// always goes on the stack and may take no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18d1765b-3b60-4dda-93a6-d02a4e53f185",
		Name:            "Mind Meanderer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.HasSubtype("Jace"))), "vigilance"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Mind Meanderer — it fights up to one target creature an opponent controls",
			Targets:   effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
			Effect:    effects.ItFightsTheFirstTarget,
		}},
	})
}
