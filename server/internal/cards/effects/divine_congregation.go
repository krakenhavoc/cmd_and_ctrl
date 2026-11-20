package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divine Congregation — Sorcery {3}{W} (Time Spiral, #5596):
//
//	"You gain 2 life for each creature target player controls.
//	 Suspend 5—{1}{W} (Rather than cast this card from your hand, you
//	 may pay {1}{W} and exile it with five time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)"
//
// Two life for each creature the targeted player controls, counted as the
// spell resolves. Any player may be the target, its caster included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2fc892b-3840-45a0-bb1c-0d2cf7d14f2f",
		Name:         "Divine Congregation",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		SpecialActions: []game.SpecialAction{
			Suspend(5, "{1}{W}"),
		},
		OnResolve: tspB1GainTwoPerCreatureTargetPlayerControls,
	})
}
