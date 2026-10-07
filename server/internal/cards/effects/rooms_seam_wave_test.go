package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_seam_wave_test.go — the eight Rooms that waited on single-door
// mechanics (home tracker #2555). Each test drives a door that works;
// the doors declared unimplemented in their caveats have no behaviour
// to test, and TestRoomsSeamWaveDeclareTheirGaps pins the caveat so the
// gap cannot silently close behind a stale claim.

const (
	moldGymOracle      = "34d0d070-fcaf-410e-a002-012cbe104fb7"
	ticketBoothOracle  = "4d01b62b-b924-4da5-8ff5-b2f29d7f19b2"
	underwaterOracle   = "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4"
	expLabOracle       = "8869df0c-fe84-4964-8d38-a8ecefa9c252"
	charredFoyerOracle = "9f50ec9f-78e6-403f-8f41-84d603fee0eb"
	crampedVentsOracle = "1e847d69-e527-4bb9-a624-17c25ac4fed4"
	dazzlingOracle     = "c46a02db-13d6-477f-9da0-822599470168"
	secretArcadeOracle = "c8abde48-a07a-42a0-a43b-c357e6d9cad4"
)

func TestRoomsSeamWaveDeclareTheirGaps(t *testing.T) {
	// Moldering Gym, Ticket Booth and Underwater Tunnel are not here any
	// more: their manifest-dread doors shipped with #2570 and they are
	// complete. So is Experimental Lab // Staff Room, whose Staff Room
	// turns a creature face up since #2590.
	for _, id := range []string{
		charredFoyerOracle, crampedVentsOracle, dazzlingOracle, secretArcadeOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
			t.Errorf("%s ships a door that does nothing, so it must declare caveats", spec.Name)
		}
	}
}

func TestMolderingGymFetchesABasicLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	other := pushLibraryCardForTest(me, game.Card{Name: "Nonbasic", TypeLine: "Land"})
	advanceToMain(t, g)
	roomsBCast(t, g, me, roomsBCard(me.ID, moldGymOracle, "Moldering Gym", "{2}{G}", "Weight Room", "{5}{G}"), 0)
	roomsBSettle(t, g, me, forest)
	c := findBattlefieldCardForTest(g, forest)
	if c == nil || !c.Tapped {
		t.Fatalf("the basic land did not enter tapped: %+v", c)
	}
	if findBattlefieldCardForTest(g, other) != nil {
		t.Error("a nonbasic land was fetched")
	}
}

func TestTunnelOfHateGivesTheChosenAttackerDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker B", 2, 2)
	c := roomsBCard(me.ID, ticketBoothOracle, "Ticket Booth", "{2}{R}", "Tunnel of Hate", "{4}{R}{R}")
	roomsBCast(t, g, me, c, 1)
	roomsBSettle(t, g, me)
	declareAttack(t, g, opp.ID, a, b)
	roomsBSettle(t, g, me, b)
	if effectiveAbilitiesContain(t, g, a, "double strike") || !effectiveAbilitiesContain(t, g, b, "double strike") {
		t.Fatal("double strike should go to the chosen attacker only")
	}
}

func TestUnderwaterTunnelSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Sorcery"})
	pushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: "Sorcery"})
	advanceToMain(t, g)
	roomsBCast(t, g, me, roomsBCard(me.ID, underwaterOracle, "Underwater Tunnel", "{U}", "Slimy Aquarium", "{3}{U}"), 0)
	for i := 0; i < 24; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == me.ID {
				if len(c.ScryCards) != 2 {
					t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
				}
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Logf("pass: %v", err)
		}
	}
	t.Fatal("no surveil prompt")
}

func TestStaffRoomCountersTheCreatureThatConnected(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	roomsBCast(t, g, me, roomsBCard(me.ID, expLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 1)
	roomsBSettle(t, g, me)
	attackWith(t, g, opp.ID, a)
	roomsBSettle(t, g, me)
	if findBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne] != 1 {
		t.Error("the connecting creature got no counter")
	}
	if findBattlefieldCardForTest(g, b).Counters[game.CounterPlusOne] != 0 {
		t.Error("a creature that didn't connect got a counter")
	}
}

func TestStaffRoomIsInertWhileLocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	// Experimental Lab, the unlocked door, manifests dread and asks which
	// of the top two to manifest.
	advanceToMain(t, g)
	top := stackTopLibrary(me, "Over", "Under")[0]
	roomsBCast(t, g, me, roomsBCard(me.ID, expLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 0)
	roomsBSettle(t, g, me, top)
	attackWith(t, g, opp.ID, a)
	roomsBSettle(t, g, me)
	if findBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne] != 0 {
		t.Error("a locked Staff Room put a counter")
	}
}

func TestCharredFoyerExilesTheTopCardAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 1)
	me := g.Seats[1]
	roomsBCast(t, g, me, roomsBCard(me.ID, charredFoyerOracle, "Charred Foyer", "{3}{R}", "Warped Space", "{4}{R}{R}"), 0)
	roomsBSettle(t, g, me)
	topCard := game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(topCard)
	top := topCard.InstanceID
	for i := 0; i < 300 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	roomsBSettle(t, g, me)
	if me.Library.Contains(top) {
		t.Fatal("the top card stayed in the library")
	}
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", z)
	}
}

func TestCrampedVentsGainsLifeEqualToTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	advanceToMain(t, g)
	roomsBCast(t, g, me, roomsBCard(me.ID, crampedVentsOracle, "Cramped Vents", "{3}{B}", "Access Maze", "{5}{B}{B}"), 0)
	roomsBSettle(t, g, me, bear)
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("6 damage left the 2/2 on the battlefield")
	}
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d (6 damage on a 2-toughness creature is 4 excess)", me.Life, life+4)
	}
}

func TestPropRoomUntapsOnlyCreaturesDuringOtherPlayersUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	roomsBCast(t, g, me, roomsBCard(me.ID, dazzlingOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}"), 1)
	roomsBSettle(t, g, me)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	tapForTest(t, g, bear, rock)
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if b16Tapped(t, g, bear) {
		t.Error("Prop Room did not untap the creature")
	}
	if !b16Tapped(t, g, rock) {
		t.Error("Prop Room untapped a noncreature permanent")
	}
}

func TestPropRoomIsInertWhileLocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	roomsBCast(t, g, me, roomsBCard(me.ID, dazzlingOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}"), 0)
	roomsBSettle(t, g, me)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, bear)
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if !b16Tapped(t, g, bear) {
		t.Error("a locked Prop Room untapped a creature")
	}
}

func TestSecretArcadeMakesYourNonlandPermanentsEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "My Forest", "Land")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	roomsBCast(t, g, me, roomsBCard(me.ID, secretArcadeOracle, "Secret Arcade", "{4}{W}", "Dusty Parlor", "{2}{W}"), 0)
	roomsBSettle(t, g, me)
	has := func(id uuid.UUID) bool {
		for _, ty := range effectiveTypes(t, g, id) {
			if ty == "Enchantment" {
				return true
			}
		}
		return false
	}
	if !has(mine) {
		t.Error("your creature is not an enchantment")
	}
	if has(land) {
		t.Error("a land became an enchantment")
	}
	if has(theirs) {
		t.Error("an opponent's creature became an enchantment")
	}
}

func TestDustyParlorPutsCountersEqualToTheSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	roomsBCast(t, g, me, roomsBCard(me.ID, secretArcadeOracle, "Secret Arcade", "{4}{W}", "Dusty Parlor", "{2}{W}"), 1)
	roomsBSettle(t, g, me)
	spell := game.Card{
		InstanceID: uuid.New(), Name: "Pricey Aura", TypeLine: "Enchantment", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	}
	me.Hand.PushTop(spell)
	pushCatalogPermanent(g, me.ID, "Wastes A", "Land", "", false)
	pushCatalogPermanent(g, me.ID, "Wastes B", "Land", "", false)
	pushCatalogPermanent(g, me.ID, "Wastes C", "Land", "", false)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cannot pay for the spell in this fixture: %v", err)
	}
	roomsBSettle(t, g, me, bear)
	if n := findBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; n != 3 {
		t.Fatalf("%d counters, want 3 (the spell's mana value)", n)
	}
}
