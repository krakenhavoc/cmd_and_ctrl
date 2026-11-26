package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hateflayer — Creature — Elemental {5}{R}{R}, 5/5 (Eventide, S82 OLD batch
// 520, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 {2}{R}, {Q}: This creature deals damage equal to its power to any target.
//	 ({Q} is the untap symbol.)"
//
// Order of Whiteclay's {Q}: AbilityCost.UntapSelf (CR 107.6) needs the Hateflayer
// tapped and under its controller's control since their turn began (CR 302.6). The
// damage is its power as the ability resolves (by last-known information if it has
// left, CR 608.2h), and wither applies to it as it does to all damage the
// Hateflayer deals.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "681d07e6-234f-4544-a2ea-bfbefe6b0b36",
		Name:            "Hateflayer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {Q}: This creature deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{Mana: "{2}{R}", UntapSelf: true},
			Targets: effects.TargetAny(),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				if info, ok := ctx.SourcePermanent(); ok && info.Power > 0 {
					return info.Power
				}
				return 0
			}),
		}},
	})
}
