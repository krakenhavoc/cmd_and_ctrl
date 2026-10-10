package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loyalty_x_cards_test.go — #1944: the loyalty cost of −X, and the
// planeswalkers that waited on it.

const (
	chandraAwakenedInfernoOracle = "7b2a600b-d6c8-45ff-a7fc-06105d27111f"
	chandraNalaarOracle          = "f4da7a43-e9d5-47ea-b2f0-3cf55a156f4a"
	chandraChillOracle           = "c3dfa1e2-6785-49a0-a194-fb842a8eb63c"
)

// pushSizedCreatureForTest seeds a creature with a size, a mana cost,
// colours and a subtype line.
func pushSizedCreatureForTest(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, colors []string, p, t int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Colors:     colors,
		Power:      p,
		Toughness:  t,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func exiled(g *game.Game, id uuid.UUID) bool { return g.Exile.Contains(id) }

// TestChandraAwakenedInfernoMinusXRemovesXAndExiles: X is announced,
// X loyalty leaves, the target takes X, and the permanent that would
// die is exiled instead.
func TestChandraAwakenedInfernoMinusXRemovesXAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Awakened Inferno", chandraAwakenedInfernoOracle, 6)
	ogre := pushSizedCreatureForTest(g, them.ID, "Ogre", "Creature — Ogre", "{2}{R}", []string{"R"}, 3, 3)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
	}); err != nil {
		t.Fatalf("−X at X=3: %v", err)
	}
	if got := loyaltyOf(g, chandra); got != 3 {
		t.Errorf("loyalty after −3 = %d, want 3", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Fatal("a 3/3 survived 3 damage")
	}
	if !exiled(g, ogre) {
		t.Error("the creature dealt lethal damage was not exiled instead of dying")
	}
}

// TestLoyaltyMinusXCannotExceedLoyalty is CR 606.6 at an announced X:
// more than the loyalty there is refused and changes nothing; exactly
// the loyalty is legal, and the planeswalker goes to CR 704.5i.
func TestLoyaltyMinusXCannotExceedLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := pushWalkerForTest(g, me.ID, "Chandra Nalaar", chandraNalaarOracle, 6)
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Bear")
	advanceToMainOf(t, g, seat)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{XValue: 7, Targets: target})
	if !errors.Is(err, game.ErrInsufficientLoyalty) {
		t.Fatalf("X=7 with 6 loyalty: err = %v, want ErrInsufficientLoyalty", err)
	}
	if got := loyaltyOf(g, chandra); got != 6 {
		t.Fatalf("a refused activation moved loyalty to %d", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{XValue: 6, Targets: target}); err != nil {
		t.Fatalf("X=6 with 6 loyalty: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chandra) {
		t.Error("a planeswalker paid down to 0 loyalty stayed on the battlefield")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("6 damage did not kill a 2/2")
	}
}

// TestChandraAwakenedInfernoMinusThreeSparesElementals.
func TestChandraAwakenedInfernoMinusThreeSparesElementals(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Awakened Inferno", chandraAwakenedInfernoOracle, 6)
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Bear")
	elemental := pushSizedCreatureForTest(g, them.ID, "Spark", "Creature — Elemental", "{R}", []string{"R"}, 1, 1)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a non-Elemental 2/2 survived 3 damage")
	}
	if !g.Battlefield.Contains(elemental) {
		t.Error("the Elemental was dealt damage")
	}
}

// TestChandraAwakenedInfernoPlusTwoGivesEachOpponentAnEmblem: one
// emblem per opponent, none for Chandra's controller, and an emblem
// deals its owner 1 damage in that owner's upkeep.
func TestChandraAwakenedInfernoPlusTwoGivesEachOpponentAnEmblem(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Awakened Inferno", chandraAwakenedInfernoOracle, 6)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	if got := loyaltyOf(g, chandra); got != 8 {
		t.Errorf("loyalty after +2 = %d, want 8", got)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 1
		if p.ID == me.ID {
			want = 0
		}
		if got := p.Emblems.Size(); got != want {
			t.Errorf("%s has %d emblems, want %d", p.Name, got, want)
		}
	}

	next := (seat + 1) % len(g.Seats)
	opp := g.Seats[next]
	life := opp.Life
	advanceToMainOf(t, g, next)
	if got := life - opp.Life; got != 1 {
		t.Errorf("the emblem dealt its owner %d damage in their upkeep, want 1", got)
	}
}

// TestChandraNalaarMinusEightHitsThePlayerAndTheirCreatures.
func TestChandraNalaarMinusEightHitsThePlayerAndTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := pushWalkerForTest(g, me.ID, "Chandra Nalaar", chandraNalaarOracle, 9)
	theirs := pushSizedCreatureForTest(g, them.ID, "Giant", "Creature — Giant", "{4}{R}", []string{"R"}, 9, 9)
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	advanceToMainOf(t, g, seat)
	life := them.Life

	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - them.Life; got != 10 {
		t.Errorf("damage to the player = %d, want 10", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("their 9/9 survived 10 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("a creature of another player was dealt damage")
	}
}

// TestChandraChillOfComplianceMinusXTapsAndStuns.
func TestChandraChillOfComplianceMinusXTapsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Chill of Compliance", chandraChillOracle, 3)
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Bear")
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("−X at X=2: %v", err)
	}
	if got := loyaltyOf(g, chandra); got != 1 {
		t.Errorf("loyalty after −2 = %d, want 1", got)
	}
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(bear)
	if !c.Tapped || c.Counters[game.CounterStun] != 2 {
		t.Errorf("target tapped=%v with %d stun counters, want tapped with 2", c.Tapped, c.Counters[game.CounterStun])
	}
}

// TestUginMinusXExilesColoredPermanentsUpToX: coloured permanents with
// mana value X or less go; a bigger one and a colourless one stay.
func TestUginMinusXExilesColoredPermanentsUpToX(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	ugin := pushWalkerForTest(g, me.ID, "Ugin, the Spirit Dragon", uginSpiritDragonOracle, 7)
	small := pushSizedCreatureForTest(g, them.ID, "Small", "Creature — Elf", "{1}{G}", []string{"G"}, 2, 2)
	big := pushSizedCreatureForTest(g, them.ID, "Big", "Creature — Wurm", "{4}{G}", []string{"G"}, 5, 5)
	golem := pushSizedCreatureForTest(g, them.ID, "Golem", "Artifact Creature — Golem", "{2}", nil, 2, 2)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, ugin, 1, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("−X at X=3: %v", err)
	}
	if got := loyaltyOf(g, ugin); got != 4 {
		t.Errorf("loyalty after −3 = %d, want 4", got)
	}
	passPriorityAroundTable(t, g)
	if !exiled(g, small) {
		t.Error("a green mana value 2 creature was not exiled at X=3")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a mana value 5 creature was exiled at X=3")
	}
	if !g.Battlefield.Contains(golem) {
		t.Error("a colourless creature was exiled")
	}
	if !g.Battlefield.Contains(ugin) {
		t.Error("Ugin, who is colourless, was exiled by his own sweep")
	}
}

// TestLoyaltyMinusXRegistersAsDemandingX: the cost announces X, and the
// fixed loyalty components still do not.
func TestLoyaltyMinusXRegistersAsDemandingX(t *testing.T) {
	cost := LoyaltyMinusX()
	if !cost.DemandsX() || cost.Loyalty == nil || *cost.Loyalty != 0 {
		t.Fatalf("LoyaltyMinusX = %+v, want a zero loyalty cost that demands X", cost)
	}
	if got := cost.LoyaltyDelta(4); got != -4 {
		t.Errorf("LoyaltyDelta(4) = %d, want -4", got)
	}
	if LoyaltyCost(-3).DemandsX() {
		t.Error("a fixed −3 demands X")
	}
	if got := LoyaltyCost(2).LoyaltyDelta(5); got != 2 {
		t.Errorf("a fixed +2 at X=5 = %d, want 2", got)
	}
}

const jeskaOracle = "b1fbfcf3-6921-4417-a58e-0f5e5d34a105"

// TestJeskaEntersWithALoyaltyCounterPerCommanderCast: the count is
// every command-zone cast of every commander, and with none she dies to
// CR 704.5i.
func TestJeskaEntersWithALoyaltyCounterPerCommanderCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2, uuid.New(): 1}
	})
	castCatalogSpell(t, g, "Jeska, Thrice Reborn", "Legendary Planeswalker — Jeska", jeskaOracle, nil)
	passPriorityAroundTable(t, g)
	jeska := findBattlefieldByName(g, "Jeska, Thrice Reborn")
	if jeska == uuid.Nil {
		t.Fatal("Jeska did not stay on the battlefield after three commander casts")
	}
	if got := loyaltyOf(g, jeska); got != 3 {
		t.Errorf("loyalty = %d, want 3", got)
	}

	g2 := newCatalogGame(t)
	castCatalogSpell(t, g2, "Jeska, Thrice Reborn", "Legendary Planeswalker — Jeska", jeskaOracle, nil)
	passPriorityAroundTable(t, g2)
	if findBattlefieldByName(g2, "Jeska, Thrice Reborn") != uuid.Nil {
		t.Error("Jeska with no commander casts survived with 0 loyalty")
	}
}

// TestJeskaMinusXHitsUpToThreeTargets: X to each, at once.
func TestJeskaMinusXHitsUpToThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	jeska := pushWalkerForTest(g, me.ID, "Jeska, Thrice Reborn", jeskaOracle, 4)
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Bear")
	ogre := pushSizedCreatureForTest(g, them.ID, "Ogre", "Creature — Ogre", "{2}{R}", []string{"R"}, 3, 3)
	advanceToMainOf(t, g, seat)
	life := them.Life

	if err := g.ActivateCatalogAbility(me.ID, jeska, 1, game.ActivateAbilityParams{
		XValue: 2,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: bear},
			{Kind: game.TargetCard, ID: ogre},
			{Kind: game.TargetPlayer, ID: them.ID},
		},
	}); err != nil {
		t.Fatalf("−X at X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 died to 2 damage")
	}
	if got := life - them.Life; got != 2 {
		t.Errorf("damage to the player = %d, want 2", got)
	}
	if got := loyaltyOf(g, jeska); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

// TestJeskaZeroRegistersATripler pins the 0 on her controller's chosen
// creature, without moving her loyalty.
func TestJeskaZeroRegistersATripler(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	jeska := pushWalkerForTest(g, me.ID, "Jeska, Thrice Reborn", jeskaOracle, 2)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, jeska, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, jeska); got != 2 {
		t.Errorf("loyalty after 0 = %d, want 2", got)
	}
	if lines := g.DamageMultiplierLines(); len(lines) != 1 {
		t.Fatalf("damage multipliers = %v, want one", lines)
	}
}
