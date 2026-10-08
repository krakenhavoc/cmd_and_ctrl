package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boast_cards_test.go — #2697: one behaviour test per boast card whose
// effect is more than a token or a counter. The gate itself (attack,
// once a turn, Birgi) is in boast_test.go.

const (
	duskwielderOracle      = "b9227f38-7db7-4ea1-9482-94e5de3984fe"
	axgardBraggartOracle   = "d0d4ab83-8b9b-49cd-86b8-720abd0550f4"
	battershieldOracle     = "f8e8cdc6-5685-4986-a14b-20bb9e31d9cf"
	draugrRecruiterOracle  = "f854ea0d-aae3-4b6a-9767-6ac1788f2190"
	fearlessPupOracle      = "61295adf-9a58-479a-b6e5-403aa0876c22"
	horizonSeekerOracle    = "e0610ab1-88d3-4502-84b7-3a08d0170167"
	usherOfTheFallenOracle = "a6eb06dc-a62d-4fdb-b336-e304d8d68c92"
)

// poolFor fills a seat with plenty of every colour, so a test is about
// the effect and not the price.
func poolFor(p *game.Player) {
	fillPool(p, 6)
	for _, c := range []string{"R", "G", "W", "B"} {
		fillPoolColored(p, c, 2)
	}
}

// attackAndBoast puts a catalog creature on the battlefield, attacks
// with it, and activates its only ability with params.
func attackAndBoast(t *testing.T, g *game.Game, oracle, name string, params game.ActivateAbilityParams) (uuid.UUID, error) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, name, "Creature — Test", oracle, false)
	attackWith(t, g, opp.ID, id)
	poolFor(me)
	return id, g.ActivateCatalogAbility(me.ID, id, 0, params)
}

func TestDuskwielderDrainsTheTargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, err := attackAndBoast(t, g, duskwielderOracle, "Duskwielder", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if err != nil {
		t.Fatalf("boast: %v", err)
	}
	myLife, theirLife := me.Life, opp.Life
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life me %d->%d opp %d->%d, want +1 and -1", myLife, me.Life, theirLife, opp.Life)
	}
}

func TestAxgardBraggartUntapsAndGrowsItself(t *testing.T) {
	g := newCatalogGame(t)
	id, err := attackAndBoast(t, g, axgardBraggartOracle, "Axgard Braggart", game.ActivateAbilityParams{})
	if err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if c, ok := battlefieldCard(g, id); !ok || c.Tapped {
		t.Errorf("Axgard Braggart should be untapped after its boast (found=%v)", ok)
	}
}

func TestBattershieldWarriorPumpsEveryCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if _, err := attackAndBoast(t, g, battershieldOracle, "Battershield Warrior", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("my Bear power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Bear power = %d, want 2 — \"creatures you control\"", got)
	}
}

func TestFearlessPupGetsPlusTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	id, err := attackAndBoast(t, g, fearlessPupOracle, "Fearless Pup", game.ActivateAbilityParams{})
	if err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, id); got != 3 {
		t.Errorf("Pup power = %d, want 3 (1 + 2)", got)
	}
}

func TestDraugrRecruiterReturnsACreatureCardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	_, err := attackAndBoast(t, g, draugrRecruiterOracle, "Draugr Recruiter", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	})
	if err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == dead {
			found = true
		}
	}
	if !found {
		t.Error("the targeted creature card did not return to hand")
	}
}

func TestHorizonSeekerFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	if _, err := attackAndBoast(t, g, horizonSeekerOracle, "Horizon Seeker", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}
}

func TestVarragothLetsTheTargetPlayerTutorForTheirOwnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(opp,
		game.Card{Name: "Filler One", TypeLine: "Creature — Bear"},
		game.Card{Name: "Needle", TypeLine: "Instant"},
		game.Card{Name: "Filler Two", TypeLine: "Creature — Bear"})
	if _, err := attackAndBoast(t, g, varragothOracle, "Varragoth, Bloodsky Sire", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("the boaster was asked to search — the TARGET searches")
	}
	answerSearchNamed(t, g, opp.ID, "Needle")
	if top, err := opp.Library.Top(); err != nil || top.Name != "Needle" {
		t.Errorf("the found card is not on top of the target's library: top = %q (%v)", top.Name, err)
	}
}

func TestEradicatorValkyrieMakesEachOpponentSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if _, err := attackAndBoast(t, g, eradicatorValkyrieOracle, "Eradicator Valkyrie", game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the boaster was asked to sacrifice — it is each OPPONENT")
	}
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent's creature survived")
	}
}

func TestUsherOfTheFallenMakesAHumanWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if _, err := attackAndBoast(t, g, usherOfTheFallenOracle, "Usher of the Fallen", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := b14Tokens(g, me.ID, "Human Warrior"); n != 1 {
		t.Errorf("Human Warrior tokens = %d, want 1", n)
	}
}
