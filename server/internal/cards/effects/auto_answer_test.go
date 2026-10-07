package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_answer_test.go — ADR 0127 (#1961) with the real cards it was
// written for: Rhystic Study's tax and draw, Consecrated Sphinx's "you
// may draw two", Esper Sentinel's changing cost. The engine-level cases
// are in internal/game/auto_answer_test.go.

// rhysticTaxOpen casts a Bolt from seat 0 into seat 1's Rhystic Study
// and resolves the trigger, leaving the caster's tax prompt open.
func rhysticTaxOpen(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	tax := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if tax == nil || tax.Chooser != g.Seats[0].ID {
		t.Fatalf("no tax prompt for the caster: %+v", tax)
	}
	return tax
}

// autoAnswerAll answers every prompt the table's standing answers
// cover, the way the room does after a commit, and returns how many.
func autoAnswerAll(t *testing.T, g *game.Game) int {
	t.Helper()
	n := 0
	for ; n < 64; n++ {
		id, _, ok := g.NextAutoAnswer()
		if !ok {
			return n
		}
		if _, err := g.AutoAnswer(id); err != nil {
			t.Fatalf("AutoAnswer: %v", err)
		}
	}
	t.Fatal("more than 64 automatic answers in a row")
	return n
}

func TestAutoAnswerRhysticStudyTaxAndDrawAreTwoKeys(t *testing.T) {
	g := newCatalogGame(t)
	tax := rhysticTaxOpen(t, g)
	if !strings.HasSuffix(tax.AutoAnswerKey, "|triggered|Rhystic Study — draw unless caster pays {1}|pay_unless#1") {
		t.Fatalf("tax key = %q, want the row's", tax.AutoAnswerKey)
	}
	if tax.AutoAnswerCard != "Rhystic Study" || tax.AutoAnswerPrompt != "Rhystic Study — pay {1}?" {
		t.Errorf("display copies = %q / %q", tax.AutoAnswerCard, tax.AutoAnswerPrompt)
	}
	answerPayUnless(t, g, g.Seats[0].ID, false)
	draw := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if draw == nil {
		t.Fatal("no draw prompt after the decline")
	}
	if want := tax.AutoAnswerKey + ">confirm#1"; draw.AutoAnswerKey != want {
		t.Errorf("draw key = %q, want %q", draw.AutoAnswerKey, want)
	}
}

func TestAutoAnswerRhysticStudyNeverPayAndAlwaysDraw(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	handBefore := owner.Hand.Size()
	tax := rhysticTaxOpen(t, g)
	if err := g.SetAutoAnswers(caster.ID, map[string]game.AutoAnswer{tax.AutoAnswerKey: game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	if err := g.SetAutoAnswers(owner.ID, map[string]game.AutoAnswer{tax.AutoAnswerKey + ">confirm#1": game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if n := autoAnswerAll(t, g); n != 2 {
		t.Fatalf("answered %d prompts automatically, want 2 (the tax, then the draw)", n)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Study's controller drew %d, want 1", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left open: %d", len(g.PendingChoices))
	}
}

func TestAutoAnswerRhysticStudyAlwaysPayTapsALand(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	handBefore := owner.Hand.Size()
	tax := rhysticTaxOpen(t, g)
	land := pushLandFor(g, caster.ID, "Island", "Basic Land — Island")
	if err := g.SetAutoAnswers(caster.ID, map[string]game.AutoAnswer{tax.AutoAnswerKey: game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if n := autoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d, want 1", n)
	}
	if c, ok := g.LookupCardForEffect(land); !ok || !c.Tapped {
		t.Error("Always pay did not tap the Island")
	}
	if owner.Hand.Size() != handBefore || latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a paid tax still offered the draw")
	}
}

func TestAutoAnswerRhysticStudyAlwaysPayWithNothingAsks(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	tax := rhysticTaxOpen(t, g)
	caster.ManaPool = nil
	if err := g.SetAutoAnswers(caster.ID, map[string]game.AutoAnswer{tax.AutoAnswerKey: game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if n := autoAnswerAll(t, g); n != 0 {
		t.Fatalf("answered %d with nothing to pay with, want 0", n)
	}
	if tax.AskedByHand != game.AskedByHandNoMana {
		t.Fatalf("AskedByHand = %q, want %q", tax.AskedByHand, game.AskedByHandNoMana)
	}
	// A land arriving now does not answer it under the player.
	pushLandFor(g, caster.ID, "Island", "Basic Land — Island")
	if n := autoAnswerAll(t, g); n != 0 {
		t.Errorf("a land drop answered a prompt asked by hand")
	}
}

func TestAutoAnswerConsecratedSphinxAlwaysBuildsTheTrigger(t *testing.T) {
	g := newCatalogGame(t)
	drawer, owner := g.Seats[0], g.Seats[1]
	sphinxID := pushDiesCreatureForTest(g, owner.ID, "Consecrated Sphinx", consecratedSphinxOracle,
		"Creature — Sphinx", 4, 6)
	advanceToMain(t, g)
	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatal(err)
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt)
	if ask == nil {
		t.Fatal("no Sphinx prompt")
	}
	if !strings.HasSuffix(ask.AutoAnswerKey, "|triggered|Consecrated Sphinx — draw two cards") {
		t.Fatalf("Sphinx key = %q, want its row", ask.AutoAnswerKey)
	}
	if err := g.SetAutoAnswers(owner.ID, map[string]game.AutoAnswer{ask.AutoAnswerKey: game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if n := autoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d, want 1", n)
	}
	if triggerOnStack(g, sphinxID) == nil {
		t.Error("Always did not put the Sphinx's trigger on the stack")
	}
}

// Esper Sentinel builds its question from its power. The key must not
// follow: a rule set at power 2 still answers at power 3.
func TestAutoAnswerEsperSentinelKeyIgnoresThePower(t *testing.T) {
	keyAt := func(power int) (string, string) {
		g := newCatalogGame(t)
		pushDiesCreatureForTest(g, g.Seats[1].ID, "Esper Sentinel", esperSentinelOracle,
			"Artifact Creature — Human Soldier", power, 1)
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
		passPriorityAroundTable(t, g)
		tax := latestChoiceOfKind(g, game.PendingChoicePayUnless)
		if tax == nil {
			t.Fatalf("no Sentinel tax at power %d", power)
		}
		return tax.AutoAnswerKey, tax.PayCost
	}
	k2, c2 := keyAt(2)
	k3, c3 := keyAt(3)
	if k2 == "" || k2 != k3 {
		t.Errorf("keys at power 2 and 3 = %q, %q; want one non-empty key", k2, k3)
	}
	if c2 == c3 {
		t.Errorf("the cost did not change with the power (%s), so this test proves nothing", c2)
	}
}

// Two Consecrated Sphinxes, both set to Always, would draw both
// libraries out with nobody clicking. An automatic answer is not a
// decision (ADR 0127 §7), so the loop breaker's count runs on and its
// notice stops the automatic answers.
func TestAutoAnswerTwoSphinxesStopAtTheLoopBreaker(t *testing.T) {
	g := newCatalogGame(t)
	g.LoopThreshold = 3
	a, b := g.Seats[0], g.Seats[1]
	for _, p := range []*game.Player{a, b} {
		for i := 0; i < 200; i++ {
			p.Library.PushTop(game.NewCard("basic-filler", uuid.Nil))
		}
	}
	pushDiesCreatureForTest(g, a.ID, "Consecrated Sphinx", consecratedSphinxOracle, "Creature — Sphinx", 4, 6)
	pushDiesCreatureForTest(g, b.ID, "Consecrated Sphinx", consecratedSphinxOracle, "Creature — Sphinx", 4, 6)
	advanceToMain(t, g)
	if err := g.DrawCard(a.ID); err != nil {
		t.Fatal(err)
	}
	key := latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt).AutoAnswerKey
	for _, p := range []*game.Player{a, b} {
		if err := g.SetAutoAnswers(p.ID, map[string]game.AutoAnswer{key: game.AutoAnswerAlways}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 500 && g.LoopNotice == nil; i++ {
		if autoAnswerAll(t, g) > 0 {
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.LoopNotice == nil {
		t.Fatalf("no loop notice; libraries at %d and %d", a.Library.Size(), b.Library.Size())
	}
	if a.Library.Size() == 0 || b.Library.Size() == 0 {
		t.Fatal("a library ran out before the breaker fired")
	}
	if autoAnswerAll(t, g) != 0 {
		t.Error("prompts were answered automatically under the loop notice")
	}
}
