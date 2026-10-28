package game

import (
	"testing"

	"github.com/google/uuid"
)

// assist_park_order_test.go — #4404, case 4. A parked cast that already
// carries an assist answer is never written (parkedAnnouncement.snapshot),
// and docs/snapshots.md closes the case because no park can be reached
// with one: castSpellLocked asks the opponent-target and commander-cost
// questions before it parks for assist (ADR 0150 §4), and the assisted
// re-cast carries both answers (opponentPicks, commanderAnswers). This
// pins that order. If it ever fails, a parked cast can carry assistPaid,
// and that answer has to be written (with a key of its own) before the
// case can stay closed.

// A commander returned to hand for an assisted spell's alternative cost
// (CR 903.9b, asked before paying) is asked first. The assist question
// comes after it, and its answer casts the spell with no park at all.
func TestAnAssistAnswerIsNeverCarriedIntoAPark(t *testing.T) {
	const oracle = "test-4404-assisted-daze"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "daze", Label: "Return a creature you control", ManaCost: "{2}", ReturnToHand: creatureCostSpec(),
	}))
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 3)
	spell := NewCard("Assisted Daze", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.ManaCost = "{5}{U}"
	spell.Keywords = []string{KeywordAssist}
	me.Hand.PushTop(spell)
	cmd := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)

	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Strict: true, AutoTap: true, AssistPlayer: helper.ID, AlternativeCost: "daze", AltCostIDs: []uuid.UUID{cmd},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	c := parkedCostPrompt(t, g, me.ID, cmd)
	if c.costCommanderResume.announcement.cast.assistPaid != nil {
		t.Fatal("the commander's question was parked with an assist answer")
	}
	if err := g.ResolveOptionalReplacement(c.ID, me.ID, false); err != nil {
		t.Fatalf("answer the commander's question: %v", err)
	}
	a := assistPrompt(t, g)
	for _, p := range g.PendingChoices {
		if p.costCommanderResume != nil || p.opponentTargetResume != nil {
			t.Fatal("a parked announcement is open beside the assist question")
		}
	}
	if err := g.ResolvePayAmount(a.ID, helper.ID, 2); err != nil {
		t.Fatalf("answer the assist question: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open after the assist answer, want none: the re-cast parked again", len(g.PendingChoices))
	}
	if g.StackMeta[spell.InstanceID] == nil {
		t.Fatal("the spell is not on the stack")
	}
	if !me.Hand.Contains(cmd) {
		t.Error("the commander was not returned to hand for the cost")
	}
}
