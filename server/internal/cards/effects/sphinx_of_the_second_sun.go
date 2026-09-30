package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx of the Second Sun — Creature — Sphinx {6}{U}{U}, 6/6:
//
//	"Flying
//	 At the beginning of each of your postcombat main phases, there is
//	 an additional beginning phase after this phase. (The beginning
//	 phase includes the untap, upkeep, and draw steps.)"
//
// An added beginning phase is untap, upkeep and draw (ADR 0059 sub-PR
// 2b, #753), so the controller untaps, "at the beginning of your
// upkeep" triggers again and they draw a card. It is not a new turn:
// summoning sickness, "until your next turn" effects and the per-turn
// tallies are keyed to the turn beginning, not to the untap step, so
// none of them change (the 2020-11-10 rulings). The trigger fires in
// EVERY postcombat main phase of the controller's turn, including one
// another effect added.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "516101be-be39-4d84-8fee-d8a79930dd0a",
		Name:            "Sphinx of the Second Sun",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourPostcombatMain("Sphinx of the Second Sun — an additional beginning phase", Do(AddPhases{
				Anchor: game.PhaseAnchor{Kind: game.AnchorThisPhase},
				Kinds:  []game.PhaseKind{game.PhaseKindBeginning},
			})),
		},
	})
}
