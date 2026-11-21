package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Petrifying Meddler — Creature — Eldrazi {4}{U}, 4/5 (Modern Horizons
// 3):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, tap up to one target creature and put a
//	 stun counter on it. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)
//	 Reach"
//
// World Breaker's cast trigger with an "up to one" target (Ice Flan's tap-and-stun, finB1TapAndStun). It resolves
// above the spell, so it taps and stuns even if the Meddler is countered.
// The stun counter goes on whether or not the creature was already tapped.
// With no target the trigger still goes on the stack and does nothing.
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Petrifying Meddler — tap up to one target creature and put a stun counter on it", effects.FinB1TapAndStun)
	cast.Targets = effects.TargetCreature("up to one target creature").WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:        "bd9b3009-f389-4831-b250-80c4bb8f25e5",
		Name:            "Petrifying Meddler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "reach"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}
