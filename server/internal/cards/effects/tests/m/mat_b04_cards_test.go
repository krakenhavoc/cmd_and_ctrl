package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
}

// --- Narset, Enlightened Exile ---------------------------------------------

func TestMatB04NarsetCopiesAndCastsACardFromAnyGraveyardForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	narset := effectstest.B12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	life := opp.Life

	matB04Attack(t, g, narset, opp.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("the copy is cast without paying its mana cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 from the copied Bolt", life, opp.Life)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original stays in exile")
	}
}

// A creature card, a land and a card whose mana value is not less than
// Narset's power are none of them targets, so the trigger is not put on the
// stack (CR 603.3d).
func TestMatB04NarsetRefusesCreaturesLandsAndTooExpensiveCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	narset := effectstest.B12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	effectstest.GyCard(opp, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", effectstest.VampireNighthawkOracle)
	effectstest.GyCard(opp, "Forest", "Basic Land — Forest", "", "")
	effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	matB04Attack(t, g, narset, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, narset) != nil {
		t.Fatal("no card qualifies: Narset's trigger must not be put on the stack")
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Fatalf("%d copies made", n)
	}
}

func TestMatB04NarsetGivesEveryCreatureYouControlProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	effectstest.B12Push(g, me.ID, "Narset, Enlightened Exile", "Legendary Creature — Human Monk", matB04NarsetOracle, 3, 4)
	mine := effectstest.P1Creature(g, me.ID, 2, 2)
	theirs := effectstest.P1Creature(g, opp.ID, 2, 2)
	if c := effectstest.EffectiveCardForTest(t, g, mine); !game.HasKeyword(&c, "prowess") {
		t.Error("my creature has prowess")
	}
	if c := effectstest.EffectiveCardForTest(t, g, theirs); game.HasKeyword(&c, "prowess") {
		t.Error("an opponent's creature does not")
	}
}

// --- Nashi, Moon's Legacy --------------------------------------------------

func TestMatB04NashiCopiesARatAndTheCopyBecomesAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := effectstest.B12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	rat := effectstest.GyCard(me, "Vampire Nighthawk", "Creature — Rat Shaman", "{1}{B}{B}", effectstest.VampireNighthawkOracle)
	notMine := effectstest.GyCard(opp, "Vampire Nighthawk", "Creature — Rat Shaman", "{1}{B}{B}", effectstest.VampireNighthawkOracle)
	plain := effectstest.GyCard(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", effectstest.VampireNighthawkOracle)

	matB04Attack(t, g, nashi, opp.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{notMine, plain} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("card %s is not a legal target", bad)
		}
	}
	effectstest.PickCard(t, g, me.ID, rat)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, rat) {
		t.Fatal("the Rat card is exiled")
	}
	cp := effectstest.CopyOf(t, g, "Vampire Nighthawk")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true); err == nil {
		t.Fatal("Nashi's copy is cast paying its costs, not free")
	}
	effectstest.FloatForTest(g, me, "BBB")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy for {1}{B}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.TokensNamed(g, "Vampire Nighthawk")
	if len(toks) != 1 || toks[0].Controller != me.ID {
		t.Fatalf("a permanent copy resolves as a token under my control, got %d", len(toks))
	}
	if !effectstest.InExile(g, rat) {
		t.Error("the original stays in exile")
	}
}

func TestMatB04NashiHasNoTriggerWithoutALegendaryOrRatCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := effectstest.B12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	matB04Attack(t, g, nashi, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, nashi) != nil {
		t.Fatal("no legendary or Rat card: no trigger")
	}
}

func TestMatB04NashiOffersLegendaryCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	nashi := effectstest.B12Push(g, me.ID, "Nashi, Moon's Legacy", "Legendary Creature — Rat Shaman", matB04NashiOracle, 3, 4)
	leg := effectstest.GyCard(me, "Lightning Bolt", "Legendary Instant", "{R}", effectstest.LightningBoltOracle)
	matB04Attack(t, g, nashi, opp.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, leg)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, leg) {
		t.Fatal("a legendary card is exiled")
	}
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining", n)
	}
}

// --- Rebuild the City ------------------------------------------------------

func TestMatB04RebuildTheCityMakesThreeSummoningSick33Lands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.ToMainForCost(t, g)
	effectstest.CastCatalogSpell(t, g, "Rebuild the City", "Sorcery", matB04RebuildOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	effectstest.PassPriorityAroundTable(t, g)

	toks := effectstest.TokensNamed(g, "Forest")
	if len(toks) != 3 {
		t.Fatalf("%d Forest tokens, want 3", len(toks))
	}
	for _, tk := range toks {
		e := effectstest.EffectiveCardForTest(t, g, tk.InstanceID)
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
	if o := effectstest.EffectiveCardForTest(t, g, forest); o.IsCreature() {
		t.Error("the original land is not changed")
	}
}
