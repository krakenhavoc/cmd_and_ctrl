package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverquill, the Disputant — Legendary Creature — Elder Dragon
// {2}{W}{B}, 4/4:
//
//	"Flying, vigilance
//	 Each instant and sorcery spell you cast has casualty 1. (As you
//	 cast that spell, you may sacrifice a creature with power 1 or
//	 greater. When you do, copy the spell and you may choose new targets
//	 for the copy.)"
//
// The grant is a GrantedOptionalCost (game/granted_optional_cost.go):
// while Silverquill is on the battlefield under your control, every
// instant or sorcery you cast offers "Casualty 1" beside its own
// optional costs. The sacrifice is paid as the spell is cast, and the
// copy is the spell's own casualty trigger (game/casualty.go), read off
// what was paid, so Silverquill can be the creature sacrificed and the
// spell is still copied.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:        "66deb3f3-0314-427e-9cb8-6befa365ae94",
		Name:            "Silverquill, the Disputant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			SpellsYouCastHave("Each instant and sorcery spell you cast has casualty 1.",
				Or(Instant(), Sorcery()), Casualty(1)),
		},
	})
}
