package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindsparker — Creature — Elemental {1}{R}{R}, 3/2 (Foundations, #4465):
//
//	"First strike
//	 Whenever an opponent casts a white or blue instant or sorcery spell,
//	 this creature deals 2 damage to that player."
//
// Gleeful Arsonist's cast trigger with a fixed 2 and a narrower spell: an
// instant or sorcery that is white, blue or both, read off the spell on
// the stack. It triggers on the cast, so it resolves first and happens
// even if the spell is countered. "That player" is the caster, read off
// the item's carried event. A Mindsparker that has left the battlefield
// in response still deals the damage, named as the departed object
// (DealDamage.SourceObject). First strike rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54866c25-b332-46c8-a2c5-1a71b702ddc0",
		Name:            "Mindsparker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.And(effects.Or(effects.OfColor("W"), effects.OfColor("U")), effects.Or(effects.Instant(), effects.Sorcery()))),
				"Mindsparker — 2 damage to that player",
				effects.SourceDealsTwoToThatPlayer),
		},
	})
}
