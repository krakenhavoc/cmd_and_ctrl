package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weathered Bodyguards — Creature — Human Soldier {5}{W}, 2/5:
//
//	"As long as this creature is untapped, all combat damage that would be
//	 dealt to you by unblocked creatures is dealt to this creature instead.
//	 Morph {3}{W}"
//
// ADR 0108 §9 decision 4 (#1905): Veteran Bodyguard's static redirection,
// combat damage only (the ruling: not noncombat damage, not a blocked
// trampler's).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "86d0ba9d-6972-4cbc-871d-3fb413565c45",
		Name:             "Weathered Bodyguards",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{W}")},
		Replacements: []game.ReplacementEffect{
			staticRedirection("Weathered Bodyguards — combat damage to you from unblocked creatures is dealt to it instead",
				redirectWhere{applies: untappedAnd(damageToYouFromAnUnblockedCreature(true)), to: toThisPermanent}),
		},
	})
}
