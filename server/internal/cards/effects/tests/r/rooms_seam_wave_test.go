package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The seam wave's Rooms are all complete now. Moldering Gym, Ticket
// Booth and Underwater Tunnel shipped their manifest-dread doors with
// #2570; Experimental Lab // Staff Room turns a creature face up since
// #2590, and Dazzling Theater // Prop Room gives creature spells convoke
// since #2137. Charred Foyer // Warped Space and Cramped Vents // Access
// Maze completed with #3174's door-gated granted alternative costs, and
// Secret Arcade // Dusty Parlor with #3862's permanent spells that are
// enchantments on the stack.

func TestMolderingGymFetchesABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Nonbasic", TypeLine: "Land"})
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.MoldGymOracle, "Moldering Gym", "{2}{G}", "Weight Room", "{5}{G}"), 0)
	effectstest.RoomsBSettle(t, g, me, forest)
	effectstest.TakeSearchPrompts(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, forest)
	if c == nil || !c.Tapped {
		t.Fatalf("the basic land did not enter tapped: %+v", c)
	}
	if effectstest.FindBattlefieldCardForTest(g, other) != nil {
		t.Error("a nonbasic land was fetched")
	}
}

func TestTunnelOfHateGivesTheChosenAttackerDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 2, 2)
	c := effectstest.RoomsBCard(me.ID, effectstest.TicketBoothOracle, "Ticket Booth", "{2}{R}", "Tunnel of Hate", "{4}{R}{R}")
	effectstest.RoomsBCast(t, g, me, c, 1)
	effectstest.RoomsBSettle(t, g, me)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.RoomsBSettle(t, g, me, b)
	if effectstest.EffectiveAbilitiesContain(t, g, a, "double strike") || !effectstest.EffectiveAbilitiesContain(t, g, b, "double strike") {
		t.Fatal("double strike should go to the chosen attacker only")
	}
}

func TestUnderwaterTunnelSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Sorcery"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: "Sorcery"})
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.UnderwaterOracle, "Underwater Tunnel", "{U}", "Slimy Aquarium", "{3}{U}"), 0)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.ExpLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 1)
	effectstest.RoomsBSettle(t, g, me)
	effectstest.AttackWith(t, g, opp.ID, a)
	effectstest.RoomsBSettle(t, g, me)
	if effectstest.FindBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne] != 1 {
		t.Error("the connecting creature got no counter")
	}
	if effectstest.FindBattlefieldCardForTest(g, b).Counters[game.CounterPlusOne] != 0 {
		t.Error("a creature that didn't connect got a counter")
	}
}

func TestStaffRoomIsInertWhileLocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	// Experimental Lab, the unlocked door, manifests dread and asks which
	// of the top two to manifest.
	effectstest.AdvanceToMain(t, g)
	top := effectstest.StackTopLibrary(me, "Over", "Under")[0]
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.ExpLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 0)
	effectstest.RoomsBSettle(t, g, me, top)
	effectstest.AttackWith(t, g, opp.ID, a)
	effectstest.RoomsBSettle(t, g, me)
	if effectstest.FindBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne] != 0 {
		t.Error("a locked Staff Room put a counter")
	}
}

func TestCharredFoyerExilesTheTopCardAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMainOf(t, g, 1)
	me := g.Seats[1]
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.CharredFoyerOracle, "Charred Foyer", "{3}{R}", "Warped Space", "{4}{R}{R}"), 0)
	effectstest.RoomsBSettle(t, g, me)
	topCard := game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(topCard)
	top := topCard.InstanceID
	for i := 0; i < 300 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.RoomsBSettle(t, g, me)
	if me.Library.Contains(top) {
		t.Fatal("the top card stayed in the library")
	}
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", z)
	}
}

func TestCrampedVentsGainsLifeEqualToTheExcess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.CrampedVentsOracle, "Cramped Vents", "{3}{B}", "Access Maze", "{5}{B}{B}"), 0)
	effectstest.RoomsBSettle(t, g, me, bear)
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("6 damage left the 2/2 on the battlefield")
	}
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d (6 damage on a 2-toughness creature is 4 excess)", me.Life, life+4)
	}
}

func TestPropRoomUntapsOnlyCreaturesDuringOtherPlayersUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.DazzlingOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}"), 1)
	effectstest.RoomsBSettle(t, g, me)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.TapForTest(t, g, bear, rock)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("Prop Room did not untap the creature")
	}
	if !effectstest.B16Tapped(t, g, rock) {
		t.Error("Prop Room untapped a noncreature permanent")
	}
}

func TestPropRoomIsInertWhileLocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.DazzlingOracle, "Dazzling Theater", "{3}{W}", "Prop Room", "{2}{W}"), 0)
	effectstest.RoomsBSettle(t, g, me)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("a locked Prop Room untapped a creature")
	}
}

func TestSecretArcadeMakesYourNonlandPermanentsEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "My Forest", "Land")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.SecretArcadeOracle, "Secret Arcade", "{4}{W}", "Dusty Parlor", "{2}{W}"), 0)
	effectstest.RoomsBSettle(t, g, me)
	has := func(id uuid.UUID) bool {
		for _, ty := range effectstest.EffectiveTypes(t, g, id) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.SecretArcadeOracle, "Secret Arcade", "{4}{W}", "Dusty Parlor", "{2}{W}"), 1)
	effectstest.RoomsBSettle(t, g, me)
	spell := game.Card{
		InstanceID: uuid.New(), Name: "Pricey Aura", TypeLine: "Enchantment", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	}
	me.Hand.PushTop(spell)
	effectstest.PushCatalogPermanent(g, me.ID, "Wastes A", "Land", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Wastes B", "Land", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Wastes C", "Land", "", false)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cannot pay for the spell in this fixture: %v", err)
	}
	effectstest.RoomsBSettle(t, g, me, bear)
	if n := effectstest.FindBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; n != 3 {
		t.Fatalf("%d counters, want 3 (the spell's mana value)", n)
	}
}
