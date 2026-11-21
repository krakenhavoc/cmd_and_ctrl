package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

func TestBatch02SecondPassCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B02bKetriaTriomeOracle:      "Ketria Triome",
		effectstest.B02bDarksteelCitadelOracle:  "Darksteel Citadel",
		effectstest.B02bEntombOracle:            "Entomb",
		effectstest.B02bBuriedAliveOracle:       "Buried Alive",
		effectstest.B02bSeethingSongOracle:      "Seething Song",
		effectstest.B02bManaGeyserOracle:        "Mana Geyser",
		effectstest.B02bHarrowOracle:            "Harrow",
		effectstest.B02bDiabolicIntentOracle:    "Diabolic Intent",
		effectstest.B02bGrayMerchantOracle:      "Gray Merchant of Asphodel",
		effectstest.B02bCraterhoofOracle:        "Craterhoof Behemoth",
		effectstest.B02bDecanterOfEndlessWaterO: "Decanter of Endless Water",
	}
	if len(want) != 11 {
		t.Fatalf("the second pass ships 11 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Ketria Triome", effectstest.B02bKetriaTriomeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Ketria Triome")

	spec, _ := effects.Lookup(effectstest.B02bKetriaTriomeOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U|R}" {
		t.Errorf("mana abilities %+v, want one producing {G|U|R}", spec.ManaAbilities)
	}
}

func TestDarksteelCitadelTapsForColorlessAndDeclaresIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Darksteel Citadel", "Artifact Land", effectstest.B02bDarksteelCitadelOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// The keyword was declared before the engine enforced it, so the
	// card was already correct when enforcement landed (#380,
	// game/indestructible.go). This test only checks the declaration
	// reaches the effective abilities; indestructible's own tests
	// cover what it stops.
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, land), "indestructible") {
		t.Error("printed indestructible did not reach the effective abilities")
	}
}

// It is an ARTIFACT land — the reason it is played at all.
func TestDarksteelCitadelCountsAsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Darksteel Citadel", "Artifact Land", effectstest.B02bDarksteelCitadelOracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Reanimation Target", "Something Else")

	effectstest.CastCatalogSpell(t, g, "Entomb", "Instant", effectstest.B02bEntombOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerSearchNamed(t, g, me.ID, "Reanimation Target")
	if !effectstest.B02bGraveyardHasNamed(me, "Reanimation Target") {
		t.Error("Entomb did not put the chosen card into the graveyard")
	}
}

func TestBuriedAliveTakesThreeCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Body A", "Body B", "Body C", "Body D"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
	}

	effectstest.CastCatalogSpell(t, g, "Buried Alive", "Sorcery", effectstest.B02bBuriedAliveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerSearchNamed(t, g, me.ID, "Body A", "Body B", "Body C")
	for _, n := range []string{"Body A", "Body B", "Body C"} {
		if !effectstest.B02bGraveyardHasNamed(me, n) {
			t.Errorf("%s should be in the graveyard", n)
		}
	}
	if effectstest.B02bGraveyardHasNamed(me, "Body D") {
		t.Error("Buried Alive takes up to THREE, not everything")
	}
}

// --- additional costs ----------------------------------------------

func TestDiabolicIntentEatsACreatureAndTutorsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	effectstest.SeedLibrary(me, "The Combo Piece")

	effectstest.CastWithSacrifice(t, g, "Diabolic Intent", "Sorcery", effectstest.B02bDiabolicIntentOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is an additional COST and is paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerSearchNamed(t, g, me.ID, "The Combo Piece")
	if !effectstest.B02bHandHasNamed(me, "The Combo Piece") {
		t.Error("Diabolic Intent did not put the chosen card into hand")
	}
}

func TestHarrowSacrificesALandAndFetchesTwoUntappedBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for _, n := range []string{"Island", "Mountain", "Plains"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Basic Land — " + n,
			Owner: me.ID, Controller: me.ID,
		})
	}

	effectstest.CastWithSacrifice(t, g, "Harrow", "Instant", effectstest.B02bHarrowOracle, land)
	if g.Battlefield.Contains(land) {
		t.Error("Harrow's land sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerSearchNamed(t, g, me.ID, "Island", "Mountain")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	effectstest.CastCatalogSpell(t, g, "Seething Song", "Instant", effectstest.B02bSeethingSongOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	got := effectstest.Batch01PoolColors(me)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	// Two tapped opponent lands, one untapped opponent land, one
	// tapped land of my own — only the first two count.
	for i := 0; i < 2; i++ {
		id := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
		effectstest.B02bSetTapped(g, id, true)
	}
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B02bSetTapped(g, mine, true)

	effectstest.CastCatalogSpell(t, g, "Mana Geyser", "Sorcery", effectstest.B02bManaGeyserOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two R (two tapped opponent lands)", got)
	}
}

func TestDecanterOfEndlessWaterLiftsTheHandSizeCapAndFixes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(effectstest.B02bDecanterOfEndlessWaterO)
	if !spec.NoMaxHandSize {
		t.Error("Decanter of Endless Water must declare NoMaxHandSize")
	}
	rock := effectstest.SeedPermanentWithOracle(g, me.ID, "Decanter of Endless Water", "Artifact", effectstest.B02bDecanterOfEndlessWaterO)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick must offer all five colours, got %+v", pick)
	}
}

// --- devotion ------------------------------------------------------

func TestGrayMerchantDrainsForDevotionAndGainsTheTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Two black pips already on the board — one plain, one HYBRID,
	// which counts toward devotion to black (CR 700.5) — plus Gary's
	// own {B}{B} once he lands. Devotion 4.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Pip", TypeLine: "Enchantment",
		ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
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
		TypeLine: "Creature — Zombie", OracleID: effectstest.B02bGrayMerchantOracle,
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
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life

	// No mana cost on the seeded Gary, so devotion is 0.
	effectstest.CastCatalogSpell(t, g, "Gray Merchant of Asphodel", "Creature — Zombie", effectstest.B02bGrayMerchantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before {
		t.Errorf("opponent life %d → %d, want unchanged at zero devotion", before, opp.Life)
	}
}

// --- Craterhoof ----------------------------------------------------

func TestCraterhoofPumpsByTheCreatureCountAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	effectstest.CastCatalogSpell(t, g, "Craterhoof Behemoth", "Creature — Beast", effectstest.B02bCraterhoofOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Bear + Craterhoof on the battlefield at resolution ⇒ X of 2.
	c, ok := b02bLookup(g, bear)
	if !ok {
		t.Fatal("the bear vanished")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("bear is %d/%d, want 4/4 (2/2 with +2/+2)", c.CurrentPower(), c.CurrentToughness())
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("Craterhoof did not grant trample")
	}
}

// CR 611.2c — the affected set is snapshotted at resolution, so a
// creature that arrives afterwards gets neither half.
func TestCraterhoofDoesNotPumpACreatureCastAfterIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	effectstest.CastCatalogSpell(t, g, "Craterhoof Behemoth", "Creature — Beast", effectstest.B02bCraterhoofOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	late := effectstest.SeedCreature(g, "Latecomer", me.ID)
	c, ok := b02bLookup(g, late)
	if !ok {
		t.Fatal("the latecomer vanished")
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("latecomer is %d/%d, want an unpumped 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func b02bLookup(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() {
		c, ok = g.LookupCardForEffect(id)
	})
	return c, ok
}
