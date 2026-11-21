package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avian Oddity — Creature — Bird {3}{U}, 2/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"Flying
//	 Cycling {2}{U} ({2}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, put a flying counter on target creature you
//	 control."
//
// The cycling trigger fires from the graveyard (CR 702.29c) and goes on
// the stack above the cycling ability, so the counter lands before the
// draw. The counter gives the keyword for as long as it stays (CR
// 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d6a03a52-7e85-46de-920c-93a3f6479c4c",
		Name:            "Avian Oddity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}{U}")},
		Triggered:       []game.TriggeredAbility{effects.IkoCycledCounterOnYourCreature("Avian Oddity", game.CounterFlying, "flying")},
	})
}
