package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b142_test.go — S82 GAP batch 142 (Dominaria Remastered), #5948:
// the cards through the real catalog.

const (
	gapSolarBlast      = "ece3e379-c586-48bd-8742-472b5deccc35"
	gapOrimsThunder    = "380429d5-82db-449c-b9b9-3e82ab987972"
	gapOvinize         = "98799336-a61e-4e35-ac30-51edb81301f6"
	gapRemedy          = "c644f190-63e5-4496-8612-bbe7be91afc9"
	gapRenewedFaith    = "aa4ecefa-808e-4706-af16-4ea5af9743fc"
	gapTerminalMoraine = "58eaaa8b-45c6-439b-bdd1-5f4e77a75a8c"
	gapDarigaaz        = "19b58ec9-bb88-4193-8ea8-c8f09ceec1ed"
	gapCrosis          = "e9a7dede-3968-4b0e-a707-419d46a6fec9"
	gapDromar          = "d8b57707-796d-4488-8f91-65bb75bc6281"
	gapSlipperyKarst   = "58bfd9a1-67ce-41d0-be38-f05addd1dd9e"
	gapRith            = "e13289e5-370b-435b-a38e-cf57c3078cec"
)

func gapB142PT(g *game.Game, id uuid.UUID) (p, tough int) {
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return
}

// gapB142PlayLand puts a land with the given type line in the active seat's hand
// and plays it.
func gapB142PlayLand(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func TestGapB142Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gapSolarBlast: "Solar Blast", gapOrimsThunder: "Orim's Thunder", gapOvinize: "Ovinize",
		gapRemedy: "Remedy", gapRenewedFaith: "Renewed Faith", gapTerminalMoraine: "Terminal Moraine",
		gapDarigaaz: "Darigaaz's Caldera", gapCrosis: "Crosis's Catacombs", gapDromar: "Dromar's Cavern",
		gapSlipperyKarst: "Slippery Karst", gapRith: "Rith's Grove",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGapB142SolarBlastDealsThreeAndItsCycleTriggerDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Solar Blast", "Instant", gapSolarBlast, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Fatalf("cast: life %d, want %d", got, life-3)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	bear := effectstest.BlbB1Push(g2, opp2.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CycleFromHand(t, g2, "Solar Blast", "Instant", gapSolarBlast, "{C}{R}{R}")
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	effectstest.AssistCardSettle(t, g2, me2.ID, &pick, true)
	g2.ReadSnapshot(func() {
		c, _ := g2.LookupCardForEffect(bear)
		if c.DamageMarked != 1 {
			t.Errorf("cycled: %d damage marked, want 1", c.DamageMarked)
		}
	})
}

func TestGapB142OrimsThunderDestroysAndKickedDamagesByManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Orim's Thunder", "Instant", gapOrimsThunder, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact survived")
	}

	// Kicked: the creature takes damage equal to the destroyed permanent's mana value.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	rock2 := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: opp2.ID, Controller: opp2.ID,
	})
	troll := effectstest.BlbB1Push(g2, opp2.ID, "Troll", "Creature — Troll", "", 5, 5)
	effectstest.AdvanceToMain(t, g2)
	effectstest.B06AddMana(me2, "W", "W", "W", "R")
	effectstest.PaCast(t, g2, "Orim's Thunder", "Instant", "{2}{W}", gapOrimsThunder, game.CastSpellParams{
		Strict: true, OptionalCosts: []int{0}, Targets: effectstest.CardRefs(rock2, troll),
	})
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(rock2) {
		t.Error("kicked: the artifact survived")
	}
	g2.ReadSnapshot(func() {
		c, _ := g2.LookupCardForEffect(troll)
		if c.DamageMarked != 3 {
			t.Errorf("kicked: %d damage marked on the creature, want 3", c.DamageMarked)
		}
	})
}

func TestGapB142OvinizeSetsBaseAndStripsAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	drake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Power: 4, Toughness: 4,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Ovinize", "Instant", gapOvinize, effectstest.CardTarget(drake))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := gapB142PT(g, drake); p != 0 || th != 1 {
		t.Errorf("got %d/%d, want 0/1", p, th)
	}
	if effectstest.P1g2HasKeyword(g, drake, "flying") {
		t.Error("the creature kept flying")
	}
}

func TestGapB142RemedyPreventsFiveDivided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Remedy", "Instant", gapRemedy)
	targets := append(effectstest.CardRefs(bear), effectstest.PlayerTarget(me.ID))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: targets, Distribution: map[uuid.UUID]int{bear: 3, me.ID: 3},
	}); err == nil {
		t.Fatal("a division of six was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: targets, Distribution: map[uuid.UUID]int{bear: 3, me.ID: 2},
	}); err != nil {
		t.Fatalf("a legal division: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 4) })
	if d := effectstest.E2Card(t, g, bear).DamageMarked; d != 1 {
		t.Errorf("4 damage to the Bear with a 3 shield left %d marked, want 1", d)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 5) })
	if got := me.Life; got != life-3 {
		t.Errorf("5 damage to a player with a 2 shield: life %d, want %d", got, life-3)
	}
}

func TestGapB142RenewedFaithGainsSixAndCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Renewed Faith", "Instant", gapRenewedFaith, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+6 {
		t.Fatalf("cast: life %d, want %d", got, life+6)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, _, _ := effectstest.BlbB1Seats(g2)
	life2 := effectstest.LifeOf(g2, me2.ID)
	effectstest.CycleFromHand(t, g2, "Renewed Faith", "Instant", gapRenewedFaith, "{C}{W}")
	effectstest.AssistCardSettle(t, g2, me2.ID, nil, true)
	if got := effectstest.LifeOf(g2, me2.ID); got != life2+2 {
		t.Errorf("cycled and accepted: life %d, want %d", got, life2+2)
	}

	g3 := effectstest.NewCatalogGame(t)
	me3, _, _ := effectstest.BlbB1Seats(g3)
	life3 := effectstest.LifeOf(g3, me3.ID)
	effectstest.CycleFromHand(t, g3, "Renewed Faith", "Instant", gapRenewedFaith, "{C}{W}")
	effectstest.AssistCardSettle(t, g3, me3.ID, nil, false)
	if got := effectstest.LifeOf(g3, me3.ID); got != life3 {
		t.Errorf("cycled and declined: life %d, want %d", got, life3)
	}
}

func TestGapB142TerminalMoraineFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	moraine := effectstest.PushCatalogPermanent(g, me.ID, "Terminal Moraine", "Land", gapTerminalMoraine, false)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	dual := effectstest.StapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	if err := g.ActivateCatalogAbility(me.ID, moraine, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {2}")
	}
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, moraine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if got, ok := effectstest.AangCardOnBF(g, forest); !ok || !got.Tapped {
		t.Error("the basic land did not enter tapped")
	}
	if _, wrong := effectstest.AangCardOnBF(g, dual); wrong {
		t.Error("a nonbasic was fetched")
	}
	if g.Battlefield.Contains(moraine) {
		t.Error("Terminal Moraine was not sacrificed")
	}
}

func TestGapB142LairLandsReturnANonLairLandOrAreSacrificed(t *testing.T) {
	for _, tc := range []struct{ oracle, name string }{
		{gapDarigaaz, "Darigaaz's Caldera"}, {gapCrosis, "Crosis's Catacombs"},
		{gapDromar, "Dromar's Cavern"}, {gapRith, "Rith's Grove"},
	} {
		// With a non-Lair land: return it.
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		other := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
		lair := gapB142PlayLand(t, g, tc.name, "Land — Lair", tc.oracle)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChooseCardsFor(g, me.ID)
		if prompt == nil {
			t.Fatalf("%s: the trigger must ask which land to return", tc.name)
		}
		if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != other {
			t.Fatalf("%s: offered %v, want only the Forest", tc.name, prompt.ChooseCards)
		}
		if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{other}); err != nil {
			t.Fatalf("%s: ResolveChooseCards: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
			t.Errorf("%s: the chosen land should be back in hand", tc.name)
		}
		if !g.Battlefield.Contains(lair) {
			t.Errorf("%s: the Lair was sacrificed despite the return", tc.name)
		}
		var tapped bool
		g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(lair); tapped = c.Tapped })
		if tapped {
			t.Errorf("%s: a Lair enters untapped", tc.name)
		}
	}
}

func TestGapB142ALairIsSacrificedWithOnlyLairsToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	otherLair := effectstest.SeedLandOnBattlefield(g, me.ID, "Treva's Ruins", "Land — Lair")
	lair := gapB142PlayLand(t, g, "Rith's Grove", "Land — Lair", gapRith)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lair) {
		t.Error("the Lair stayed with no non-Lair land to return")
	}
	if !g.Battlefield.Contains(otherLair) {
		t.Error("a Lair was returned as though it were a non-Lair land")
	}
}

func TestGapB142LairsTapForTheirThreeColors(t *testing.T) {
	for oracle, want := range map[string]string{
		gapDarigaaz: "{B|R|G}", gapCrosis: "{U|B|R}", gapDromar: "{W|U|B}", gapRith: "{R|G|W}",
	} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s produces %+v, want %s", spec.Name, spec.ManaAbilities, want)
		}
	}
}

func TestGapB142SlipperyKarstEntersTappedAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	karst := gapB142PlayLand(t, g, "Slippery Karst", "Land", gapSlipperyKarst)
	effectstest.Top100AssertEnteredTapped(t, g, karst, "Slippery Karst")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLibrary(me2, "A", "B")
	n := len(me2.Library.Cards)
	effectstest.CycleFromHand(t, g2, "Slippery Karst", "Land", gapSlipperyKarst, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g2)
	if len(me2.Library.Cards) != n-1 {
		t.Errorf("cycling drew %d cards, want 1", n-len(me2.Library.Cards))
	}
}
