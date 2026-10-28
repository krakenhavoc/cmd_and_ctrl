package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_damage_keyed_restore_test.go — #4362: a prevention follow-up that
// asks a question in the combat damage step (#2040) is a restore point,
// and the restored answer finishes the step. Khârn the Betrayer blocks:
// the attacker's combat damage to it is prevented and its controller is
// asked which opponent gains control of it, before the state-based
// actions (CR 615.5, CR 510.2).
func TestKharnsCombatDamageQuestionIsARestorePoint(t *testing.T) {
	build := func() (*game.Game, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kharn := kharnPush(g, opp.ID, opp.ID)
		attacker := pr7Creature(g, me.ID, "Attacker", 2, "R")
		pr7aBlockedCombat(t, g, opp.ID, attacker, kharn)
		advanceTo(t, g, game.StepCombatDamage)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, opp.ID) == nil {
			t.Fatal("Khârn asked nothing in the combat damage step")
		}
		return g, kharn, attacker
	}
	control, ck, ca := build()
	kharnAnswer(t, control, control.Seats[1].ID, control.Seats[2].ID)
	wantController := kharnController(t, control, ck)
	wantAttackerGone := findBattlefieldCardForTest(control, ca) == nil

	g, kharn, attacker := build()
	r := restoreRoundTrip(t, g, true)
	if findBattlefieldCardForTest(r, attacker) == nil {
		t.Fatal("the restored table swept the attacker before the question was answered")
	}
	kharnAnswer(t, r, r.Seats[1].ID, r.Seats[2].ID)
	if got := kharnController(t, r, kharn); got != r.Seats[2].ID || wantController != control.Seats[2].ID {
		t.Errorf("Khârn's controller %s on the restored table (control: %s), want the chosen opponent", got, wantController)
	}
	if gone := findBattlefieldCardForTest(r, attacker) == nil; gone != wantAttackerGone || !gone {
		t.Errorf("the attacker gone %v on the restored table, %v on the control: Khârn's 5 damage should kill it", gone, wantAttackerGone)
	}
}
