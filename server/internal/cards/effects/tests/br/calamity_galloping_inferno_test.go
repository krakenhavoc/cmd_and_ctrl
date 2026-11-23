package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The headline: two token copies enter tapped and attacking, the copied
// creature's attack trigger does not fire (CR 508.4), and both tokens are
// sacrificed at the beginning of the next end step.
func TestCalamityMakesTwoTappedAttackingCopiesSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	saddler := effectstest.PushProbeSaddler(g, me.ID)
	effectstest.CalamityProbeAttacks = 0

	choice := effectstest.CalamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled Calamity")
	}
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 {
		t.Errorf("the choice is %d..%d, want exactly one", choice.ChooseMin, choice.ChooseMax)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("first choice: %v", err)
	}
	effectstest.AnswerAttackSplit(t, g, me, 0)
	if got := len(effectstest.CalamityTokens(g, me.ID, "Calamity Probe Saddler")); got != 1 {
		t.Fatalf("after the first round there are %d copies, want 1", got)
	}

	choice = effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("the second round did not ask")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("second choice: %v", err)
	}
	effectstest.AnswerAttackSplit(t, g, me, 0)

	tokens := effectstest.CalamityTokens(g, me.ID, "Calamity Probe Saddler")
	if len(tokens) != 2 {
		t.Fatalf("%d token copies, want 2", len(tokens))
	}
	var ids []uuid.UUID
	for _, tok := range tokens {
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Errorf("token %s: tapped=%v attacking=%s, want tapped and attacking %s", tok.InstanceID, tok.Tapped, tok.AttackingTarget, opp.ID)
		}
		ids = append(ids, tok.InstanceID)
	}
	// The saddler tapped to pay; the probe trigger would have fired for a
	// token declared as an attacker.
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CalamityProbeAttacks != 0 {
		t.Errorf("the copies' attack trigger fired %d times; they were never declared as attackers (CR 508.4)", effectstest.CalamityProbeAttacks)
	}
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler itself left the battlefield")
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Errorf("token %s survived the beginning of the next end step", id)
		}
	}
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was sacrificed with its copies")
	}
}

// Each token's controller chooses what it attacks: the two copies can
// attack different players.
func TestCalamityCopiesCanAttackDifferentPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	saddler := effectstest.PushProbeSaddler(g, me.ID)

	choice := effectstest.CalamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerAttackSplit(t, g, me, 0) // the first copy keeps Calamity's defender
	choice = effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("the second round did not ask")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerAttackSplit(t, g, me, 1) // the second goes to the next player offered

	tokens := effectstest.CalamityTokens(g, me.ID, "Calamity Probe Saddler")
	if len(tokens) != 2 {
		t.Fatalf("%d token copies, want 2", len(tokens))
	}
	targets := map[uuid.UUID]bool{}
	for _, tok := range tokens {
		targets[tok.AttackingTarget] = true
	}
	if len(targets) != 2 {
		t.Errorf("the copies attack %v, want two different players", targets)
	}
	if targets[me.ID] || targets[uuid.Nil] {
		t.Errorf("a copy attacks %v, which is not an opponent", targets)
	}
}

// A legendary saddler is not offered, and a nonlegendary one is.
func TestCalamityDoesNotOfferALegendarySaddler(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	nonlegendary := effectstest.PushProbeSaddler(g, me.ID)
	legend := effectstest.PushDiesCreatureForTest(g, me.ID, "Legendary Saddler", "", "Legendary Creature — Test", 2, 2)
	bystander := effectstest.PushCrewerForTest(g, me.ID, "Bystander", 3)

	choice := effectstest.CalamityAttackPrompt(t, g, me, opp.ID, calamity, nonlegendary, legend)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != nonlegendary {
		t.Errorf("offered %v, want only the nonlegendary saddler %s (not the legend %s or the bystander %s)",
			choice.ChooseCards, nonlegendary, legend, bystander)
	}
}

// With no nonlegendary saddler there is nothing to copy: no prompt, no
// token, in either round.
func TestCalamityWithOnlyALegendarySaddlerDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	legend := effectstest.PushDiesCreatureForTest(g, me.ID, "Legendary Saddler", "", "Legendary Creature — Test", 2, 2)

	if choice := effectstest.CalamityAttackPrompt(t, g, me, opp.ID, calamity, legend); choice != nil {
		t.Fatalf("a prompt was queued with only a legendary saddler: %v", choice.ChooseCards)
	}
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("an attack-split prompt was queued")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].IsToken() {
			t.Errorf("token %s was created", g.Battlefield.Cards[i].Name)
		}
	}
}

// Saddlers that left the battlefield before the trigger resolves are not
// offered: with the only nonlegendary saddler gone, nothing happens.
func TestCalamityDoesNotOfferASaddlerThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	saddler := effectstest.PushProbeSaddler(g, me.ID)
	effectstest.ToMain(t, g)
	if err := effectstest.SaddleAbility(t, g, me.ID, calamity, saddler); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, calamity)
	g.WithWriteLock(func() {
		if err := (effects.BounceToHand{Target: saddler}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if choice := effectstest.LatestChooseCards(g, me.ID); choice != nil {
		t.Errorf("offered %v after the saddler left", choice.ChooseCards)
	}
}

// "Repeat this process once" may choose the same creature, and that makes
// two copies of it.
func TestCalamityMayChooseTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Lone Saddler", 3)

	for round := 1; round <= 2; round++ {
		choice := effectstest.LatestChooseCards(g, me.ID)
		if round == 1 {
			choice = effectstest.CalamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
		}
		if choice == nil {
			t.Fatalf("round %d did not ask", round)
		}
		if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
			t.Fatalf("round %d offered %v, want the one saddler", round, choice.ChooseCards)
		}
		if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		effectstest.AnswerAttackSplit(t, g, me, 0)
	}
	if got := len(effectstest.CalamityTokens(g, me.ID, "Lone Saddler")); got != 2 {
		t.Errorf("%d copies of the same creature, want 2", got)
	}
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("a third round was offered")
	}
}

// The probe really does count a declared attack, so the zero in the
// headline test means something.
func TestCalamityProbeCountsADeclaredAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	probe := effectstest.PushProbeSaddler(g, me.ID)
	effectstest.CalamityProbeAttacks = 0
	effectstest.ToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, probe)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CalamityProbeAttacks != 1 {
		t.Errorf("the probe counted %d attacks, want 1", effectstest.CalamityProbeAttacks)
	}
}

// Calamity only triggers while saddled.
func TestCalamityUnsaddledAttackDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := effectstest.PushCalamity(g, me.ID)
	effectstest.PushProbeSaddler(g, me.ID)
	effectstest.ToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, calamity)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("an unsaddled Calamity prompted")
	}
}
