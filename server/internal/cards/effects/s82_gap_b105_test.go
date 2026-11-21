package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b105_test.go — S82 GAP batch 105 (#5948). Each test plays the real
// catalog entry.

const (
	gapB105CapriciousOracle = "1f59b3d9-1535-4e13-a0a1-0f839fe90ac6"
	gapB105RipleyOracle     = "eefeeab1-b9f2-451a-9250-eb7b54d00fc1"
	gapB105ChampionOracle   = "d9dbe40f-38e9-460a-b84f-4aa22650017e"
	gapB105GargadonOracle   = "73f808f0-3d56-46c9-967b-d452bb3611af"
	gapB105RakkaMarOracle   = "6d92f14b-470b-40bc-b294-f725d46f307e"
	gapB105RorixOracle      = "48fe0f52-1676-49d5-acf6-5476081e8e1b"
	gapB105SpikeshotOracle  = "3cbd4c8e-fe3f-4a7c-90a5-7bc656d7bb10"
)

func TestGapB105CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB105CapriciousOracle: "Capricious Sliver",
		gapB105RipleyOracle:     "Captain Ripley Vance",
		gapB105ChampionOracle:   "Champion of the Flame",
		gapB105GargadonOracle:   "Gargadon",
		gapB105RakkaMarOracle:   "Rakka Mar",
		gapB105RorixOracle:      "Rorix Bladewing",
		gapB105SpikeshotOracle:  "Spikeshot Goblin",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB105CapriciousSliverGrantsTheImpulseToYourSlivers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Capricious Sliver", "Creature — Sliver", gapB105CapriciousOracle, 3, 3)
	other := b12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 2, 2)
	top := seedLibraryTop(me, "Top Card", "Sorcery", "{1}")
	attackWith(t, g, opp.ID, other)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile after a Sliver connected", b12ZoneOf(g, top))
	}
}

func TestGapB105CapriciousSliverDoesNotGrantItToNonSlivers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Capricious Sliver", "Creature — Sliver", gapB105CapriciousOracle, 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	top := seedLibraryTop(me, "Top Card", "Sorcery", "{1}")
	attackWith(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) == game.ZoneExile {
		t.Error("a non-Sliver's combat damage exiled the top card")
	}
}

func TestGapB105CapriciousSliverDoesNotGrantItToAnOpponentsSlivers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Capricious Sliver", "Creature — Sliver", gapB105CapriciousOracle, 3, 3)
	mine := b12Creature(g, me.ID, "My Sliver", "Creature — Sliver", 2, 2)
	top := seedLibraryTop(me, "Top Card", "Sorcery", "{1}")
	attackWith(t, g, opp.ID, mine)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) == game.ZoneExile {
		t.Error("the opponent's Capricious Sliver granted the impulse to my Sliver")
	}
}

func TestGapB105RipleyVanceGrowsAndShootsOnTheThirdSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ripley := b12Push(g, me.ID, "Captain Ripley Vance", "Legendary Creature — Human Pirate", gapB105RipleyOracle, 3, 2)
	toMainForCost(t, g)
	life := opp.Life
	tdmMPCastFreeSpell(t, g, me, "First")
	tdmMPCastFreeSpell(t, g, me, "Second")
	if opp.Life != life {
		t.Fatalf("life moved before the third spell: %d -> %d", life, opp.Life)
	}
	id := mayhemHandCard(g, me, "Third", "Sorcery", "{0}", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if p := effectivePowerWithCounters(t, g, ripley); p != 4 {
		t.Errorf("Ripley is %d power, want 4 after the counter", p)
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 4 damage (the power after the counter)", life, opp.Life)
	}
}

func TestGapB105RipleyVanceIgnoresAFourthSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ripley := b12Push(g, me.ID, "Captain Ripley Vance", "Legendary Creature — Human Pirate", gapB105RipleyOracle, 3, 2)
	toMainForCost(t, g)
	for _, n := range []string{"First", "Second"} {
		tdmMPCastFreeSpell(t, g, me, n)
	}
	id := mayhemHandCard(g, me, "Third", "Sorcery", "{0}", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	answerPickTargetPlayer(t, g, g.Seats[1].ID)
	passPriorityAroundTable(t, g)
	tdmMPCastFreeSpell(t, g, me, "Fourth")
	if p := effectivePowerWithCounters(t, g, ripley); p != 4 {
		t.Errorf("Ripley is %d power, want 4 (only the third spell triggers)", p)
	}
}

func TestGapB105ChampionOfTheFlameGrowsPerAuraAndEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := b12Push(g, me.ID, "Champion of the Flame", "Creature — Human Warrior", gapB105ChampionOracle, 1, 1)
	if p, tough := effectivePTOf(t, g, champ); p != 1 || tough != 1 {
		t.Fatalf("bare Champion is %d/%d, want 1/1", p, tough)
	}
	somB1Equip(t, g, me.ID, "Sword", "", champ)
	if p, tough := effectivePTOf(t, g, champ); p != 3 || tough != 3 {
		t.Errorf("with an Equipment it is %d/%d, want 3/3", p, tough)
	}
	// An opponent's Aura on it counts too.
	aura := b12Push(g, opp.ID, "Curse", "Enchantment — Aura", "", 0, 0)
	attachForTest(t, g, aura, champ)
	if p, tough := effectivePTOf(t, g, champ); p != 5 || tough != 5 {
		t.Errorf("with an Equipment and an Aura it is %d/%d, want 5/5", p, tough)
	}
	// A bare Equipment on another creature does not.
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	somB1Equip(t, g, me.ID, "Other Sword", "", other)
	if p, _ := effectivePTOf(t, g, champ); p != 5 {
		t.Errorf("an Equipment on another creature changed the Champion to %d power", p)
	}
	if !hasKeywordOnBattlefield(t, g, champ, "trample") {
		t.Error("the Champion lacks trample")
	}
}

func TestGapB105GargadonIsSuspend4ForOneRed(t *testing.T) {
	spec, _ := Lookup(gapB105GargadonOracle)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	sa := spec.SpecialActions[0]
	if sa.Kind != game.SpecialActionSuspend || sa.Counters != 4 || sa.Cost != "{1}{R}" {
		t.Errorf("special action %+v, want suspend 4—{1}{R}", sa)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Gargadon", "Creature — Beast", gapB105GargadonOracle, 7, 5)
	if !hasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("Gargadon lacks trample")
	}
}

func TestGapB105RakkaMarMakesAHastyElementalForRedAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rakka := b12Push(g, me.ID, "Rakka Mar", "Legendary Creature — Human Shaman", gapB105RakkaMarOracle, 2, 2)
	if err := ravB2Activate(t, g, me.ID, rakka, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was paid for without {R}")
	}
	ravB2MustActivate(t, g, me.ID, rakka, 0, game.ActivateAbilityParams{}, "R")
	ids := battlefieldIDsNamed(g, "Elemental")
	if len(ids) != 1 {
		t.Fatalf("%d Elementals, want 1", len(ids))
	}
	tok, _ := g.LookupCardForEffect(ids[0])
	if tok.Power != 3 || tok.Toughness != 1 || !tok.IsCreature() || !hasKeywordOnBattlefield(t, g, ids[0], "haste") {
		t.Errorf("token %d/%d %q, want a 3/1 Elemental with haste", tok.Power, tok.Toughness, tok.TypeLine)
	}
	if err := ravB2Activate(t, g, me.ID, rakka, 0, game.ActivateAbilityParams{}, "R"); err == nil {
		t.Error("Rakka Mar activated again while tapped")
	}
	if !hasKeywordOnBattlefield(t, g, rakka, "haste") {
		t.Error("Rakka Mar lacks haste")
	}
}

func TestGapB105RorixBladewingFliesWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Rorix Bladewing", "Legendary Creature — Dragon", gapB105RorixOracle, 6, 5)
	for _, kw := range []string{"flying", "haste"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("Rorix lacks %s", kw)
		}
	}
}

func TestGapB105SpikeshotGoblinDealsItsPowerAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	goblin := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spikeshot Goblin", TypeLine: "Creature — Goblin Shaman", OracleID: gapB105SpikeshotOracle,
		Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, goblin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}, "R")
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 damage from a 3-power Goblin", life, opp.Life)
	}
	if err := ravB2Activate(t, g, me.ID, goblin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}, "R"); err == nil {
		t.Error("the tapped Goblin activated again")
	}
}
