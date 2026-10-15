package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// calamity_galloping_inferno_test.go — #3548, ADR 0186 slice 6.

const (
	calamityOracle      = "eb37496e-374f-41b9-bd52-f52f37bdac31"
	calamityProbeOracle = "test-calamity-attack-probe"
)

// calamityProbeAttacks counts "whenever this creature attacks" triggers of
// the probe creature below, which Calamity copies.
var calamityProbeAttacks int

func init() {
	// A creature with "Whenever this creature attacks, ..." so a token copy
	// that was (wrongly) declared as an attacker would show it.
	Register(Spec{
		OracleID: calamityProbeOracle,
		Name:     "Calamity Probe Saddler",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			Key:     "Calamity Probe Saddler — count an attack",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Effect: func(_ *game.Game, _ *game.StackItem) error {
				calamityProbeAttacks++
				return nil
			},
		}},
	})
}

func pushCalamity(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Calamity, Galloping Inferno", calamityOracle, "Legendary Creature — Horse Mount", 4, 6)
}

func pushProbeSaddler(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Calamity Probe Saddler", calamityProbeOracle, "Creature — Test", 2, 2)
}

// calamityAttackPrompt saddles Calamity, attacks `defender` with it and
// settles the stack up to the first choice, which it returns.
func calamityAttackPrompt(t *testing.T, g *game.Game, me *game.Player, defender uuid.UUID, calamity uuid.UUID, saddlers ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, calamity, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, defender, calamity)
	passPriorityAroundTable(t, g)
	return latestChooseCards(g, me.ID)
}

// answerAttackSplit answers the "how many of the tokens attack X" prompts
// of ONE token: no for the first `skip` players offered, yes for the next.
// The last player offered takes the token without a prompt.
func answerAttackSplit(t *testing.T, g *game.Game, me *game.Player, skip int) {
	t.Helper()
	for i := 0; ; i++ {
		c := payAmountPrompt(g, me.ID)
		if c == nil {
			return
		}
		n := 0
		if i >= skip {
			n = 1
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, n); err != nil {
			t.Fatalf("attack split answer %d: %v", n, err)
		}
	}
}

func calamityTokens(g *game.Game, controller uuid.UUID, name string) []*game.Card {
	var out []*game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// The headline: two token copies enter tapped and attacking, the copied
// creature's attack trigger does not fire (CR 508.4), and both tokens are
// sacrificed at the beginning of the next end step.
func TestCalamityMakesTwoTappedAttackingCopiesSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)
	calamityProbeAttacks = 0

	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled Calamity")
	}
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 {
		t.Errorf("the choice is %d..%d, want exactly one", choice.ChooseMin, choice.ChooseMax)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("first choice: %v", err)
	}
	answerAttackSplit(t, g, me, 0)
	if got := len(calamityTokens(g, me.ID, "Calamity Probe Saddler")); got != 1 {
		t.Fatalf("after the first round there are %d copies, want 1", got)
	}

	choice = latestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("the second round did not ask")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("second choice: %v", err)
	}
	answerAttackSplit(t, g, me, 0)

	tokens := calamityTokens(g, me.ID, "Calamity Probe Saddler")
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
	passPriorityAroundTable(t, g)
	if calamityProbeAttacks != 0 {
		t.Errorf("the copies' attack trigger fired %d times; they were never declared as attackers (CR 508.4)", calamityProbeAttacks)
	}
	if _, ok := battlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler itself left the battlefield")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("token %s survived the beginning of the next end step", id)
		}
	}
	if _, ok := battlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was sacrificed with its copies")
	}
}

// Each token's controller chooses what it attacks: the two copies can
// attack different players.
func TestCalamityCopiesCanAttackDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)

	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	answerAttackSplit(t, g, me, 0) // the first copy keeps Calamity's defender
	choice = latestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("the second round did not ask")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	answerAttackSplit(t, g, me, 1) // the second goes to the next player offered

	tokens := calamityTokens(g, me.ID, "Calamity Probe Saddler")
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	nonlegendary := pushProbeSaddler(g, me.ID)
	legend := pushDiesCreatureForTest(g, me.ID, "Legendary Saddler", "", "Legendary Creature — Test", 2, 2)
	bystander := pushCrewerForTest(g, me.ID, "Bystander", 3)

	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, nonlegendary, legend)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	legend := pushDiesCreatureForTest(g, me.ID, "Legendary Saddler", "", "Legendary Creature — Test", 2, 2)

	if choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, legend); choice != nil {
		t.Fatalf("a prompt was queued with only a legendary saddler: %v", choice.ChooseCards)
	}
	if payAmountPrompt(g, me.ID) != nil {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, calamity, saddler); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, calamity)
	g.WithWriteLock(func() {
		if err := (BounceToHand{Target: saddler}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if choice := latestChooseCards(g, me.ID); choice != nil {
		t.Errorf("offered %v after the saddler left", choice.ChooseCards)
	}
}

// "Repeat this process once" may choose the same creature, and that makes
// two copies of it.
func TestCalamityMayChooseTheSameCreatureTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushCrewerForTest(g, me.ID, "Lone Saddler", 3)

	for round := 1; round <= 2; round++ {
		choice := latestChooseCards(g, me.ID)
		if round == 1 {
			choice = calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
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
		answerAttackSplit(t, g, me, 0)
	}
	if got := len(calamityTokens(g, me.ID, "Lone Saddler")); got != 2 {
		t.Errorf("%d copies of the same creature, want 2", got)
	}
	if latestChooseCards(g, me.ID) != nil {
		t.Error("a third round was offered")
	}
}

// The probe really does count a declared attack, so the zero in the
// headline test means something.
func TestCalamityProbeCountsADeclaredAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	probe := pushProbeSaddler(g, me.ID)
	calamityProbeAttacks = 0
	toMain(t, g)
	declareAttack(t, g, opp.ID, probe)
	passPriorityAroundTable(t, g)
	if calamityProbeAttacks != 1 {
		t.Errorf("the probe counted %d attacks, want 1", calamityProbeAttacks)
	}
}

// Calamity only triggers while saddled.
func TestCalamityUnsaddledAttackDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	pushProbeSaddler(g, me.ID)
	toMain(t, g)
	declareAttack(t, g, opp.ID, calamity)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("an unsaddled Calamity prompted")
	}
}
