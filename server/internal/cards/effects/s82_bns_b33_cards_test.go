package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsJungleDelverOracle     = "2f299e69-d82b-48c9-b7f5-cf07cd9dced2"
	bnsKapshoKitefinsOracle   = "b9b205db-068e-44c7-9169-5d8f13d4b73a"
	bnsKazanduNectarpotOracle = "263b526c-8b81-44ee-a7c4-5c24bf9b42a4"
	bnsKazanduStomperOracle   = "121f6c96-2911-4ee3-b14e-d3b143eb14c3"
	bnsKeldonRaiderOracle     = "004f82f0-74cc-4bc7-ab22-ab1a5247bc29"
	bnsKilnFiendOracle        = "eac8c196-8477-4b79-9875-21afa1e61708"
	bnsKomasFaithfulOracle    = "1d705d77-4b1a-4a17-b3b7-e502f2f6f6e6"
	bnsKrosanTuskerOracle     = "82406175-63b3-4972-be8b-db0e70c908cb"
	bnsKrovikanMistOracle     = "6cae42b6-8b09-4188-8e79-203eebc77585"
	bnsKruinStrikerOracle     = "3e261fa8-1256-4525-8115-5880758338fe"
	bnsKujarSeedsculptorOracl = "93f817c7-2a3d-4346-94c1-3ba8bafb485d"
)

func TestBnsB33EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsJungleDelverOracle:     "Jungle Delver",
		bnsKapshoKitefinsOracle:   "Kapsho Kitefins",
		bnsKazanduNectarpotOracle: "Kazandu Nectarpot",
		bnsKazanduStomperOracle:   "Kazandu Stomper",
		bnsKeldonRaiderOracle:     "Keldon Raider",
		bnsKilnFiendOracle:        "Kiln Fiend",
		bnsKomasFaithfulOracle:    "Koma's Faithful",
		bnsKrosanTuskerOracle:     "Krosan Tusker",
		bnsKrovikanMistOracle:     "Krovikan Mist",
		bnsKruinStrikerOracle:     "Kruin Striker",
		bnsKujarSeedsculptorOracl: "Kujar Seedsculptor",
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

func TestBnsB33JungleDelverGrowsForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	delver := b12Push(g, me.ID, "Jungle Delver", "Creature — Merfolk Warrior", bnsJungleDelverOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, delver, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{G} was activated with an empty pool")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, delver, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, delver, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if p, tough := layeredCard(t, g, delver).CurrentPower(), layeredCard(t, g, delver).CurrentToughness(); p != 2 || tough != 2 {
		t.Errorf("Delver is %d/%d, want 2/2", p, tough)
	}
}

func TestBnsB33KapshoKitefinsTapsOnItsOwnEntryAndOnAnotherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fins := castAndResolveCreature(t, g, "Kapsho Kitefins", "Creature — Fish", bnsKapshoKitefinsOracle)
	if !hasEffectiveKeyword(t, g, fins, "flying") {
		t.Error("no flying")
	}
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("a creature its controller controls was a legal target")
	}
	answerPickTargetIfOpen(t, g, ogre)
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, ogre).Tapped {
		t.Error("the Ogre was not tapped when the Kitefins entered")
	}
	// Another creature entering taps again.
	ogre2 := b12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 3)
	castAndResolveCreature(t, g, "Grizzly", "Creature — Bear", "")
	answerPickTargetIfOpen(t, g, ogre2)
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, ogre2).Tapped {
		t.Error("the Troll was not tapped when another creature entered")
	}
}

func TestBnsB33KazanduNectarpotGainsLifeOnYourLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Kazandu Nectarpot", "Creature — Insect", bnsKazanduNectarpotOracle, 1, 3)
	advanceToMain(t, g)
	life := me.Life
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	life = me.Life
	b13PlayAs(t, g, 1, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's land changed your life %d -> %d", life, me.Life)
	}
}

func TestBnsB33KazanduStomperReturnsUpToTwoLands(t *testing.T) {
	for _, n := range []int{0, 1, 2} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		var lands []uuid.UUID
		for i := 0; i < 3; i++ {
			lands = append(lands, mgbLand(g, me.ID, "Forest", "Basic Land — Forest"))
		}
		stomper := castAndResolveCreature(t, g, "Kazandu Stomper", "Creature — Beast", bnsKazanduStomperOracle)
		passPriorityAroundTable(t, g)
		pick := latestChooseCardsFor(g, me.ID)
		if pick == nil {
			t.Fatal("no prompt for the lands")
		}
		if slices.Contains(pick.ChooseCards, stomper) {
			t.Error("the Stomper was offered as a land")
		}
		if err := g.ResolveChooseCards(pick.ID, me.ID, lands); err == nil {
			t.Fatal("three lands were accepted")
		}
		if err := g.ResolveChooseCards(pick.ID, me.ID, lands[:n]); err != nil {
			t.Fatalf("ResolveChooseCards(%d): %v", n, err)
		}
		passPriorityAroundTable(t, g)
		for i, l := range lands {
			if want := i < n; me.Hand.Contains(l) != want {
				t.Errorf("n=%d land %d in hand = %v, want %v", n, i, !want, want)
			}
		}
		if !hasEffectiveKeyword(t, g, stomper, "trample") {
			t.Error("no trample")
		}
	}
}

func TestBnsB33KeldonRaiderLootsOnlyIfYouDiscard(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pitch := pushCatalogHandCard(me, "Pitch", "Sorcery", "")
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		hand, graves := me.Hand.Size(), me.Graveyard.Size()
		raider := castAndResolveCreature(t, g, "Keldon Raider", "Creature — Human Warrior", bnsKeldonRaiderOracle)
		passPriorityAroundTable(t, g)
		if discard {
			answerDiscard(t, g, me.ID, pitch)
		} else {
			answerDiscard(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		// The Raider left the hand when it was cast.
		if got := me.Hand.Size(); got != hand {
			t.Errorf("discard=%v: hand %d, started with %d", discard, got, hand)
		}
		if want := graves + boolInt(discard); me.Graveyard.Size() != want {
			t.Errorf("discard=%v: graveyard %d, want %d", discard, me.Graveyard.Size(), want)
		}
		if !g.Battlefield.Contains(raider) {
			t.Error("the Raider is not on the battlefield")
		}
	}
}

func TestBnsB33KilnFiendPumpsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fiend := b12Push(g, me.ID, "Kiln Fiend", "Creature — Elemental Beast", bnsKilnFiendOracle, 1, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Think", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, fiend); p != 4 {
		t.Errorf("power %d after a sorcery, want 4", p)
	}
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, fiend); p != 7 {
		t.Errorf("power %d after two spells, want 7", p)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, fiend); p != 7 || tough != 2 {
		t.Errorf("a creature spell moved the Fiend to %d/%d", p, tough)
	}
}

func TestBnsB33KomasFaithfulMillsEachPlayerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, s := range g.Seats {
		for i := 0; i < 5; i++ {
			pushLibraryCardForTest(s, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: s.ID, Controller: s.ID})
		}
	}
	faithful := b12Push(g, me.ID, "Koma's Faithful", "Creature — Elf Cleric", bnsKomasFaithfulOracle, 3, 1)
	if !hasEffectiveKeyword(t, g, faithful, "lifelink") {
		t.Error("no lifelink")
	}
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Graveyard.Size()
	}
	killCreature(t, g, me.ID, faithful)
	passPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		want := before[i] + 3
		if s == me {
			want++
		}
		if s.Graveyard.Size() != want {
			t.Errorf("seat %d graveyard %d, want %d", i, s.Graveyard.Size(), want)
		}
	}
}

func TestBnsB33KrosanTuskerCyclingFetchesABasicLandBeforeTheDraw(t *testing.T) {
	for _, take := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		forest := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest", Owner: me.ID, Controller: me.ID})
		id, _ := cycleFromHand(t, g, "Krosan Tusker", "Creature — Boar Beast", bnsKrosanTuskerOracle, "{C}{C}{G}")
		if !me.Graveyard.Contains(id) {
			t.Fatal("the cycled Tusker is not in the graveyard")
		}
		handBefore := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if take {
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if searchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil {
				t.Error("a nonbasic land was offered")
			}
			answerSearchByID(t, g, me.ID, forest)
		} else {
			answerSearchFailToFind(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		want := handBefore + 1 // the draw
		if take {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("take=%v: hand %d -> %d, want %d", take, handBefore, got, want)
		}
		if take && !me.Hand.Contains(forest) {
			t.Error("the Forest is not in hand")
		}
	}
}

func TestBnsB33KrovikanMistCountsEveryIllusion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mist := b12Push(g, me.ID, "Krovikan Mist", "Creature — Illusion", bnsKrovikanMistOracle, 0, 0)
	if p, tough := effectivePTOf(t, g, mist); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", p, tough)
	}
	b12Creature(g, opp.ID, "Phantom", "Creature — Illusion", 2, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectivePTOf(t, g, mist); p != 2 || tough != 2 {
		t.Errorf("with another Illusion it is %d/%d, want 2/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, mist, "flying") {
		t.Error("no flying")
	}
}

func TestBnsB33KruinStrikerGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	striker := b12Push(g, me.ID, "Kruin Striker", "Creature — Human Warrior", bnsKruinStrikerOracle, 2, 1)
	if hasEffectiveKeyword(t, g, striker, "trample") {
		t.Error("trample before any creature entered")
	}
	castAndResolveCreature(t, g, "Grizzly", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, striker); p != 3 || tough != 1 {
		t.Errorf("Striker is %d/%d, want 3/1", p, tough)
	}
	if !hasEffectiveKeyword(t, g, striker, "trample") {
		t.Error("no trample")
	}
}

func TestBnsB33KujarSeedsculptorPutsACounterOnATargetYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castAndResolveCreature(t, g, "Kujar Seedsculptor", "Creature — Elf Druid", bnsKujarSeedsculptorOracl)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ogre}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters on the Bear, want 1", got)
	}
	if got := counterOn(g, ogre, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters on the Ogre, want 0", got)
	}
}
