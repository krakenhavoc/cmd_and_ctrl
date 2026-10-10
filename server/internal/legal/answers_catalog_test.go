package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestCombatFlagsKeepEveryCatalogVerdict is ADR 0142's acceptance test
// for the combat tier: over every catalog activated row (own and
// granted) that is not targeted, the combat_interacts and
// combat_defender_only flags answersOf gives are #2871's, except where
// owner answer 2 moved a first strike, double strike or deathtouch
// grant out of interacts and into the combat tier.
// effects.TestAnswersOfKeepsEveryInteractsVerdict holds interacts.
func TestCombatFlagsKeepEveryCatalogVerdict(t *testing.T) {
	rows := 0
	var moved []string
	check := func(name string, a effects.ActivatedAbility) {
		if a.Targets != nil {
			return
		}
		ab := game.ActivatedAbilityShape{Label: a.Label, Cost: a.Cost, Modes: a.Modes, SorcerySpeed: a.SorcerySpeed, Purpose: a.Purpose}
		answers, declared := legal.AnswersOf(ab)
		if declared {
			return
		}
		rows++
		interacts, combat, defender := legal.AnswerFlags(answers)
		kind := legal.CombatKindForTest(ab)
		wasCombat := !interacts && kind != 0
		wasDefender := wasCombat && kind == 2
		if combat == wasCombat && defender == wasDefender {
			return
		}
		lower := strings.ToLower(a.Label)
		q2 := strings.Contains(lower, "first strike") || strings.Contains(lower, "double strike") || strings.Contains(lower, "deathtouch")
		if q2 && combat && !defender && !interacts {
			moved = append(moved, name+": "+a.Label)
			return
		}
		t.Errorf("%s %q: combat=%v defender_only=%v, #2871 said combat=%v defender_only=%v (%s)", name, a.Label, combat, defender, wasCombat, wasDefender, answers)
	}
	for _, spec := range effects.All() {
		for _, a := range spec.Activated {
			check(spec.Name, a)
		}
		for _, gr := range spec.Grants {
			for _, a := range gr.Activated {
				check(spec.Name+" (grant "+gr.Key+")", a)
			}
		}
	}
	if rows == 0 {
		t.Fatal("no catalog rows read: the test is reading the wrong registry")
	}
	t.Logf("%d untargeted rows; %d moved into the combat tier by owner answer 2:\n%s", rows, len(moved), strings.Join(moved, "\n"))
}
