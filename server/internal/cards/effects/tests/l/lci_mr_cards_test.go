package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_mr_cards_test.go — The Lost Caverns of Ixalan, cards M–R, second
// batch (#4470). Each test plays the real catalog entry and asserts what
// its text says.

const (
	lciMRScytheOracle   = "99c5630e-cc04-4203-9006-3bc6a6807762"
	lciMROutOfAirOracle = "fcea912b-58b1-4068-ac4d-1bc89aa7f33c"
	lciMRAltisaurOracle = "6de64e59-0a47-46c4-90bf-c56e63266f19"
	lciMRPirateOracle   = "3ab13412-1eaf-40c0-905d-8a3ca0c51be5"
	lciMRFrogOracle     = "e88a2b71-c90d-4e16-99a3-93d11fe7a968"
)

func TestLciMREveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciMRScytheOracle:   "Malamet Scythe",
		lciMROutOfAirOracle: "Out of Air",
		lciMRAltisaurOracle: "Panicked Altisaur",
		lciMRPirateOracle:   "Plundering Pirate",
		lciMRFrogOracle:     "Poison Dart Frog",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Malamet Scythe ---------------------------------------------------------------

func TestLciMRMalametScytheAttachesOnEntryForPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	scythe := effectstest.CastCatalogSpell(t, g, "Malamet Scythe", "Artifact — Equipment", lciMRScytheOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered; the target is a creature you control")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, scythe); c == nil || c.AttachedTo.ID != bear {
		t.Fatal("the Scythe is not attached to the target")
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("equipped creature is %d/%d, want 4/4", p, tt)
	}
	spec, _ := effects.Lookup(lciMRScytheOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "Equip {4}" || !spec.Activated[0].Equip {
		t.Fatalf("activated abilities %+v, want Equip {4}", spec.Activated)
	}
}

func TestLciMRMalametScytheHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	scythe, err := effectstest.Ecl2CastAt(t, g, "Malamet Scythe", "Artifact — Equipment", lciMRScytheOracle, nil)
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, scythe); c == nil || c.AttachedTo.ID != bear {
		t.Fatal("the Scythe cast in the end step did not attach")
	}
}

// --- Out of Air -------------------------------------------------------------------

func TestLciMROutOfAirCostsTwoLessAtACreatureSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		generic        int
	}{
		{"creature", "Creature — Bear", 0},
		{"instant", "Instant", 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.CastCatalogSpell(t, g, "Test Spell", tc.typeLine, "", nil)
		c := effectstest.SelfPriced(t, g, me, lciMROutOfAirOracle, "Instant", "{2}{U}{U}", game.ZoneHand,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
		if c.Generic != tc.generic || len(c.Required) != 2 {
			t.Errorf("at a %s spell: generic %d, %d coloured, want generic %d and 2 blue", tc.name, c.Generic, len(c.Required), tc.generic)
		}
	}
}

func TestLciMROutOfAirCountersACreatureSpellForTwoBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.B06AddMana(me, "U", "U")
	effectstest.CastFromHandForTest(t, g, me, "Out of Air", "Instant", "{2}{U}{U}", lciMROutOfAirOracle, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left in the pool, the discounted cost is two", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(target) {
		t.Error("the targeted creature spell was not countered into the graveyard")
	}
}

func TestLciMROutOfAirAtANoncreatureSpellPaysFullPrice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.B06AddMana(me, "U", "U")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Out of Air", TypeLine: "Instant", ManaCost: "{2}{U}{U}", OracleID: lciMROutOfAirOracle,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Fatal("two mana paid for Out of Air at an instant spell")
	}
}

// --- Panicked Altisaur ------------------------------------------------------------

func TestLciMRPanickedAltisaurPingsEachOpponentForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saur := effectstest.B12Push(g, me.ID, "Panicked Altisaur", "Creature — Dinosaur", lciMRAltisaurOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, saur, "reach") {
		t.Error("Panicked Altisaur has no reach")
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, saur, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 2
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", p.Seat, before[p.ID], p.Life, want)
		}
	}
	if !effectstest.B13Tapped(t, g, saur) {
		t.Error("the ability did not tap the Altisaur")
	}
	if err := g.ActivateCatalogAbility(me.ID, saur, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated a second time while tapped")
	}
}

// --- Plundering Pirate ------------------------------------------------------------

func TestLciMRPlunderingPirateMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Plundering Pirate", TypeLine: "Creature — Orc Pirate", OracleID: lciMRPirateOracle, Power: 3, Toughness: 2})
	treasures := effectstest.TrekB33Tokens(g, me.ID, "Treasure")
	if len(treasures) != 1 {
		t.Fatalf("%d Treasure tokens, want 1", len(treasures))
	}
	if c := effectstest.LayeredCard(t, g, treasures[0].InstanceID); !c.IsArtifact() || !c.HasSubtype("Treasure") {
		t.Errorf("token = %s, want a Treasure artifact", c.TypeLine)
	}
}

// --- Poison Dart Frog -------------------------------------------------------------

func TestLciMRPoisonDartFrogTapsForAnyColorAndBuysDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	frog := effectstest.B12Push(g, me.ID, "Poison Dart Frog", "Creature — Frog", lciMRFrogOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, frog, "reach") {
		t.Error("Poison Dart Frog has no reach")
	}
	if effectstest.HasEffectiveKeyword(t, g, frog, "deathtouch") {
		t.Fatal("the Frog has deathtouch before the ability resolves")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, frog, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := me.ManaPool.CountColor("B"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("the Frog added %d black of %d mana, want {B}", n, me.ManaPool.Total())
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateCatalogAbility(me.ID, frog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the deathtouch ability activated with no mana for {2}")
	}
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, frog, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the deathtouch ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, frog, "deathtouch") {
		t.Fatal("the Frog did not gain deathtouch")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, frog, "deathtouch") {
		t.Error("deathtouch lasts only until end of turn")
	}
}
