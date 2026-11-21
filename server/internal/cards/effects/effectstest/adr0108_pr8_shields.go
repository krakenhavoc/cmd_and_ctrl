package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_shields_test.go — ADR 0108 PR 8, owner decision 2 (#1906):
// the scoped "prevent the next N damage" and "prevent all combat damage"
// shields whose CR 615.5 additional effect does something with what they
// prevented. Each test proves the shield, the follow-up's amount, and
// CR 615.12: damage that can't be prevented gets through, makes the
// follow-up do nothing, and leaves a charge whole.

const (
	Pr8Temper         = "9cefff18-96b1-4090-93cc-1ca981a58fbd"
	Pr8SacredBoon     = "32a8d49b-cdfb-4944-bfb5-afd704cc3658"
	Pr8ScarsOfVeteran = "5f0ad09f-a9ee-42df-b32d-9d842ce5dd00"
	Pr8AcolytesReward = "944689b2-f9f7-45c3-952c-54287b9824d5"
	Pr8VengefulArchon = "c3ef4311-c7f1-45ae-8bd5-9c05bdf2ae88"
	Pr8Inkshield      = "da93264e-4e04-401a-9809-e4e1056cf604"
)

// pr8CastX casts a catalogued spell from the active player's hand with X.
func Pr8CastX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c := game.NewCard(name, active.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	active.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
}
