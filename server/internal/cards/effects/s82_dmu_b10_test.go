package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82DmuArtilleryBlastOracle  = "dad0b69f-ef33-475b-aebc-3a3ca880feb1"
	s82DmuDragToTheBottomOracle = "c4d7a200-7401-4c9a-a7cf-6c98fa8eb61a"
	s82DmuHerdMigrationOracle   = "dfdc0eab-309b-4441-a94c-930c6dde4c48"
	s82DmuMeriasOutriderOracle  = "6ae9aa10-38f5-4b10-8b44-1164b85acd7f"
	s82DmuShadowProphecyOracle  = "620f3851-493d-496e-880a-4a9dee074e79"
	s82DmuSlimefootsOracle      = "8ace9425-fcdd-45b0-abf0-1a7d601de25e"
	s82DmuVodaOracle            = "81a9da9d-c422-4293-b210-d89222633b4c"
	s82DmuBortukOracle          = "d965e8c2-34b1-42eb-9647-59951d9fc06b"
)

func TestS82DmuB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82DmuArtilleryBlastOracle:  "Artillery Blast",
		s82DmuDragToTheBottomOracle: "Drag to the Bottom",
		s82DmuHerdMigrationOracle:   "Herd Migration",
		s82DmuMeriasOutriderOracle:  "Meria's Outrider",
		s82DmuShadowProphecyOracle:  "Shadow Prophecy",
		s82DmuSlimefootsOracle:      "Slimefoot's Survey",
		s82DmuVodaOracle:            "Voda Sea Scavenger",
		s82DmuBortukOracle:          "Bortuk Bonerattle",
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

// s82DmuDomain gives `p` one land of each of the first n basic land types
// (two Plains first, to show a type counts once).
func s82DmuDomain(g *game.Game, p *game.Player, n int) {
	pushLandFor(g, p.ID, "Plains", "Basic Land — Plains")
	for i, ty := range game.BasicLandTypes {
		if i >= n {
			break
		}
		pushLandFor(g, p.ID, ty, "Basic Land — "+ty)
	}
}

func TestS82DmuB10ArtilleryBlastDealsOnePlusDomainToATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82DmuDomain(g, me, 3)
	untapped := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Ogre", TypeLine: "Creature — Ogre",
		Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	if err := castCatalogSpellErr(t, g, "Artillery Blast", "Instant", s82DmuArtilleryBlastOracle, cardTarget(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Artillery Blast", "Instant", s82DmuArtilleryBlastOracle, cardTarget(tapped))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, tapped).DamageMarked; got != 4 {
		t.Errorf("damage %d, want 4 (1 plus three basic land types)", got)
	}
}

func TestS82DmuB10DragToTheBottomShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82DmuDomain(g, me, 2)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 2, 2)
	big := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castCatalogSpell(t, g, "Drag to the Bottom", "Sorcery", s82DmuDragToTheBottomOracle, nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 3 or less survived -3/-3 (1 plus two types)")
	}
	if c := layeredCard(t, g, big); c.Effective().Power != 2 || c.Effective().Toughness != 2 {
		t.Errorf("the Ogre is %d/%d, want 2/2", c.Effective().Power, c.Effective().Toughness)
	}
}

func TestS82DmuB10HerdMigrationMakesABeastPerType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82DmuDomain(g, me, 4)
	castCatalogSpell(t, g, "Herd Migration", "Sorcery", s82DmuHerdMigrationOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" && c.Controller == me.ID {
			n++
			if c.Power != 3 || c.Toughness != 3 || !c.HasColor("G") {
				t.Errorf("a Beast is %d/%d", c.Power, c.Toughness)
			}
		}
	}
	if n != 4 {
		t.Errorf("%d Beasts, want 4", n)
	}
}

func TestS82DmuB10HerdMigrationDiscardSearchesForABasicAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"), searchTestLand("Shock", "Land — Mountain Forest"))
	card := handCardForTest(me, "Herd Migration", "Sorcery", s82DmuHerdMigrationOracle)
	fillPoolColored(me, "G", 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded as a cost")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land is not in hand")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestS82DmuB10MeriasOutriderHitsEachOpponentForDomain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82DmuDomain(g, me, 3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castRingCreature(t, g, "Meria's Outrider", "Creature — Elf Archer", s82DmuMeriasOutriderOracle, 4, 4)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 0
		if p.ID != me.ID {
			want = 3
		}
		if got := before[p.ID] - p.Life; got != want {
			t.Errorf("seat %s lost %d, want %d", p.ID, got, want)
		}
	}
}

func TestS82DmuB10ShadowProphecyTakesTwoGravesTheRestAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82DmuDomain(g, me, 4)
	ids := seedLibrary(me, "A", "B", "C", "D", "E")
	life := me.Life
	castCatalogSpell(t, g, "Shadow Prophecy", "Instant", s82DmuShadowProphecyOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[0], ids[2])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("the two chosen cards are not in hand")
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[3]) {
		t.Error("the rest of the four looked at are not in the graveyard")
	}
	if me.Graveyard.Contains(ids[4]) || me.Hand.Contains(ids[4]) {
		t.Error("a card beyond the fourth was touched")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestS82DmuB10ShadowProphecyWithNoBasicTypesStillCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Shadow Prophecy", "Instant", s82DmuShadowProphecyOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestS82DmuB10VodaKeepsOneOnTopAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82DmuDomain(g, me, 3)
	ids := seedLibrary(me, "A", "B", "C", "D")
	castRingCreature(t, g, "Voda Sea Scavenger", "Creature — Merfolk Rogue", s82DmuVodaOracle, 3, 2)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[2])
	passPriorityAroundTable(t, g)
	top := me.Library.Cards[me.Library.Size()-1]
	if top.InstanceID != ids[2] {
		t.Errorf("the top card is %s, want the chosen C", top.Name)
	}
	if me.Library.Cards[me.Library.Size()-2].InstanceID != ids[3] {
		t.Error("the card beyond the look is not second from the top")
	}
	bottom := map[uuid.UUID]bool{}
	for i := 0; i < 2; i++ {
		bottom[me.Library.Cards[i].InstanceID] = true
	}
	if !bottom[ids[0]] || !bottom[ids[1]] {
		t.Error("A and B are not the two bottom cards")
	}
}

func TestS82DmuB10SlimefootsSurveyFetchesLandsWithABasicTypeThenLooks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	ids := seedSearchLibrary(me,
		searchTestLand("A", "Sorcery"), searchTestLand("B", "Sorcery"), searchTestLand("C", "Sorcery"), searchTestLand("D", "Sorcery"),
		searchTestLand("Dual", "Land — Forest Swamp"), searchTestLand("Wastes", "Basic Land"), searchTestLand("Forest", "Basic Land — Forest"))
	dual, wastes, forest := ids[4], ids[5], ids[6]
	castCatalogSpell(t, g, "Slimefoot's Survey", "Sorcery", s82DmuSlimefootsOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, dual, forest)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{dual, forest} {
		c := findCardAnywhere(t, g, id)
		if !g.Battlefield.Contains(id) || !c.Tapped {
			t.Errorf("%s did not enter tapped", c.Name)
		}
	}
	if g.Battlefield.Contains(wastes) {
		t.Error("Wastes has no basic land type and was fetched")
	}
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no look prompt after the search")
	}
	// Island, Forest and Swamp are three types, so the look is at three
	// (the library was shuffled by the search, so read the looked-at cards
	// off the prompt).
	looked := append([]uuid.UUID(nil), chooseCardsChoiceFor(g, me.ID).OptionCarry...)
	if len(looked) != 3 {
		t.Fatalf("looked at %d cards, want 3", len(looked))
	}
	size := me.Library.Size()
	answerChooseCards(t, g, me.ID, looked[1])
	passPriorityAroundTable(t, g)
	if me.Library.Size() != size {
		t.Errorf("library %d cards, want %d", me.Library.Size(), size)
	}
	if me.Library.Cards[size-1].InstanceID != looked[1] {
		t.Error("the kept card is not on top")
	}
	for i := 0; i < 2; i++ {
		if id := me.Library.Cards[i].InstanceID; id != looked[0] && id != looked[2] {
			t.Errorf("bottom card %d is not one of the cards that were not kept", i)
		}
	}
}

func TestS82DmuB10BortukReturnsOrTutorsByManaValue(t *testing.T) {
	for _, tc := range []struct {
		name   string
		types  int
		cost   string
		onBF   bool
		inHand bool
	}{
		{"mana value within domain returns to the battlefield", 3, "{1}{G}{G}", true, false},
		{"mana value above domain goes to hand", 2, "{1}{G}{G}", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			s82DmuDomain(g, me, tc.types)
			dead := pushGraveyardCreature(g, me.ID, "Bear", tc.cost)
			castHoldCreature(t, g, "Bortuk Bonerattle", "Legendary Creature — Troll Shaman", s82DmuBortukOracle, 4, 4)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, dead)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(dead) != tc.onBF || me.Hand.Contains(dead) != tc.inHand {
				t.Errorf("battlefield %v hand %v, want %v and %v", g.Battlefield.Contains(dead), me.Hand.Contains(dead), tc.onBF, tc.inHand)
			}
		})
	}
	t.Run("not cast does nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		s82DmuDomain(g, me, 3)
		dead := pushGraveyardCreature(g, me.ID, "Bear", "{1}")
		b12Creature(g, me.ID, "x", "Creature", 1, 1)
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bortuk Bonerattle", TypeLine: "Legendary Creature — Troll Shaman",
			OracleID: s82DmuBortukOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(dead) || me.Hand.Contains(dead) {
			t.Error("a Bortuk that was not cast returned a card")
		}
	})
}
