package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_second_pass_test.go — the eleven cards of roadmap batch 02
// (#295, `edhrec_rank` 238–360) that were NOT covered by the first
// pass in batch02_test.go / #351.
//
// Two sessions worked #295 in parallel without knowing it. #351
// landed the 25-card "no new machinery" group; this file covers the
// eleven that group did not reach — nine of them because the batch
// triage had filed them under a blocker that had already moved, and
// two (Entomb, Buried Alive) because they needed a one-case engine
// fix to `searchDestZoneLocked`, which ships alongside.
//
// Naming is `b02b`-prefixed to stay clear of batch02_test.go's `b02`
// helpers, which this file deliberately does not touch.

const (
	b02bKetriaTriomeOracle      = "6bae00e8-06cf-4ac4-a1cc-757e454109fe"
	b02bDarksteelCitadelOracle  = "8dc067bf-f78f-4ac4-b6e7-b305c42cf0bc"
	b02bEntombOracle            = "299fc083-0834-4064-8344-f895aff68867"
	b02bBuriedAliveOracle       = "8203c621-a1a0-4865-8c9a-0d4064c86107"
	b02bSeethingSongOracle      = "64bf8929-f5f2-4d50-8667-13b1d007bcfc"
	b02bManaGeyserOracle        = "a8dba58b-2956-492e-ae30-49db2ae68e53"
	b02bHarrowOracle            = "705509e9-a034-4a5a-9c65-66f58748b8a2"
	b02bDiabolicIntentOracle    = "038519b9-bca8-4b27-b5ac-2409595469d0"
	b02bGrayMerchantOracle      = "38f3b157-0df4-409b-89cc-086e1531cd5b"
	b02bCraterhoofOracle        = "8c52bd39-0586-48ca-b263-17210cf9feb6"
	b02bDecanterOfEndlessWaterO = "8ae98ef8-8f52-4877-a08c-1fae5514184e"
)

// --- registration --------------------------------------------------

func TestBatch02SecondPassCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b02bKetriaTriomeOracle:      "Ketria Triome",
		b02bDarksteelCitadelOracle:  "Darksteel Citadel",
		b02bEntombOracle:            "Entomb",
		b02bBuriedAliveOracle:       "Buried Alive",
		b02bSeethingSongOracle:      "Seething Song",
		b02bManaGeyserOracle:        "Mana Geyser",
		b02bHarrowOracle:            "Harrow",
		b02bDiabolicIntentOracle:    "Diabolic Intent",
		b02bGrayMerchantOracle:      "Gray Merchant of Asphodel",
		b02bCraterhoofOracle:        "Craterhoof Behemoth",
		b02bDecanterOfEndlessWaterO: "Decanter of Endless Water",
	}
	if len(want) != 11 {
		t.Fatalf("the second pass ships 11 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestKetriaTriomeEntersTappedAndOffersThreeColours(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Ketria Triome", b02bKetriaTriomeOracle)
	top100AssertEnteredTapped(t, g, id, "Ketria Triome")

	spec, _ := Lookup(b02bKetriaTriomeOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U|R}" {
		t.Errorf("mana abilities %+v, want one producing {G|U|R}", spec.ManaAbilities)
	}
}

func TestDarksteelCitadelTapsForColorlessAndDeclaresIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Darksteel Citadel", "Artifact Land", b02bDarksteelCitadelOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// The keyword is declared even though the engine does not yet
	// enforce it (#176), so the day enforcement lands the card is
	// already correct.
	if !eotHasAbility(effectiveAbilities(t, g, land), "indestructible") {
		t.Error("printed indestructible did not reach the effective abilities")
	}
}

// It is an ARTIFACT land — the reason it is played at all.
func TestDarksteelCitadelCountsAsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Darksteel Citadel", "Artifact Land", b02bDarksteelCitadelOracle)
	c, ok := b02bLookup(g, land)
	if !ok || !c.IsArtifact() || !c.IsLand() {
		t.Errorf("Darksteel Citadel must be both artifact and land: %+v", c)
	}
}

// --- library searches into the graveyard ---------------------------
//
// Both of these fail without the ZoneGraveyard case added to
// searchDestZoneLocked: the search emitted an effect error and found
// nothing at all.

func TestEntombPutsTheChosenCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Reanimation Target", "Something Else")

	castCatalogSpell(t, g, "Entomb", "Instant", b02bEntombOracle, nil)
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Reanimation Target")
	if !b02bGraveyardHasNamed(me, "Reanimation Target") {
		t.Error("Entomb did not put the chosen card into the graveyard")
	}
}

func TestBuriedAliveTakesThreeCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Body A", "Body B", "Body C", "Body D"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
	}

	castCatalogSpell(t, g, "Buried Alive", "Sorcery", b02bBuriedAliveOracle, nil)
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Body A", "Body B", "Body C")
	for _, n := range []string{"Body A", "Body B", "Body C"} {
		if !b02bGraveyardHasNamed(me, n) {
			t.Errorf("%s should be in the graveyard", n)
		}
	}
	if b02bGraveyardHasNamed(me, "Body D") {
		t.Error("Buried Alive takes up to THREE, not everything")
	}
}

// --- additional costs ----------------------------------------------

func TestDiabolicIntentEatsACreatureAndTutorsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := seedCreature(g, "Fodder", me.ID)
	seedLibrary(me, "The Combo Piece")

	castWithSacrifice(t, g, "Diabolic Intent", "Sorcery", b02bDiabolicIntentOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is an additional COST and is paid at announce")
	}
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "The Combo Piece")
	if !b02bHandHasNamed(me, "The Combo Piece") {
		t.Error("Diabolic Intent did not put the chosen card into hand")
	}
}

func TestHarrowSacrificesALandAndFetchesTwoUntappedBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for _, n := range []string{"Island", "Mountain", "Plains"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Basic Land — " + n,
			Owner: me.ID, Controller: me.ID,
		})
	}

	castWithSacrifice(t, g, "Harrow", "Instant", b02bHarrowOracle, land)
	if g.Battlefield.Contains(land) {
		t.Error("Harrow's land sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Island", "Mountain")
	fetched := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID {
			continue
		}
		if c.Name == "Island" || c.Name == "Mountain" {
			fetched++
			if c.Tapped {
				t.Errorf("%s entered tapped; Harrow has no tapped clause", c.Name)
			}
		}
	}
	if fetched != 2 {
		t.Errorf("Harrow put %d lands onto the battlefield, want 2", fetched)
	}
}

// --- mana from a spell ---------------------------------------------

func TestSeethingSongAddsFiveRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	castCatalogSpell(t, g, "Seething Song", "Instant", b02bSeethingSongOracle, nil)
	passPriorityAroundTable(t, g)

	got := batch01PoolColors(me)
	if len(got) != 5 {
		t.Fatalf("pool %v, want five tokens", got)
	}
	for _, c := range got {
		if c != "R" {
			t.Errorf("pool %v, want all R", got)
			break
		}
	}
}

func TestManaGeyserCountsOnlyTappedOpponentLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	// Two tapped opponent lands, one untapped opponent land, one
	// tapped land of my own — only the first two count.
	for i := 0; i < 2; i++ {
		id := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
		b02bSetTapped(g, id, true)
	}
	seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	mine := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	b02bSetTapped(g, mine, true)

	castCatalogSpell(t, g, "Mana Geyser", "Sorcery", b02bManaGeyserOracle, nil)
	passPriorityAroundTable(t, g)

	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two R (two tapped opponent lands)", got)
	}
}

func TestDecanterOfEndlessWaterLiftsTheHandSizeCapAndFixes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, _ := Lookup(b02bDecanterOfEndlessWaterO)
	if !spec.NoMaxHandSize {
		t.Error("Decanter of Endless Water must declare NoMaxHandSize")
	}
	rock := seedPermanentWithOracle(g, me.ID, "Decanter of Endless Water", "Artifact", b02bDecanterOfEndlessWaterO)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick must offer all five colours, got %+v", pick)
	}
}

// --- devotion ------------------------------------------------------

func TestGrayMerchantDrainsForDevotionAndGainsTheTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Two black pips already on the board — one plain, one HYBRID,
	// which counts toward devotion to black (CR 700.5) — plus Gary's
	// own {B}{B} once he lands. Devotion 4.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Pip", TypeLine: "Enchantment",
		ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hybrid Pip", TypeLine: "Enchantment",
		ManaCost: "{B/R}", Owner: me.ID, Controller: me.ID,
	})
	lifeBefore := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			oppLife[p.ID] = p.Life
		}
	}

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gray Merchant of Asphodel",
		TypeLine: "Creature — Zombie", OracleID: b02bGrayMerchantOracle,
		ManaCost: "{3}{B}{B}", Power: 2, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	const wantX = 4
	drained := 0
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if got := oppLife[p.ID] - p.Life; got != wantX {
			t.Errorf("opponent lost %d, want %d (devotion)", got, wantX)
		}
		drained += wantX
	}
	if me.Life != lifeBefore+drained {
		t.Errorf("life %d → %d, want +%d (the TOTAL lost, not X)", lifeBefore, me.Life, drained)
	}
}

// A colourless board drains for nothing rather than erroring.
func TestGrayMerchantWithNoDevotionDrainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life

	// No mana cost on the seeded Gary, so devotion is 0.
	castCatalogSpell(t, g, "Gray Merchant of Asphodel", "Creature — Zombie", b02bGrayMerchantOracle, nil)
	passPriorityAroundTable(t, g)

	if opp.Life != before {
		t.Errorf("opponent life %d → %d, want unchanged at zero devotion", before, opp.Life)
	}
}

// --- Craterhoof ----------------------------------------------------

func TestCraterhoofPumpsByTheCreatureCountAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedCreature(g, "Bear", me.ID)

	castCatalogSpell(t, g, "Craterhoof Behemoth", "Creature — Beast", b02bCraterhoofOracle, nil)
	passPriorityAroundTable(t, g)

	// Bear + Craterhoof on the battlefield at resolution ⇒ X of 2.
	c, ok := b02bLookup(g, bear)
	if !ok {
		t.Fatal("the bear vanished")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("bear is %d/%d, want 4/4 (2/2 with +2/+2)", c.CurrentPower(), c.CurrentToughness())
	}
	if !eotHasAbility(effectiveAbilities(t, g, bear), "trample") {
		t.Error("Craterhoof did not grant trample")
	}
}

// CR 611.2c — the affected set is snapshotted at resolution, so a
// creature that arrives afterwards gets neither half.
func TestCraterhoofDoesNotPumpACreatureCastAfterIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	castCatalogSpell(t, g, "Craterhoof Behemoth", "Creature — Beast", b02bCraterhoofOracle, nil)
	passPriorityAroundTable(t, g)

	late := seedCreature(g, "Latecomer", me.ID)
	c, ok := b02bLookup(g, late)
	if !ok {
		t.Fatal("the latecomer vanished")
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("latecomer is %d/%d, want an unpumped 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- local helpers -------------------------------------------------

func b02bGraveyardHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func b02bHandHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func b02bSetTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
				return
			}
		}
	})
}

func b02bLookup(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() {
		c, ok = g.LookupCardForEffect(id)
	})
	return c, ok
}
