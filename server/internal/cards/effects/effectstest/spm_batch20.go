package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Spm20Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: toughness})
}

func Spm20Counters(g *game.Game, id uuid.UUID) int {
	return FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

// spm20Combat attacks with `attacker` into opp, blocked by `blockers`,
// answers any damage assignment prompt with `toBlockers` to each blocker
// and `toPlayer` to the player, and runs the combat damage steps
// (`firstStrike` adds the first-strike step).
func Spm20Combat(t *testing.T, g *game.Game, attacker uuid.UUID, opp *game.Player, blockers []uuid.UUID, toBlockers, toPlayer int, firstStrike bool) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatal(err)
		}
	}
	steps := []game.Step{game.StepCombatDamage}
	if firstStrike {
		steps = []game.Step{game.StepFirstStrikeDamage, game.StepCombatDamage}
	}
	for _, step := range steps {
		AdvanceTo(t, g, step)
		for i := 0; i < 4; i++ {
			var c *game.PendingChoice
			for _, p := range g.PendingChoices {
				if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
					c = p
				}
			}
			if c == nil {
				break
			}
			var entries []game.DamageAssignmentEntry
			for _, b := range blockers {
				entries = append(entries, game.DamageAssignmentEntry{BlockerID: b, Amount: toBlockers})
			}
			if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, toPlayer); err != nil {
				t.Fatalf("assign: %v", err)
			}
		}
	}
}
