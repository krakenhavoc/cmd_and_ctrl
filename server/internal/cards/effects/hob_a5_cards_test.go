package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a5_cards_test.go — The Hobbit / The Hobbit Eternal, half A, fifth
// batch (#4454): the "second card each turn" creatures, four Dwarven
// equipment cards' worth of enters triggers, and Along the Crooked Way.
// Each test plays the real catalog entry and asserts what its text says.

const (
	hobA5KnightsOracle     = "3f8b189e-775c-41ad-839a-153a3200c708"
	hobA5ApothecaryOracle  = "80bcb154-b4be-4bbd-bbc1-b116a59045b7"
	hobA5CouncillorsOracle = "b9079564-5c86-4765-988a-f2334e016fa6"
	hobA5BardOracle        = "ec076f5b-b0e3-4b6f-9293-a8fc42f20bd8"
	hobA5MattockOracle     = "f75bb13b-41fc-4614-b35e-f456069ce9c6"
	hobA5CrudeBladeOracle  = "eec81460-899d-4b4c-b33c-ce2ca704df17"
	hobA5SpearOracle       = "bb67d0a4-864e-4f46-8538-fdd920bf0197"
	hobA5CrookedWayOracle  = "f4437cb3-85f7-49b5-bb41-6a948549d987"
	hobA5CouncillorsName   = "Master's Councillors"
	hobA5CouncillorsTypeLn = "Creature — Human Advisor"
	hobA5EquipmentTypeLine = "Artifact — Equipment"
)

func TestHobA5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA5KnightsOracle:     "Knights of Dol Amroth",
		hobA5ApothecaryOracle:  "Lakeshore Apothecary",
		hobA5CouncillorsOracle: hobA5CouncillorsName,
		hobA5BardOracle:        "Bard the Bowman",
		hobA5MattockOracle:     "Dwarven Mattock",
		hobA5CrudeBladeOracle:  "Crude Bent Blade",
		hobA5SpearOracle:       "Ragged Short Spear",
		hobA5CrookedWayOracle:  "Along the Crooked Way",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- the "second card each turn" creatures -------------------------------

// hobA5Draw has p draw one card and lets the stack settle, answering a
// target prompt with pick when one is open.
func hobA5Draw(t *testing.T, g *game.Game, p *game.Player, pick *game.TargetRef) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	assistCardSettle(t, g, p.ID, pick, true)
}

func hobA5Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return ringBFCard(t, g, id).Counters[game.CounterPlusOne]
}

// hobA5SecondDrawGrows runs the shared shape of Knights of Dol Amroth and
// Lakeshore Apothecary: a counter on the second draw, none on the first or
// the third, and none for a draw by another player.
func hobA5SecondDrawGrows(t *testing.T, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	id := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	if n := hobA5Counters(t, g, id); n != 0 {
		t.Fatalf("%d counters after another player drew twice, want 0", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 0 {
		t.Fatalf("%d counters after the first draw, want 0", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters after the second draw, want 1", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters after the third draw, want it to stay 1", n)
	}
	return id
}

func TestHobA5KnightsOfDolAmrothGrowsOnTheSecondDrawOnly(t *testing.T) {
	hobA5SecondDrawGrows(t, "Knights of Dol Amroth", "Creature — Human Knight", hobA5KnightsOracle)
}

func TestHobA5LakeshoreApothecaryGrowsOnTheSecondDrawAndHasVigilance(t *testing.T) {
	hobA5SecondDrawGrows(t, "Lakeshore Apothecary", "Creature — Human Cleric", hobA5ApothecaryOracle)
	g := newCatalogGame(t)
	vig := pushCatalogPermanent(g, g.Seats[0].ID, "Lakeshore Apothecary", "Creature — Human Cleric", hobA5ApothecaryOracle, false)
	if !hasKeywordOnBattlefield(t, g, vig, "vigilance") {
		t.Error("the Apothecary has no vigilance")
	}
}

func TestHobA5SecondDrawResetsEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Knights of Dol Amroth", "Creature — Human Knight", hobA5KnightsOracle, false)
	hobA5Draw(t, g, me, nil)
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters on turn one, want 1", n)
	}
	advanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 1 {
		advanceToNextSeatsTurn(t, g)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters after one more draw on the next turn, want 1", n)
	}
}

func hobA5Graveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: p.ID, Controller: p.ID,
		})
	}
}

func hobA5Councillors(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: hobA5CouncillorsName, TypeLine: hobA5CouncillorsTypeLn,
		OracleID: hobA5CouncillorsOracle, Power: 1, Toughness: 3, Owner: owner, Controller: owner,
	})
}

func TestHobA5MastersCouncillorsGetsTwoPowerForEachGraveyardWithSevenCards(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[1], g.Seats[2]
	hobA5Graveyard(a, 6)
	id := hobA5Councillors(g, me.ID)
	if p, tt := ptOf(t, g, id); p != 1 || tt != 3 {
		t.Fatalf("with no full graveyard it is %d/%d, want 1/3", p, tt)
	}
	if !hasKeywordOnBattlefield(t, g, id, "vigilance") {
		t.Error("no vigilance")
	}

	g2 := newCatalogGame(t)
	me2, a2, b2 := g2.Seats[1], g2.Seats[2], g2.Seats[3]
	hobA5Graveyard(a2, 7)
	hobA5Graveyard(b2, 8)
	hobA5Graveyard(me2, 3)
	id2 := hobA5Councillors(g2, me2.ID)
	if p, tt := ptOf(t, g2, id2); p != 5 || tt != 3 {
		t.Fatalf("with two full graveyards it is %d/%d, want 5/3", p, tt)
	}
}

func TestHobA5MastersCouncillorsMillsThreeOnTheSecondDrawAndGrowsFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[1], g.Seats[2]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	hobA5Graveyard(victim, 5)
	id := hobA5Councillors(g, me.ID)
	pick := &game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID}

	hobA5Draw(t, g, me, pick)
	if n := victim.Graveyard.Size(); n != 5 {
		t.Fatalf("the first draw milled: graveyard %d, want 5", n)
	}
	hobA5Draw(t, g, me, pick)
	if n := victim.Graveyard.Size(); n != 8 {
		t.Fatalf("the second draw: graveyard %d, want 8", n)
	}
	if p, _ := ptOf(t, g, id); p != 3 {
		t.Errorf("the milled graveyard now has eight cards: power %d, want 3", p)
	}
	hobA5Draw(t, g, me, pick)
	if n := victim.Graveyard.Size(); n != 8 {
		t.Errorf("the third draw milled again: graveyard %d, want 8", n)
	}
}

func TestHobA5BardPutsACounterOnTargetCreatureAndGivesLifelinkOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Bard the Bowman", "Legendary Creature — Human Archer", hobA5BardOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pick := &game.TargetRef{Kind: game.TargetCard, ID: theirs}

	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, theirs); n != 0 {
		t.Fatalf("the first draw put %d counters, want 0", n)
	}
	hobA5Draw(t, g, me, pick)
	if n := hobA5Counters(t, g, theirs); n != 1 {
		t.Fatalf("%d counters on the target, want 1 (any creature may be targeted)", n)
	}
	if !hasKeywordOnBattlefield(t, g, theirs, "lifelink") {
		t.Error("the target did not gain lifelink")
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, theirs); n != 1 {
		t.Errorf("the third draw changed the counters to %d", n)
	}
	advanceToNextSeatsTurn(t, g)
	if hasKeywordOnBattlefield(t, g, theirs, "lifelink") {
		t.Error("lifelink lasted past the turn")
	}
}

func TestHobA5BardHasReach(t *testing.T) {
	g := newCatalogGame(t)
	id := pushCatalogPermanent(g, g.Seats[0].ID, "Bard the Bowman", "Legendary Creature — Human Archer", hobA5BardOracle, false)
	if !hasKeywordOnBattlefield(t, g, id, "reach") {
		t.Error("no reach")
	}
}

// --- the Equipment -------------------------------------------------------

func TestHobA5DwarvenMattockAttachesToATargetDwarfAndWardsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := b12Creature(g, me.ID, "Test Dwarf", "Creature — Dwarf", 2, 2)
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	mattock := castCatalogSpell(t, g, "Dwarven Mattock", hobA5EquipmentTypeLine, hobA5MattockOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: dwarf}, true)

	if p, tt := ptOf(t, g, dwarf); p != 4 || tt != 4 {
		t.Fatalf("the Dwarf is %d/%d, want 4/4", p, tt)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the non-Dwarf bear is power %d, want 2", p)
	}
	if c, _ := battlefieldCard(g, mattock); !c.IsAttachedTo(dwarf) {
		t.Fatal("the Mattock is not attached to the Dwarf")
	}

	castAtWardedCreature(t, g, opp, dwarf)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} prompt for the opponent's spell")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dwarf) {
		t.Error("a declined ward must counter the spell; the Dwarf should live")
	}
}

func TestHobA5DwarvenMattockStaysUnattachedWithoutADwarfAndEquipsForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	mattock := castCatalogSpell(t, g, "Dwarven Mattock", hobA5EquipmentTypeLine, hobA5MattockOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	if c, _ := battlefieldCard(g, mattock); c.IsAttachedTo(bear) {
		t.Fatal("the Mattock attached to a creature that is not a Dwarf")
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Fatalf("the bear is power %d, want 2", p)
	}
	equipTo(t, g, me.ID, mattock, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("after equipping the bear is %d/%d, want 4/4", p, tt)
	}
}

func TestHobA5CrudeBentBladeMakesTheOpponentSacrificeOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := b12Creature(g, opp.ID, "Keeper", "Creature — Bear", 2, 2)
	lose := b12Creature(g, opp.ID, "Loser", "Creature — Bear", 1, 1)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	blade := castCatalogSpell(t, g, "Crude Bent Blade", hobA5EquipmentTypeLine, hobA5CrudeBladeOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	answerSacrifice(t, g, opp.ID, lose)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(lose) {
		t.Error("the creature the opponent chose is still on the battlefield")
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(mine) {
		t.Error("a creature other than the chosen one was sacrificed")
	}
	equipTo(t, g, me.ID, blade, mine)
	if p, tt := ptOf(t, g, mine); p != 4 || tt != 3 {
		t.Errorf("equipped creature is %d/%d, want 4/3", p, tt)
	}
}

func TestHobA5CrudeBentBladeLeavesItsControllersCreaturesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Crude Bent Blade", hobA5EquipmentTypeLine, hobA5CrudeBladeOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("only the opponent's creature should be sacrificed")
	}
}

func TestHobA5RaggedShortSpearDiscardsToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Ragged Short Spear", hobA5EquipmentTypeLine, hobA5SpearOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	pitch := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, pitch)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (one discarded, two drawn)", got, hand+1)
	}
}

func TestHobA5RaggedShortSpearDrawsNothingWhenYouDeclineTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Ragged Short Spear", hobA5EquipmentTypeLine, hobA5SpearOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if c := discardChoiceFor(g, me.ID); c != nil {
		answerDiscard(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declined: hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestHobA5RaggedShortSpearEquipsForThreeAndGivesTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	spear := seedEquipment(g, me.ID, "Ragged Short Spear", hobA5SpearOracle)
	equipTo(t, g, me.ID, spear, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("equipped creature is %d/%d, want 4/2", p, tt)
	}
}

// --- Along the Crooked Way -----------------------------------------------

// hobA5Armies counts the Army creatures `owner` controls and returns the
// largest one's power.
func hobA5Armies(t *testing.T, g *game.Game, owner uuid.UUID) (n, power int) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.IsCreature() && c.HasSubtypeWord("Army") {
				n++
				if p := c.PowerForComparison(); p > power {
					power = p
				}
			}
		}
	})
	return n, power
}

func hobA5CreatureCard(owner uuid.UUID, name string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
}

func hobA5CrookedWayTable(t *testing.T) (g *game.Game, me *game.Player, way uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	way = pushCatalogPermanent(g, me.ID, "Along the Crooked Way", "Enchantment", hobA5CrookedWayOracle, false)
	return g, me, way
}

func TestHobA5AlongTheCrookedWayReturnsACreatureCardAndAmassesOnceForIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	other := seedGraveyardCard(t, g, "Dead Land", "Land", "")
	castCatalogSpell(t, g, "Along the Crooked Way", "Enchantment", hobA5CrookedWayOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, true)

	if !me.Hand.Contains(target) {
		t.Fatal("the creature card did not return to hand")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the land card left the graveyard")
	}
	if n, p := hobA5Armies(t, g, me.ID); n != 1 || p != 1 {
		t.Errorf("%d Army creatures (power %d), want one 1/1 from the card leaving", n, p)
	}
}

func TestHobA5AlongTheCrookedWayAmassesForEachCreatureCardLeavingYourGraveyard(t *testing.T) {
	g, me, _ := hobA5CrookedWayTable(t)
	opp := g.Seats[1]
	mine := hobA5CreatureCard(me.ID, "My Dead Bear")
	land := game.Card{InstanceID: uuid.New(), Name: "Dead Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID}
	theirs := hobA5CreatureCard(opp.ID, "Their Dead Bear")
	me.Graveyard.PushTop(mine)
	me.Graveyard.PushTop(land)
	opp.Graveyard.PushTop(theirs)

	// A creature card leaving an opponent's graveyard, and a land card
	// leaving mine, do nothing.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs.InstanceID) })
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(land.InstanceID) })
	passPriorityAroundTable(t, g)
	if n, _ := hobA5Armies(t, g, me.ID); n != 0 {
		t.Fatalf("%d Armies from cards that are not creature cards of mine", n)
	}

	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine.InstanceID) })
	passPriorityAroundTable(t, g)
	if n, p := hobA5Armies(t, g, me.ID); n != 1 || p != 1 {
		t.Fatalf("%d Armies (power %d) after one creature card left, want one 1/1", n, p)
	}
	second := hobA5CreatureCard(me.ID, "Second Dead Bear")
	me.Graveyard.PushTop(second)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(second.InstanceID) })
	passPriorityAroundTable(t, g)
	if n, p := hobA5Armies(t, g, me.ID); n != 1 || p != 2 {
		t.Errorf("%d Armies (power %d) after the second, want the same Army at 2/2", n, p)
	}
}

func TestHobA5AlongTheCrookedWayGivesGoblinsAndOrcsMenace(t *testing.T) {
	g, me, way := hobA5CrookedWayTable(t)
	advanceToMain(t, g)
	goblin := b12Creature(g, me.ID, "Test Goblin", "Creature — Goblin", 1, 1)
	orc := b12Creature(g, me.ID, "Test Orc", "Creature — Orc", 1, 1)
	human := b12Creature(g, me.ID, "Test Human", "Creature — Human", 1, 1)
	theirGoblin := b12Creature(g, g.Seats[1].ID, "Their Goblin", "Creature — Goblin", 1, 1)

	if err := g.ActivateCatalogAbility(me.ID, way, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{goblin: true, orc: true, human: false, theirGoblin: false} {
		if got := hasKeywordOnBattlefield(t, g, id, "menace"); got != want {
			t.Errorf("menace on %s = %v, want %v", id, got, want)
		}
	}
}
