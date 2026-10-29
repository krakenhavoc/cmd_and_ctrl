package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 5. Every test drives
// the registered catalog row through the engine.

// --- Elven Passage ----------------------------------------------------

// hobB5Fetch activates Elven Passage (ability 0) and answers the search with
// a basic Island, returning the Island's id.
func hobB5Fetch(t *testing.T, g *game.Game, me *game.Player) (passage, island uuid.UUID) {
	t.Helper()
	passage = pushCatalogPermanent(g, me.ID, "Elven Passage", "Land", hobOracle(t, "Elven Passage"), false)
	stapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	island = stapleLibraryCard(me, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, island)
	return passage, island
}

func hobB5Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func TestHobElvenPassagePaysLifeSacrificesAndFetchesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	me.Hand.Cards = nil
	life := me.Life
	passage, island := hobB5Fetch(t, g, me)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the ability costs 1 life", me.Life, life-1)
	}
	if g.Battlefield.Contains(passage) {
		t.Error("Elven Passage was not sacrificed")
	}
	if !g.Battlefield.Contains(island) {
		t.Fatal("the basic land did not reach the battlefield")
	}
	if !hobB5Tapped(t, g, island) {
		t.Error("with no Elf to behold the land must stay tapped")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to reveal with no Elf card in hand")
	}
}

func TestHobElvenPassageBeholdingAnElfYouControlUntapsTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	me.Hand.Cards = nil
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	_, island := hobB5Fetch(t, g, me)
	if hobB5Tapped(t, g, island) {
		t.Error("an Elf you control beholds for free: the land should be untapped")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("choosing a controlled Elf needs no question")
	}
}

func TestHobElvenPassageAnOpponentsElfDoesNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	me.Hand.Cards = nil
	pushTribalCreature(g, opp.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	_, island := hobB5Fetch(t, g, me)
	if !hobB5Tapped(t, g, island) {
		t.Error("an Elf you do not control cannot be beheld")
	}
}

func TestHobElvenPassageRevealingAnElfCardFromHandUntapsTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	me.Hand.Cards = nil
	handCardFull(me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", nil)
	elf := handCardFull(me, "Elvish Visionary", "Creature — Elf Shaman", "{1}{G}", "", nil)
	_, island := hobB5Fetch(t, g, me)
	ch := chooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no reveal prompt with an Elf card in hand")
	}
	if len(ch.ChooseCards) != 1 || ch.ChooseCards[0] != elf {
		t.Errorf("candidates %v, want only the Elf card", ch.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, elf)
	if hobB5Tapped(t, g, island) {
		t.Error("revealing an Elf card should untap the land")
	}
	if !me.Hand.Contains(elf) {
		t.Error("the revealed Elf left the hand")
	}
}

func TestHobElvenPassageDecliningTheRevealLeavesTheLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	me.Hand.Cards = nil
	handCardFull(me, "Elvish Visionary", "Creature — Elf Shaman", "{1}{G}", "", nil)
	_, island := hobB5Fetch(t, g, me)
	answerChooseCards(t, g, me.ID)
	if !hobB5Tapped(t, g, island) {
		t.Error("declining to behold should leave the land tapped")
	}
}

// --- Long-Lost Lances -------------------------------------------------

func TestHobLongLostLancesGivesAllEquippedCreaturesFirstStrikeAndVigilanceOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	holder, other, bare := seedBear(g, me.ID), seedBear(g, me.ID), seedBear(g, me.ID)
	lances := seedEquipment(g, me.ID, "Long-Lost Lances", hobOracle(t, "Long-Lost Lances"))
	boots := seedEquipment(g, me.ID, "Swiftfoot Boots", hobOracle(t, "Swiftfoot Boots"))
	equipTo(t, g, me.ID, lances, holder)
	equipTo(t, g, me.ID, boots, other)

	if got := effectivePower(t, g, holder); got != 4 {
		t.Errorf("equipped creature power %d, want 2+2", got)
	}
	if got := effectivePower(t, g, other); got != 2 {
		t.Errorf("a creature wearing other Equipment gets no +2/+0, power %d", got)
	}
	for _, id := range []uuid.UUID{holder, other} {
		for _, kw := range []string{"first strike", "vigilance"} {
			if !effectiveAbilitiesContain(t, g, id, kw) {
				t.Errorf("an equipped creature lacks %s on its controller's turn", kw)
			}
		}
	}
	for _, kw := range []string{"first strike", "vigilance"} {
		if effectiveAbilitiesContain(t, g, bare, kw) {
			t.Errorf("an unequipped creature has %s", kw)
		}
	}

	advanceToNextSeatsTurn(t, g)
	for _, id := range []uuid.UUID{holder, other} {
		for _, kw := range []string{"first strike", "vigilance"} {
			if effectiveAbilitiesContain(t, g, id, kw) {
				t.Errorf("%s outside its controller's turn", kw)
			}
		}
	}
	if got := effectivePower(t, g, holder); got != 4 {
		t.Errorf("the +2/+0 is not turn-limited, power %d", got)
	}
}

func TestHobLongLostLancesGrantsNothingToAnOpponentsEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedEquipment(g, me.ID, "Long-Lost Lances", hobOracle(t, "Long-Lost Lances"))
	theirs := seedBear(g, opp.ID)
	boots := seedEquipment(g, opp.ID, "Swiftfoot Boots", hobOracle(t, "Swiftfoot Boots"))
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(boots, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	if effectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Error("the Lances gave first strike to an opponent's creature")
	}
}

// --- The Black Arrow --------------------------------------------------

// hobB5Arrow casts The Black Arrow from the active seat's hand and leaves the
// enters trigger's target prompt open.
func hobB5Arrow(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfdHandCard(me, "The Black Arrow", "Legendary Artifact — Equipment", "", hobOracle(t, "The Black Arrow"), 0, 0)
	advanceTo(t, g, game.StepBeginCombat) // only a flash spell is castable here
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestHobBlackArrowDestroysADragonItDamages(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dragon := hobB3Perm(g, opp.ID, game.Card{Name: "Dragon Whelp", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	hobB5Arrow(t, g)
	answerPickTarget(t, g, dragon)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) || !opp.Graveyard.Contains(dragon) {
		t.Error("a Dragon dealt damage by The Black Arrow should be destroyed")
	}
}

func TestHobBlackArrowOnlyPingsOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := hobB3Perm(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	hobB5Arrow(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, bear)
	if !ok || c.DamageMarked != 1 {
		t.Errorf("the Bear should survive with 1 damage: on battlefield %v, damage %d", ok, c.DamageMarked)
	}
}

func TestHobBlackArrowCanTargetAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	hobB5Arrow(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestHobBlackArrowEquipGivesPlusOneReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	arrow := seedEquipment(g, me.ID, "The Black Arrow", hobOracle(t, "The Black Arrow"))
	equipTo(t, g, me.ID, arrow, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped Bear is %d/%d, want 3/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("equipped creature lacks reach")
	}
}

// --- Troop of Ponies --------------------------------------------------

func TestHobTroopOfPoniesFetchesOneTappedAndOneToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	ponies := pushCatalogPermanent(g, me.ID, "Troop of Ponies", "Creature — Horse", hobOracle(t, "Troop of Ponies"), false)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	island := stapleLibraryCard(me, "Island", "Basic Land — Island")
	stapleLibraryCard(me, "Rancor", "Enchantment — Aura")
	if err := hobB3Activate(t, g, me.ID, ponies, "{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(ponies) {
		t.Error("the ponies should be sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	answerSearchByID(t, g, me.ID, island)
	if !g.Battlefield.Contains(forest) || !hobB5Tapped(t, g, forest) {
		t.Error("the first land should enter the battlefield tapped")
	}
	if !me.Hand.Contains(island) {
		t.Error("the second land should go to the hand")
	}
}

func TestHobTroopOfPoniesMayFindFewerThanTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	ponies := pushCatalogPermanent(g, me.ID, "Troop of Ponies", "Creature — Horse", hobOracle(t, "Troop of Ponies"), false)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	if err := hobB3Activate(t, g, me.ID, ponies, "{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchFailToFind(t, g, me.ID)
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the one land found should be on the battlefield")
	}
}

func TestHobTroopOfPoniesNeedsTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	ponies := pushCatalogPermanent(g, me.ID, "Troop of Ponies", "Creature — Horse", hobOracle(t, "Troop of Ponies"), false)
	if err := hobB3Activate(t, g, me.ID, ponies, "{C}"); err == nil {
		t.Error("activated for one mana")
	}
}

// --- Little Bear ------------------------------------------------------

func TestHobLittleBearUntapsABearAndGrowsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := hobB3Perm(g, me.ID, game.Card{Name: "Grizzly", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	if _, err := hobFlashCast(t, g, "Little Bear", "Creature — Bear", hobOracle(t, "Little Bear"), 3, 2); err != nil {
		t.Fatalf("flash cast at the beginning of combat: %v", err)
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if hobB5Tapped(t, g, bear) {
		t.Error("the Bear should be untapped")
	}
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1 on a Bear", n)
	}
}

func TestHobLittleBearUntapsOtherCreaturesWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wolf := hobB3Perm(g, me.ID, game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2, Tapped: true})
	hobCast(t, g, "Little Bear", "Creature — Bear", 3, 2)
	answerPickTarget(t, g, wolf)
	passPriorityAroundTable(t, g)
	if hobB5Tapped(t, g, wolf) {
		t.Error("the Wolf should be untapped")
	}
	if n := counterCount(g, wolf, "+1/+1"); n != 0 {
		t.Errorf("+1/+1 counters %d, want none on a non-Bear", n)
	}
}

func TestHobLittleBearCannotTargetItselfOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := hobB3Perm(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	hobCast(t, g, "Little Bear", "Creature — Bear", 3, 2)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the trigger asked for a target with no other creature you control")
	}
	if !hobB5Tapped(t, g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
}

// --- Old Fat Spider ---------------------------------------------------

func TestHobOldFatSpiderCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := hobPush(t, g, me.ID, "Old Fat Spider", "Creature — Spider", 6, 7)
	small := hobB3Perm(g, opp.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := hobB3Block(t, g, small, spider); err == nil {
		t.Error("a power 2 creature blocked Old Fat Spider")
	}
}

func TestHobOldFatSpiderCanBeBlockedByPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := hobPush(t, g, me.ID, "Old Fat Spider", "Creature — Spider", 6, 7)
	big := hobB3Perm(g, opp.ID, game.Card{Name: "Big", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	if err := hobB3Block(t, g, big, spider); err != nil {
		t.Errorf("a power 3 creature could not block: %v", err)
	}
}

func TestHobOldFatSpiderDrawsWhenAnOpponentTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := hobPush(t, g, me.ID, "Old Fat Spider", "Creature — Spider", 6, 7)
	g.Seats[0].Hand.Cards = nil
	castAtWardedCreature(t, g, opp, spider)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1: an opponent's spell targeted it", got)
	}
}

func TestHobOldFatSpiderDoesNotDrawWhenYouTargetIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := hobPush(t, g, me.ID, "Old Fat Spider", "Creature — Spider", 6, 7)
	me.Hand.Cards = nil
	castAtWardedCreature(t, g, me, spider)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 0 {
		t.Errorf("hand %d, want 0: its own controller targeted it", got)
	}
}

// --- Warg Tactics -----------------------------------------------------

func hobB5Tactics(t *testing.T, g *game.Game, mode int, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	_, err := hobB3Cast(t, g, me, "Warg Tactics", "Instant", "{1}{G}", "{C}{G}", game.CastSpellParams{
		Modes:   []int{mode},
		Targets: b16TargetCard(target),
	})
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

func TestHobWargTacticsDestroysAFlyingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bird := hobB3Perm(g, opp.ID, game.Card{Name: "Roc", TypeLine: "Creature — Bird", Power: 3, Toughness: 3, Keywords: []string{"flying"}})
	if err := hobB5Tactics(t, g, 0, bird); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(bird) || !opp.Graveyard.Contains(bird) {
		t.Error("the flier should be destroyed")
	}
}

func TestHobWargTacticsCannotDestroyAGroundCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := hobB3Perm(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := hobB5Tactics(t, g, 0, bear); err == nil {
		t.Error("the destroy mode accepted a creature without flying")
	}
}

func TestHobWargTacticsCountersAndGrantsTrampleAndHexproofUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	if err := hobB5Tactics(t, g, 1, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	for _, kw := range []string{"trample", "hexproof"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("the creature lacks %s", kw)
		}
	}
	advanceToNextSeatsTurn(t, g)
	for _, kw := range []string{"trample", "hexproof"} {
		if effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("%s lasted past the turn", kw)
		}
	}
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("the counter should stay, have %d", n)
	}
}

func TestHobWargTacticsSecondModeNeedsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	theirs := seedBear(g, g.Seats[1].ID)
	if err := hobB5Tactics(t, g, 1, theirs); err == nil {
		t.Error("the counter mode accepted an opponent's creature")
	}
}

// --- Galion, Elvenking's Butler --------------------------------------

func hobB5AddCounters(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, n); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
}

func TestHobGalionCopiesItsCurrentPowerAndToughnessOntoAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galion := hobPush(t, g, me.ID, "Galion, Elvenking's Butler", "Legendary Creature — Elf Advisor", 4, 4)
	hobB5AddCounters(t, g, galion, 1) // 5/5
	ally := seedBear(g, me.ID)
	hobB5AddCounters(t, g, ally, 1) // 3/3 before
	attackWithTrigger(t, g, galion, opp.ID, ally)
	if c := gaLayered(t, g, ally); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("ally is %d/%d, want base 5/5 (Galion's, counter included) plus its own counter", c.CurrentPower(), c.CurrentToughness())
	}
	advanceToNextSeatsTurn(t, g)
	if c := gaLayered(t, g, ally); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("ally is %d/%d after the turn, want 3/3 again", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobGalionCanChooseNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galion := hobPush(t, g, me.ID, "Galion, Elvenking's Butler", "Legendary Creature — Elf Advisor", 4, 4)
	ally := seedBear(g, me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(galion, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, ally); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("ally is %d/%d, want it left at 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobGalionCannotChooseAnOpponentsCreatureOrItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galion := hobPush(t, g, me.ID, "Galion, Elvenking's Butler", "Legendary Creature — Elf Advisor", 4, 4)
	theirs := seedBear(g, opp.ID)
	ally := seedBear(g, me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(galion, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	for _, bad := range []uuid.UUID{theirs, galion} {
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("Galion's trigger accepted %s", bad)
		}
	}
	answerPickTarget(t, g, ally)
}

// --- Mirkwood Nurturer ------------------------------------------------

func TestHobMirkwoodNurturerBouncesAPermanentAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := pushVanillaCreature(g, me.ID, "Ally", 1, 1)
	nurturer := hobCast(t, g, "Mirkwood Nurturer", "Creature — Elf Ranger", 3, 2)
	answerPickTarget(t, g, ally)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ally) || !me.Hand.Contains(ally) {
		t.Error("the chosen permanent should return to its owner's hand")
	}
	if n := counterCount(g, nurturer, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1 after the bounce", n)
	}
}

func TestHobMirkwoodNurturerCanBounceALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")
	nurturer := hobCast(t, g, "Mirkwood Nurturer", "Creature — Elf Ranger", 3, 2)
	answerPickTarget(t, g, land)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("a permanent that is not a creature should be returnable")
	}
	if n := counterCount(g, nurturer, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

func TestHobMirkwoodNurturerDecliningReturnsNothingAndAddsNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := pushVanillaCreature(g, me.ID, "Ally", 1, 1)
	nurturer := hobCast(t, g, "Mirkwood Nurturer", "Creature — Elf Ranger", 3, 2)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ally) {
		t.Error("declining should leave the permanent alone")
	}
	if n := counterCount(g, nurturer, "+1/+1"); n != 0 {
		t.Errorf("+1/+1 counters %d, want none when nothing returned", n)
	}
}

func TestHobMirkwoodNurturerCannotReturnAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	nurturer := hobCast(t, g, "Mirkwood Nurturer", "Creature — Elf Ranger", 3, 2)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("asked for a target with no other permanent you control")
	}
	if !g.Battlefield.Contains(theirs) || counterCount(g, nurturer, "+1/+1") != 0 {
		t.Error("an opponent's permanent was returned")
	}
}

// --- Orcrist, Goblin-cleaver -----------------------------------------

func TestHobOrcristPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	orcrist := seedEquipment(g, me.ID, "Orcrist, Goblin-cleaver", hobOracle(t, "Orcrist, Goblin-cleaver"))
	equipTo(t, g, me.ID, orcrist, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("equipped creature lacks trample")
	}
}

func hobB5TreasuresOf(g *game.Game, who uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == who && c.Name == "Treasure" {
			n++
		}
	}
	return n
}

func TestHobOrcristMakesATreasureForEachCreatureOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	carrier := pushTribalCreature(g, me.ID, "Goblin Raider", "Creature — Goblin Warrior", 2, 2)
	pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	pushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, opp.ID, "Their Goblin", "Creature — Goblin Scout", 1, 1)
	// "creature you control of that type": a Goblin that is not a creature
	// does not count.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Banner", TypeLine: "Kindred Artifact — Goblin",
		Owner: me.ID, Controller: me.ID,
	})
	orcrist := seedEquipment(g, me.ID, "Orcrist, Goblin-cleaver", hobOracle(t, "Orcrist, Goblin-cleaver"))
	equipTo(t, g, me.ID, orcrist, carrier)

	attackWith(t, g, opp.ID, carrier)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-4 {
		t.Fatalf("opponent life %d, want 36: the carrier is 4/4 with Orcrist", opp.Life)
	}
	if got := hobB5TreasuresOf(g, me.ID); got != 0 {
		t.Fatalf("%d Treasures before a type was chosen", got)
	}
	answerCreatureType(t, g, me.ID, "Goblin")
	passPriorityAroundTable(t, g)
	// Raider, Guide and the changeling.
	if got := hobB5TreasuresOf(g, me.ID); got != 3 {
		t.Errorf("%d Treasures, want 3 (two Goblins and a changeling)", got)
	}
	if got := hobB5TreasuresOf(g, opp.ID); got != 0 {
		t.Errorf("the opponent got %d Treasures", got)
	}
}

func TestHobOrcristTypeNobodyHasMakesNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	carrier := pushTribalCreature(g, me.ID, "Goblin Raider", "Creature — Goblin Warrior", 2, 2)
	orcrist := seedEquipment(g, me.ID, "Orcrist, Goblin-cleaver", hobOracle(t, "Orcrist, Goblin-cleaver"))
	equipTo(t, g, me.ID, orcrist, carrier)
	attackWith(t, g, opp.ID, carrier)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Sliver")
	passPriorityAroundTable(t, g)
	if got := hobB5TreasuresOf(g, me.ID); got != 0 {
		t.Errorf("%d Treasures for a type with no creatures", got)
	}
}

// --- Glamdring --------------------------------------------------------

func TestHobGlamdringPumpsPerInstantAndSorceryInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	glamdring := seedEquipment(g, me.ID, "Glamdring", hobOracle(t, "Glamdring"))
	equipTo(t, g, me.ID, glamdring, bear)
	if !effectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("equipped creature lacks first strike")
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("power %d with an empty graveyard, want 2", got)
	}
	pushGraveyardTyped(me, "Shock", "Instant")
	pushGraveyardTyped(me, "Divination", "Sorcery")
	pushGraveyardTyped(me, "Grizzly Bears", "Creature — Bear")
	pushGraveyardTyped(opp, "Their Shock", "Instant")
	seedBear(g, me.ID) // a battlefield entry invalidates the cached layers
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4: one instant and one sorcery in your graveyard, none of the others", got)
	}
	if p, tough := effectivePTOf(t, g, bear); tough != 2 || p != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

// hobB5GlamdringTable is a combat in which a Glamdring carrier connects for
// 2: the table sits in the combat damage step with the free-cast offer open.
func hobB5GlamdringTable(t *testing.T) (g *game.Game, me, opp *game.Player, bolt, bigger, creature uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	glamdring := seedEquipment(g, me.ID, "Glamdring", hobOracle(t, "Glamdring"))
	equipTo(t, g, me.ID, glamdring, bear)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bolt = handCardFull(me, "Lightning Bolt", "Instant", "{R}", boltOracleCombat, []string{"R"})
	bigger = handCardFull(me, "Big Sorcery", "Sorcery", "{2}{B}", "", []string{"B"})
	creature = handCardFull(me, "Cheap Bear", "Creature — Bear", "{G}", "", []string{"G"})
	// Glamdring grants first strike, so the damage is dealt in the first
	// combat damage step.
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	return g, me, opp, bolt, bigger, creature
}

func TestHobGlamdringOffersInstantsAndSorceriesWithinTheDamageDealt(t *testing.T) {
	g, me, opp, bolt, bigger, creature := hobB5GlamdringTable(t)
	if opp.Life != 38 {
		t.Fatalf("opponent life %d, want 38", opp.Life)
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	if !hasID(offer.DiscardOptions, bolt) {
		t.Errorf("the offer %v lacks the mana value 1 instant", offer.DiscardOptions)
	}
	if hasID(offer.DiscardOptions, bigger) {
		t.Error("a mana value 3 sorcery was offered for 2 damage")
	}
	if hasID(offer.DiscardOptions, creature) {
		t.Error("a creature spell was offered")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != 35 {
		t.Errorf("opponent life %d, want 35 after the free Lightning Bolt", opp.Life)
	}
	if me.Hand.Contains(bolt) {
		t.Error("the bolt is still in hand")
	}
}

func TestHobGlamdringFreeCastMayBeDeclined(t *testing.T) {
	g, me, _, bolt, _, _ := hobB5GlamdringTable(t)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	declineResolutionCast(t, g, me.ID)
	if !me.Hand.Contains(bolt) {
		t.Error("declining should leave the spell in hand")
	}
}

// --- Smaug, Wicked Worm ----------------------------------------------

func TestHobSmaugMakesATappedTreasureForEachArtifactYourOpponentsControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")
	seedPermanentFor(g, opp.ID, "Walking Ballista", "Artifact Creature — Construct")
	seedPermanentFor(g, g.Seats[2].ID, "Mind Stone", "Artifact")
	seedPermanentFor(g, me.ID, "My Trinket", "Artifact")
	seedPermanentFor(g, opp.ID, "Forest", "Basic Land — Forest")
	hobCast(t, g, "Smaug, Wicked Worm", "Legendary Creature — Dragon", 5, 5)
	if got := hobB5TreasuresOf(g, me.ID); got != 3 {
		t.Fatalf("%d Treasures, want 3 (three artifacts your opponents control)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" && !c.Tapped {
			t.Error("a Treasure entered untapped")
		}
	}
}

func TestHobSmaugWithNoOpponentArtifactsMakesNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentFor(g, me.ID, "My Trinket", "Artifact")
	hobCast(t, g, "Smaug, Wicked Worm", "Legendary Creature — Dragon", 5, 5)
	if got := hobB5TreasuresOf(g, me.ID); got != 0 {
		t.Errorf("%d Treasures, want none", got)
	}
}

func TestHobSmaugDrawsAndLosesLifeForSpellsPaidWithTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		treasure bool
		draw     int
	}{
		{"Treasure mana", true, 1},
		{"lands only", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			hobPush(t, g, me.ID, "Smaug, Wicked Worm", "Legendary Creature — Dragon", 5, 5)
			advanceToMain(t, g)
			if tc.treasure {
				crackATreasureFor(t, g, me, "B")
				floatForTest(g, me, "B")
			} else {
				floatForTest(g, me, "BB")
			}
			hand, life := handSizeOf(g, me), me.Life
			castFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", "{1}{B}", "", game.CastSpellParams{Strict: true})
			passPriorityAroundTable(t, g)
			if got := handSizeOf(g, me) - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
			if got := life - me.Life; got != tc.draw {
				t.Errorf("lost %d life, want %d", got, tc.draw)
			}
		})
	}
}
