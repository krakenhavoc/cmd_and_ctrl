package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_reselect_test.go — the enumerator half of #1329.
//
// The CR 508.7 prompt is an option_pick, a kind the enumerator already
// answers; what is new is its option list — seats AND a permanent, read
// back by the option's own subject rather than its index. So the claim
// here is the #544 one for that list: every option is offered, every
// offered move dispatches, and the FIRST ("keep attacking …") is the
// always-legal one, so a bot owing it can always move and, with no
// opinion, leaves somebody else's combat alone.
func TestReselectAttackPromptIsEnumeratedAndEveryAnswerDispatches(t *testing.T) {
	g := newTable(t)
	active, defender, chooser := g.Seats[0], g.Seats[1], g.Seats[1]
	attacker := battlefieldCard(g, active, creature("Charger", "{2}{R}", 3, 3))
	battlefieldCard(g, g.Seats[3], game.Card{
		Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 4},
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
			Chooser: chooser.ID, Attacker: attacker, Question: "Misleading Signpost",
		})
	})
	if id == uuid.Nil {
		t.Fatalf("no prompt was queued")
	}

	moves := legal.EnumerateFor(g, chooser.ID)
	var choiceMoves []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindChoice {
			choiceMoves = append(choiceMoves, m)
		}
	}
	// keep + seat 2 + seat 3 + the planeswalker.
	if len(choiceMoves) != 4 {
		t.Fatalf("enumerated %d answers, want 4: %v", len(choiceMoves), labels(choiceMoves))
	}
	if !choiceMoves[0].AlwaysLegal {
		t.Errorf("the first answer (%q) is not marked always-legal", choiceMoves[0].Label)
	}
	if !hasLabel(choiceMoves, "Misleading Signpost: Keep attacking") {
		t.Errorf("no \"keep\" answer: %v", labels(choiceMoves))
	}
	dispatchAll(t, g, chooser.ID, choiceMoves)
}
