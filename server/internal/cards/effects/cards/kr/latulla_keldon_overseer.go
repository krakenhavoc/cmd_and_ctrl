package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Latulla, Keldon Overseer — Legendary Creature — Human Spellshaper {3}{R}{R},
// 3/3 (Prophecy, S82 OLD b670, #5617):
//
//	"{X}{R}, {T}, Discard two cards: Latulla deals X damage to any target."
//
// Pyre-Sledge Arsonist's tapped ping with X read from the activation; the two
// discards are an announce-time cost the activator picks from hand (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86b2879b-6c54-4d96-ab81-a3f2c87d42d9",
		Name:         "Latulla, Keldon Overseer",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{R}, {T}, Discard two cards: Latulla deals X damage to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{X}{R}"), effects.TapCost(), effects.DiscardN(2, "two cards")),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int { return ctx.X() }),
		}},
	})
}
