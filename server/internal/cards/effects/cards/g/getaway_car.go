package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Getaway Car — {3} Artifact — Vehicle, 4/3:
//
//	"Haste
//	 Whenever this Vehicle attacks or blocks, return up to one target
//	 creature that crewed it this turn to its owner's hand.
//	 Crew 1"
//
// One printed ability with two trigger conditions is one TriggeredAbility
// watching two kinds (Smuggler's Copter's shape): the attack declaration
// or the block declaration (once however many attackers it blocks, CR
// 509.3a). The target clause is CrewedThis with "up to one": the trigger
// still goes on the stack with no target chosen, and a crewer that left
// or came back as a new object is not offered (CR 400.7).
//
// No simplification.
func init() {
	t := effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
		},
		"Getaway Car — return up to one target creature that crewed it to its owner's hand",
		effects.BounceEachLegalTarget)
	t.Targets = effects.CrewedThis(effects.TargetCreature("up to one target creature that crewed it this turn").WithCount(0, 1))

	effects.Register(effects.Spec{
		OracleID:        "954b30b2-b88c-4558-b06c-4518b49d7b20",
		Name:            "Getaway Car",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Getaway Car"),
		}},
		Triggered: []game.TriggeredAbility{t},
	})
}
