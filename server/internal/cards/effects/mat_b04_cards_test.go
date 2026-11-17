package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mat_b04_cards_test.go — March of the Machine: The Aftermath, batch 4
// (#5476): Narset, Nashi, Rebuild the City, Calix, Vesuvan Drifter, Jolrael,
// Spark Rupture, Urborg Scavengers, Tazri.

const (
	matB04NarsetOracle  = "d92725de-ead0-4a6c-83db-55ec3acc7184"
	matB04NashiOracle   = "ffa6c14e-3d8e-47a2-b102-d37538d60ef6"
	matB04RebuildOracle = "2aac0ddc-1005-4049-a337-fe86e092b461"
)

func matB04Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
}

// matB04Attack declares atk as attacking opp and locks the attack in.
func matB04Attack(t *testing.T, g *game.Game, atk, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}

// --- Narset, Enlightened Exile ---------------------------------------------

func TestMatB04NarsetCopiesAndCastsACardFromAnyGraveyardForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	narset := b12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	life := opp.Life

	matB04Attack(t, g, narset, opp.ID)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	if !inExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("the copy is cast without paying its mana cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 from the copied Bolt", life, opp.Life)
	}
	if !inExile(g, bolt) {
		t.Error("the original stays in exile")
	}
}

// A creature card, a land and a card whose mana value is not less than
// Narset's power are none of them targets, so the trigger is not put on the
// stack (CR 603.3d).
func TestMatB04NarsetRefusesCreaturesLandsAndTooExpensiveCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	narset := b12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	gyCard(opp, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", vampireNighthawkOracle)
	gyCard(opp, "Forest", "Basic Land — Forest", "", "")
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	matB04Attack(t, g, narset, opp.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, narset) != nil {
		t.Fatal("no card qualifies: Narset's trigger must not be put on the stack")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Fatalf("%d copies made", n)
	}
}

func TestMatB04NarsetGivesEveryCreatureYouControlProwess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	b12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	mine := p1Creature(g, me.ID, 2, 2)
	theirs := p1Creature(g, opp.ID, 2, 2)
	if c := effectiveCardForTest(t, g, mine); !game.HasKeyword(&c, "prowess") {
		t.Error("my creature has prowess")
	}
	if c := effectiveCardForTest(t, g, theirs); game.HasKeyword(&c, "prowess") {
		t.Error("an opponent's creature does not")
	}
}

// --- Nashi, Moon's Legacy --------------------------------------------------

func TestMatB04NashiCopiesARatAndTheCopyBecomesAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := b12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	rat := gyCard(me, "Vampire Nighthawk", "Creature — Rat Shaman", "{1}{B}{B}", vampireNighthawkOracle)
	notMine := gyCard(opp, "Vampire Nighthawk", "Creature — Rat Shaman", "{1}{B}{B}", vampireNighthawkOracle)
	plain := gyCard(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", vampireNighthawkOracle)

	matB04Attack(t, g, nashi, opp.ID)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{notMine, plain} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("card %s is not a legal target", bad)
		}
	}
	pickCard(t, g, me.ID, rat)
	passPriorityAroundTable(t, g)

	if !inExile(g, rat) {
		t.Fatal("the Rat card is exiled")
	}
	cp := copyOf(t, g, "Vampire Nighthawk")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err == nil {
		t.Fatal("Nashi's copy is cast paying its costs, not free")
	}
	floatForTest(g, me, "BBB")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy for {1}{B}{B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := tokensNamed(g, "Vampire Nighthawk")
	if len(toks) != 1 || toks[0].Controller != me.ID {
		t.Fatalf("a permanent copy resolves as a token under my control, got %d", len(toks))
	}
	if !inExile(g, rat) {
		t.Error("the original stays in exile")
	}
}

func TestMatB04NashiHasNoTriggerWithoutALegendaryOrRatCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := b12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	matB04Attack(t, g, nashi, opp.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, nashi) != nil {
		t.Fatal("no legendary or Rat card: no trigger")
	}
}

func TestMatB04NashiOffersLegendaryCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := b12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	leg := gyCard(me, "Lightning Bolt", "Legendary Instant", "{R}", lightningBoltOracle)
	matB04Attack(t, g, nashi, opp.ID)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, leg)
	passPriorityAroundTable(t, g)
	if !inExile(g, leg) {
		t.Fatal("a legendary card is exiled")
	}
	declineResolutionCasts(t, g, me.ID)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining", n)
	}
}

// --- Rebuild the City ------------------------------------------------------

func TestMatB04RebuildTheCityMakesThreeSummoningSick33Lands(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := matB04Seats(g)
	forest := b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	toMainForCost(t, g)
	castCatalogSpell(t, g, "Rebuild the City", "Sorcery", matB04RebuildOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	passPriorityAroundTable(t, g)

	toks := tokensNamed(g, "Forest")
	if len(toks) != 3 {
		t.Fatalf("%d Forest tokens, want 3", len(toks))
	}
	for _, tk := range toks {
		e := effectiveCardForTest(t, g, tk.InstanceID)
		if !e.IsLand() || !e.IsCreature() {
			t.Errorf("token is a land creature: %q", e.TypeLine)
		}
		if e.CurrentPower() != 3 || e.CurrentToughness() != 3 {
			t.Errorf("token is %d/%d, want 3/3", e.CurrentPower(), e.CurrentToughness())
		}
		if !game.HasKeyword(&e, "vigilance") || !game.HasKeyword(&e, "menace") {
			t.Errorf("token has vigilance and menace: %v", e.Keywords)
		}
	}
	if o := effectiveCardForTest(t, g, forest); o.IsCreature() {
		t.Error("the original land is not changed")
	}
}
