package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b5_cards_test.go — Final Fantasy (FIN), batch 5 (#4462). Each test
// plays the real catalog entry and asserts what its text says.

const (
	finB5LoporritOracle   = "24630190-a630-4dd3-a8f7-82f129c881e3"
	finB5ReachOracle      = "14bc29ee-4f84-4e62-8ede-5741b99414c9"
	finB5ShoopufOracle    = "81b67745-0b1f-4433-963d-9b9653fb4d95"
	finB5RydiasReturnOrc  = "0d007479-c5e8-4aa1-b53e-fa962783deee"
	finB5SidequestOracle  = "aebeaf0b-4df1-4c34-919b-740bfab19e56"
	finB5GrimoireOracle   = "d4453b3e-733e-4038-81da-135a6a03079c"
	finB5TorgalOracle     = "9d6ab594-aea9-4f13-8802-f3794ee45425"
	finB5CloudOracle      = "525f57df-7d4d-4089-9003-4b01005558dc"
	finB5ExdeathOracle    = "98c9d798-8c48-4a2d-a076-ce17d31df6f5"
	finB5GiottOracle      = "c629fa91-e13c-494b-a95f-a91b7c52035a"
	finB5ForestOracleNone = ""
)

func TestFinB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB5LoporritOracle:         "Loporrit Scout",
		finB5ReachOracle:            "Reach the Horizon",
		finB5ShoopufOracle:          "Ride the Shoopuf",
		finB5RydiasReturnOrc:        "Rydia's Return",
		finB5SidequestOracle:        "Sidequest: Raise a Chocobo",
		finB5SidequestOracle + "#1": "Black Chocobo",
		finB5GrimoireOracle:         "Summoner's Grimoire",
		finB5TorgalOracle:           "Torgal, A Fine Hound",
		finB5CloudOracle:            "Cloud of Darkness",
		finB5ExdeathOracle:          "Exdeath, Void Warlock",
		finB5ExdeathOracle + "#1":   "Neo Exdeath, Dimension's End",
		finB5GiottOracle:            "Giott, King of the Dwarves",
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

// --- Loporrit Scout ------------------------------------------------------------------

func TestFinB5LoporritScoutGrowsOncePerOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	scout := effectstest.B12Push(g, me.ID, "Loporrit Scout", "Creature — Rabbit Scout", finB5LoporritOracle, 3, 2)
	effectstest.TdCastCreature(t, g, "Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, scout); p != 4 || th != 3 {
		t.Fatalf("after one creature the Scout is %d/%d, want 4/3", p, th)
	}
	effectstest.TdCastCreature(t, g, "Cub", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, scout); p != 5 || th != 4 {
		t.Errorf("after two creatures the Scout is %d/%d, want 5/4", p, th)
	}
	// An opponent's creature, and the Scout itself entering, do not count.
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, scout); p != 5 {
		t.Errorf("an opponent's creature changed the Scout to power %d", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, scout); p != 3 || th != 2 {
		t.Errorf("next turn the Scout is %d/%d, want 3/2", p, th)
	}
}

func TestFinB5LoporritScoutDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.TdCastFromHand(t, g, game.Card{Name: "Loporrit Scout", TypeLine: "Creature — Rabbit Scout",
		OracleID: finB5LoporritOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, id); p != 3 || th != 2 {
		t.Errorf("the Scout is %d/%d after entering, want 3/2", p, th)
	}
	_ = me
}

// --- Reach the Horizon ---------------------------------------------------------------

func TestFinB5ReachTheHorizonFetchesTwoDifferentlyNamedLandsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	f1, f2, plains, town := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: f1, Name: "Forest", TypeLine: "Basic Land — Forest"},
		{InstanceID: f2, Name: "Forest", TypeLine: "Basic Land — Forest"},
		{InstanceID: plains, Name: "Plains", TypeLine: "Basic Land — Plains"},
		{InstanceID: town, Name: "Gilded Town", TypeLine: "Land — Town"},
		{InstanceID: uuid.New(), Name: "Grove", TypeLine: "Land — Forest"},
		{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	effectstest.CastCatalogSpell(t, g, "Reach the Horizon", "Sorcery", finB5ReachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Grove") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land without the Town type, or a nonland, was offered")
	}
	// Two Forests are not a legal pair.
	c = effectstest.SearchChoiceFor(g, me.ID)
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{f1, f2}); err == nil {
		t.Fatal("two cards with the same name were accepted")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, f1, town)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{f1, town} {
		if !g.Battlefield.Contains(id) || !effectstest.IsTapped(g, id) {
			t.Errorf("%s did not enter tapped", id)
		}
	}
	if g.Battlefield.Contains(plains) || g.Battlefield.Contains(f2) {
		t.Error("more than two cards were fetched")
	}
}

func TestFinB5ReachTheHorizonMayFetchFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: plains, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Reach the Horizon", "Sorcery", finB5ReachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plains) {
		t.Error("the lone Plains did not enter")
	}
}

// --- Rydia's Return ------------------------------------------------------------------

func TestFinB5RydiasReturnPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Rydia's Return", "Sorcery", finB5RydiasReturnOrc, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, a); p != 5 || th != 5 {
		t.Errorf("A is %d/%d, want 5/5", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, b); p != 4 || th != 4 {
		t.Errorf("B is %d/%d, want 4/4", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, th)
	}
}

func TestFinB5RydiasReturnReturnsUpToTwoPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardCardWithTypeLineID(me, "Rock", "Artifact")
	bolt := effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Rydia's Return", "Sorcery", finB5RydiasReturnOrc, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(rock) {
		t.Error("the targeted cards are not in hand")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the instant left the graveyard")
	}
}

func TestFinB5RydiasReturnRefusesAnInstantCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	if _, err := effectstest.CastModalErr(t, g, "Rydia's Return", "Sorcery", finB5RydiasReturnOrc, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Error("an instant card was accepted as a permanent card")
	}
}

// --- Cloud of Darkness ---------------------------------------------------------------

func TestFinB5CloudOfDarknessShrinksByPermanentCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(me, "Bear Card", "Creature — Bear")
	}
	effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Cloud of Darkness", TypeLine: "Legendary Creature — Avatar",
		OracleID: finB5CloudOracle, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("-3/-3 did not kill the 3/3 (three permanent cards, the instant not counted)")
	}
}

func TestFinB5CloudOfDarknessCountsOnceAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.PushGraveyardCardWithTypeLineID(me, "Bear Card", "Creature — Bear")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Cloud of Darkness", TypeLine: "Legendary Creature — Avatar",
		OracleID: finB5CloudOracle, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, victim); p != 2 || th != 2 {
		t.Errorf("the victim is %d/%d, want 2/2", p, th)
	}
	effectstest.PushGraveyardCardWithTypeLineID(me, "Another", "Creature — Bear")
	if p, th := effectstest.EffectivePTOf(t, g, victim); p != 2 || th != 2 {
		t.Errorf("the shrink changed when the graveyard did: %d/%d", p, th)
	}
}

// --- Giott, King of the Dwarves ------------------------------------------------------

func TestFinB5GiottRummagesForADwarfAndForAnEquipment(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"Dwarf", "Creature — Dwarf Warrior", true},
		{"Equipment", "Artifact — Equipment", true},
		{"Bear", "Creature — Bear", false},
		{"Aura", "Enchantment — Aura", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B12Push(g, me.ID, "Giott, King of the Dwarves", "Legendary Creature — Dwarf Noble", finB5GiottOracle, 1, 1)
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			effectstest.TdCastFromHand(t, g, game.Card{Name: tc.name, TypeLine: tc.typeLine, Power: 1, Toughness: 1}, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DiscardChoiceFor(g, me.ID) != nil; got != tc.want {
				t.Fatalf("rummage offered = %v, want %v", got, tc.want)
			}
			if !tc.want {
				return
			}
			hand := me.Hand.Size()
			effectstest.DiscardFromHand(t, g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand {
				t.Errorf("hand %d → %d, want a discard and a draw", hand, me.Hand.Size())
			}
		})
	}
}

func TestFinB5GiottTriggersForItselfAndNotForAnOpponentsDwarf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Giott, King of the Dwarves", TypeLine: "Legendary Creature — Dwarf Noble",
		OracleID: finB5GiottOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("Giott entering did not offer the rummage")
	}
	hand := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d", hand, me.Hand.Size())
	}
	effectstest.B12Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("an opponent's Dwarf entering offered the rummage")
	}
}

// --- Ride the Shoopuf ----------------------------------------------------------------

func TestFinB5RideTheShoopufPutsACounterOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Ride the Shoopuf", "Enchantment", finB5ShoopufOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", finB5ForestOracleNone)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
}

func TestFinB5RideTheShoopufBecomesASevenSevenBeastThatStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shoopuf := effectstest.B12Push(g, me.ID, "Ride the Shoopuf", "Enchantment", finB5ShoopufOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "G", "G", "G", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, shoopuf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	view := effectstest.S58p6View(t, g, shoopuf)
	if !view.IsCreature() || !view.HasCardType("enchantment") || !view.HasSubtype("Beast") {
		t.Fatalf("Ride the Shoopuf is %v, want an enchantment Beast creature", view.Effective().Types)
	}
	if p, th := effectstest.EffectivePTOf(t, g, shoopuf); p != 7 || th != 7 {
		t.Errorf("it is %d/%d, want 7/7", p, th)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, shoopuf); p != 7 || th != 7 {
		t.Errorf("a turn later it is %d/%d, want 7/7 (no duration)", p, th)
	}
}

// --- Torgal, A Fine Hound ------------------------------------------------------------

func finB5Torgal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Torgal, A Fine Hound", "Legendary Creature — Wolf", finB5TorgalOracle, 2, 2)
}

func TestFinB5TorgalTapsForAnyColor(t *testing.T) {
	spec, _ := effects.Lookup(finB5TorgalOracle)
	if len(spec.ManaAbilities) != 1 || !spec.ManaAbilities[0].Cost.Tap || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("mana abilities %+v, want {T}: any one color", spec.ManaAbilities)
	}
}

func TestFinB5TorgalGivesTheFirstHumanACounterPerDogAndWolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB5Torgal(g, me.ID)
	effectstest.B12Creature(g, me.ID, "Hound", "Creature — Dog", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := effectstest.TdCastFromHand(t, g, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, human, game.CounterPlusOne); got != 2 {
		t.Errorf("the first Human has %d +1/+1 counters, want 2 (Torgal and the Dog)", got)
	}
	second := effectstest.TdCastFromHand(t, g, game.Card{Name: "Cleric", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("the second Human has %d counters, want 0 (only the first each turn)", got)
	}
}

func TestFinB5TorgalIgnoresANonHumanFirstSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB5Torgal(g, me.ID)
	bear := effectstest.TdCastFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a Bear got %d counters", got)
	}
	// A non-Human creature spell is not "your first Human creature spell":
	// the Human cast after it still is.
	human := effectstest.TdCastFromHand(t, g, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, human, game.CounterPlusOne); got != 1 {
		t.Errorf("the first Human after a Bear has %d counters, want 1 (Torgal)", got)
	}
}

func TestFinB5TorgalCountsTheDogsAndWolvesAsTheCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB5Torgal(g, me.ID)
	human := effectstest.TdCastFromHand(t, g, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	// Torgal's trigger resolves first; a Wolf arrives before the spell does.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.B12Creature(g, me.ID, "Late Wolf", "Creature — Wolf", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, human, game.CounterPlusOne); got != 2 {
		t.Errorf("the Human has %d counters, want 2 (Torgal and the Wolf that arrived in between)", got)
	}
}

func TestFinB5TorgalDoesNotWatchAnOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	finB5Torgal(g, opp.ID)
	human := effectstest.TdCastFromHand(t, g, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, human, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Torgal gave %s's Human %d counters", me.Name, got)
	}
}

// --- Sidequest: Raise a Chocobo // Black Chocobo -------------------------------------

func finB5SidequestRow() cards.Card {
	return effectstest.TransformRow(finB5SidequestOracle, "Sidequest: Raise a Chocobo", "Enchantment", "{1}{G}",
		"Black Chocobo", "Creature — Bird", "2", "2", []string{"G"})
}

func finB5Birds(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Creature(g, owner, "Bird", "Creature — Bird", 1, 1)
	}
}

func TestFinB5SidequestMakesAChocoboBirdWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.ImportToBattlefield(t, g, finB5SidequestRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Bird", me.ID); got != 1 {
		t.Fatalf("%d Birds after it entered, want 1", got)
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if p, th := effectstest.EffectivePTOf(t, g, bird); p != 2 || th != 2 {
		t.Errorf("the Bird is %d/%d, want 2/2", p, th)
	}
}

func TestFinB5SidequestTransformsWithFourBirdsAndFetchesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	sq := effectstest.ImportToBattlefield(t, g, finB5SidequestRow(), me)
	effectstest.PassPriorityAroundTable(t, g) // its Bird
	finB5Birds(g, me.ID, 3)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	// The draw step comes first: keep the land out of that draw.
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, sq); c == nil || c.ActiveFace != 1 {
		t.Fatalf("the Sidequest did not transform with four Birds (face %v)", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, land)
	effectstest.MonSettle(t, g) // the land's landfall: the Bird token's and Black Chocobo's are ordered
	if !g.Battlefield.Contains(land) || !effectstest.IsTapped(g, land) {
		t.Error("Black Chocobo's search did not put a land onto the battlefield tapped")
	}
	// Landfall: Birds get +1/+0, including Black Chocobo itself.
	chocobo := effectstest.S58p6View(t, g, sq)
	if !chocobo.IsCreature() || !chocobo.HasSubtype("Bird") {
		t.Fatalf("Black Chocobo is %v, want a Bird creature", chocobo.Effective().Types)
	}
	before, _ := effectstest.EffectivePTOf(t, g, sq)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.MonSettle(t, g) // the Bird token's landfall and Black Chocobo's are ordered
	if after, _ := effectstest.EffectivePTOf(t, g, sq); after != before+1 {
		t.Errorf("Black Chocobo's power %d → %d after landfall, want +1", before, after)
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if p, _ := effectstest.EffectivePTOf(t, g, bird); p < 2 {
		t.Errorf("a Bird has power %d after landfall, want it pumped", p)
	}
	opp := effectstest.FinABOpp(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.MonSettle(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Bird got +1/+0 (power %d)", p)
	}
}

func TestFinB5SidequestDoesNotTransformWithThreeBirds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	sq := effectstest.ImportToBattlefield(t, g, finB5SidequestRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	finB5Birds(g, me.ID, 2) // three in all
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, sq); c == nil || c.ActiveFace != 0 {
		t.Error("the Sidequest transformed with only three Birds")
	}
}

func TestFinB5SidequestRechecksItsConditionOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	sq := effectstest.ImportToBattlefield(t, g, finB5SidequestRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	finB5Birds(g, me.ID, 3)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
		t.Fatal("the transform trigger did not go on the stack")
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bird); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, sq); c == nil || c.ActiveFace != 0 {
		t.Error("the Sidequest transformed although a Bird left in response")
	}
}

// --- Exdeath, Void Warlock // Neo Exdeath, Dimension's End ---------------------------

func finB5ExdeathRow() cards.Card {
	row := effectstest.TransformRow(finB5ExdeathOracle, "Exdeath, Void Warlock", "Legendary Creature — Spirit Warlock", "{1}{B}{G}",
		"Neo Exdeath, Dimension's End", "Legendary Creature — Spirit Avatar", "*", "3", []string{"B", "G"})
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "3", "3"
	return row
}

func finB5Graveyard(p *game.Player, permanents int) {
	for i := 0; i < permanents; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(p, "Bear Card", "Creature — Bear")
	}
}

func TestFinB5ExdeathGainsThreeLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.ImportToBattlefield(t, g, finB5ExdeathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestFinB5ExdeathTransformsAtYourEndStepWithSixPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	ex := effectstest.ImportToBattlefield(t, g, finB5ExdeathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	finB5Graveyard(me, 5)
	effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant") // not a permanent card
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, ex); c == nil || c.ActiveFace != 0 {
		t.Fatal("Exdeath transformed with only five permanent cards (an instant is not one)")
	}
	finB5Graveyard(me, 1)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, ex); c == nil || c.ActiveFace != 1 {
		t.Fatalf("Exdeath did not transform with six permanent cards")
	}
	p, th := effectstest.EffectivePTOf(t, g, ex)
	if want := 6; p != want || th != 3 {
		t.Errorf("Neo Exdeath is %d/%d, want %d/3 (its power is the permanent cards in your graveyard)", p, th, want)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ex, "trample") {
		t.Error("Neo Exdeath lacks trample")
	}
	for i := 0; i < 2; i++ {
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, fodder); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, ex); p != 8 {
		t.Errorf("after two more cards Neo Exdeath has power %d, want 8", p)
	}
}

func TestFinB5ExdeathDoesNotTransformOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	ex := effectstest.ImportToBattlefield(t, g, finB5ExdeathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g) // past your own end step
	finB5Graveyard(me, 6)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, ex); c == nil || c.ActiveFace != 0 {
		t.Error("Exdeath transformed on an opponent's end step")
	}
}

// --- Summoner's Grimoire -------------------------------------------------------------

func finB5Grimoire(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Summoner's Grimoire", "Artifact — Book Equipment", finB5GrimoireOracle, 0, 0)
}

func TestFinB5GrimoireJobSelectMakesAShamanHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grim := effectstest.TdCastFromHand(t, g, game.Card{Name: "Summoner's Grimoire", TypeLine: "Artifact — Book Equipment", OracleID: finB5GrimoireOracle},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.FindBattlefieldByName(g, "Hero")
	if hero == uuid.Nil {
		t.Fatal("job select made no Hero")
	}
	if c := effectstest.FindBattlefieldCardByID(g, grim); c == nil || c.AttachedTo.ID != hero {
		t.Error("the Grimoire is not attached to the Hero")
	}
	v := effectstest.S58p6View(t, g, hero)
	if !v.HasSubtype("Shaman") || !v.HasSubtype("Hero") {
		t.Errorf("the equipped Hero is %v, want a Hero Shaman", v.Effective().Subtypes)
	}
	_ = me
}

// finB5Equip equips the Grimoire to a creature for {3}.
func finB5Equip(t *testing.T, g *game.Game, me *game.Player, grim, creature uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, grim, 0, game.ActivateAbilityParams{Targets: effectstest.TdCard(creature)})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, grim); c == nil || c.AttachedTo.ID != creature {
		t.Fatal("equip did not attach the Grimoire")
	}
}

func finB5Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
}

func TestFinB5GrimoirePutsACreatureFromHandWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	grim := finB5Grimoire(g, me.ID)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	finB5Equip(t, g, me, grim, knight)
	fat := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fat, Name: "Fatty", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	finB5Attack(t, g, knight, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature card")
	}
	for _, id := range pick.ChooseCards {
		if id == spell {
			t.Error("a noncreature card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, fat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fat) {
		t.Fatal("the creature card did not enter")
	}
	if c := effectstest.FindBattlefieldCardByID(g, fat); c == nil || c.Tapped || c.AttackingTarget != uuid.Nil {
		t.Error("a non-enchantment creature entered tapped or attacking")
	}
}

func TestFinB5GrimoireEnchantmentCreatureEntersTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	grim := finB5Grimoire(g, me.ID)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	finB5Equip(t, g, me, grim, knight)
	god := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: god, Name: "Sage", TypeLine: "Enchantment Creature — Human", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	finB5Attack(t, g, knight, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, god)
	effectstest.PassPriorityAroundTable(t, g)
	// One attack target per opponent: the controller names which.
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.FindBattlefieldCardByID(g, god)
	if c == nil {
		t.Fatal("the enchantment creature did not enter")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the enchantment creature entered tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
}

func TestFinB5GrimoireMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	grim := finB5Grimoire(g, me.ID)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	finB5Equip(t, g, me, grim, knight)
	fat := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fat, Name: "Fatty", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	finB5Attack(t, g, knight, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fat) || !me.Hand.Contains(fat) {
		t.Error("declining still put the card onto the battlefield")
	}
}

func TestFinB5GrimoireGrantsNothingOnceUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	finB5Grimoire(g, me.ID)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Fatty", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	finB5Attack(t, g, knight, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an unequipped creature's attack offered the put")
	}
}
