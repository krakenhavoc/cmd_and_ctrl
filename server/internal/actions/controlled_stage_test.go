package actions

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_stage_test.go — #4025, ADR 0166 amendment of 2026-10-19:
// what a controller may send as the player they control, per stage of
// the control (CR 723.5b). The allowed sets are written out here on
// purpose, independent of controlledStageVerbs, so widening the table
// means editing a test too; every other verb Dispatch knows must be
// refused.

func typeSet(ts ...Type) map[Type]bool {
	m := make(map[Type]bool, len(ts))
	for _, t := range ts {
		m[t] = true
	}
	return m
}

func TestControlledVerbAllowedPerStage(t *testing.T) {
	stages := []struct {
		name    string
		stage   game.ControlStage
		allowed map[Type]bool
	}{
		{"word of command, the forced play", game.ControlPlaying,
			typeSet(TypeCastSpell, TypeActivateManaAbility, TypeResolveChoice)},
		{"word of command, the forced spell resolving", game.ControlSpellResolving,
			typeSet(TypeResolveChoice)},
		{"opposition agent, the search", game.ControlSearching,
			typeSet(TypeResolveChoice)},
		{"emrakul, the controlled turn", game.ControlTurn,
			typeSet(TypePassPriority, TypeEndTurn, TypeCastSpell, TypeActivateAbility,
				TypeActivateLoyalty, TypeActivateManaAbility, TypeSpecialAction, TypeOutsidePlay,
				TypeDeclareAttacker, TypeDeclareAttackers, TypeResolveChoice, TypeDiscardSelection)},
		{"a stage that is waiting, not live", game.ControlAwaitingTurn, typeSet()},
		{"no live control", "", typeSet()},
		{"a stage nobody has defined", "from_the_future", typeSet()},
	}
	known := typeSet(Types()...)
	for _, s := range stages {
		t.Run(s.name, func(t *testing.T) {
			for _, ty := range Types() {
				if got := controlledVerbAllowed(s.stage, ty); got != s.allowed[ty] {
					t.Errorf("%s as a controlled player: allowed = %v, want %v", ty, got, s.allowed[ty])
				}
			}
			for ty := range s.allowed {
				if !known[ty] {
					t.Errorf("%s in the allowed set is not an action type", ty)
				}
			}
		})
	}
}

// No stage allows a sandbox verb as the controlled player (CR 723.5b).
func TestNoControlledStageAllowsASandboxVerb(t *testing.T) {
	for stage := range controlledStageVerbs {
		for ty := range sandboxVerbs {
			if controlledVerbAllowed(stage, ty) {
				t.Errorf("stage %q allows the sandbox verb %s", stage, ty)
			}
		}
	}
}
