package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// A table can be sitting in declare_blockers with nothing attacking
// when a deploy lands: the old binary walked that step in every combat.
// The restore point it wrote has no combatHadAttackers key (ADR 0044's
// 2026-10-12 amendment: an absent field falls back to the live board),
// and the new binary must read it, let the table leave the step, and
// walk on to the end of combat without stalling or entering a combat
// damage step the rules no longer give the combat (CR 508.8, #2899).
func TestRestorePointFromBeforeThisChangeRestingInDeclareBlockersMovesOn(t *testing.T) {
	for _, drive := range []string{"advance_step", "pass_priority"} {
		t.Run(drive, func(t *testing.T) {
			g := newActiveGame(t)
			attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
			advanceIntoStep(t, g, StepDeclareAttackers)
			if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			advanceIntoStep(t, g, StepDeclareBlockers)
			// What the old binary could hold here: the step entered
			// with nothing attacking. The attacker is out of combat.
			g.WithWriteLock(func() {
				findBattlefieldCard(g, attacker).AttackingTarget = uuid.Nil
			})

			raw, err := json.Marshal(g.CaptureSnapshot())
			if err != nil {
				t.Fatal(err)
			}
			var generic map[string]any
			if err := json.Unmarshal(raw, &generic); err != nil {
				t.Fatal(err)
			}
			if _, ok := generic["combatHadAttackers"]; !ok {
				t.Fatal("setup: the new snapshot does not carry combatHadAttackers")
			}
			delete(generic, "combatHadAttackers")
			old, err := json.Marshal(generic)
			if err != nil {
				t.Fatal(err)
			}
			var back GameSnapshot
			if err := json.Unmarshal(old, &back); err != nil {
				t.Fatal(err)
			}
			restored, err := back.Restore()
			if err != nil {
				t.Fatalf("Restore: %v", err)
			}
			if restored.Turn.Step != StepDeclareBlockers {
				t.Fatalf("restored on %s, want declare_blockers", restored.Turn.Step)
			}
			if restored.combatHadAttackers {
				t.Fatal("setup: the old file should not carry the record")
			}
			mark := eventHighWater(restored)
			for i := 0; i < 16 && restored.Turn.Step != StepEndCombat; i++ {
				if drive == "advance_step" {
					if _, err := restored.AdvanceStep(); err != nil {
						t.Fatalf("AdvanceStep: %v", err)
					}
					continue
				}
				if restored.Turn.PriorityHolder < 0 {
					t.Fatalf("nobody holds priority in %s", restored.Turn.Step)
				}
				if err := restored.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if restored.Turn.Step != StepEndCombat {
				t.Fatalf("stalled on %s, want end_combat", restored.Turn.Step)
			}
			if began := stepsBegunSince(restored, mark); stepsContain(began, StepCombatDamage) {
				t.Errorf("a combat damage step began after the restore (began: %v)", began)
			}
		})
	}
}
