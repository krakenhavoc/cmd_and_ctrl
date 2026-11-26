package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindreaver — {U}{U} Creature — Human Wizard 2/1 (Born of the Gods):
//
//	"Heroic — Whenever you cast a spell that targets this creature, exile the top three
//	 cards of target player's library.
//	 {U}{U}, Sacrifice this creature: Counter target spell with the same name as a card
//	 exiled with this creature."
//
// The exile is linked to this object (CR 607.2a), as Circle of Confinement's: three cards
// as one exile, from the top. The counter's target reads that link when it is announced and
// again as it resolves (CR 608.2b), by name (game.TargetSpec.SameNameAsExiledWithSource);
// the link outlives the sacrifice that paid for it, and a Mindreaver that left and came
// back is a new object that has nothing exiled with it (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f606ab91-1f20-4351-acd3-6366b1895f72",
		Name:         "Mindreaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Heroic("Mindreaver — exile the top three cards of target player's library (heroic)",
				effects.ExileTopCardsOfTargetPlayerLinked(3)),
				effects.TargetPlayer("target player")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, Sacrifice this creature: Counter target spell with the same name as a card exiled with this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.SacrificeThis()),
			Targets: effects.SameNameAsCardExiledWithThis(effects.TargetSpell("target spell with the same name as a card exiled with this creature")),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
