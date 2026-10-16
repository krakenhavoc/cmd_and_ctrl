package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remove_two_counters_cards_test.go — #3987: the cards whose cost is
// "Remove two counters from this creature", two counters of any kinds
// off the source (RemoveCountersOfAnyKindsFromThis). Gwen Stacy //
// Ghost-Spider first, then the three Treefolk-cycle creatures that
// enter with -1/-1 counters and spend them.

const (
	gnarlbarkElmOracle  = "a848ea46-347d-49c7-85c8-023cf836d0b3"
	lochMareOracle      = "631eb3e9-a3e0-4341-ad2d-e79142ce537d"
	reapingWillowOracle = "ce60402c-34cd-44fe-ae1d-4a5404298f5b"
)

// gwenStacyCard is the modal double-faced card as the deck importer
// builds it, front face up.
func gwenStacyCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   gwenStacyOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Gwen Stacy",
				TypeLine:  "Legendary Creature — Human Performer Hero",
				ManaCost:  "{1}{R}",
				Colors:    []string{"R"},
				Power:     2,
				Toughness: 1,
			},
			{
				Name:      "Ghost-Spider",
				TypeLine:  "Legendary Creature — Spider Human Hero",
				ManaCost:  "{2}{U}{R}{W}",
				Colors:    []string{"R", "U", "W"},
				Power:     4,
				Toughness: 4,
			},
		},
	}
	c.SetFace(0)
	return c
}

// setCountersOn writes a battlefield card's counters directly.
func setCountersOn(g *game.Game, id uuid.UUID, counters map[string]int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = counters
			}
		}
	})
}

// castCreatureCard casts a catalog creature with its printed body, so an
// "enters with -1/-1 counters" replacement has something to shrink, and
// lets it resolve.
func castCreatureCard(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- Gwen Stacy // Ghost-Spider ---------------------------------------

// Gwen enters, exiles the top card with a play window tied to her, and
// the window survives her transform: Ghost-Spider is the same object
// (CR 712.18).
func TestGwenStacyExilesOnEntryAndTheWindowSurvivesTheTransform(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	held := topCard(me, "Held Bear", "Creature — Bear", "{G}")
	gwen := gwenStacyCard(me.ID)
	me.Hand.PushTop(gwen)
	if err := g.CastSpell(me.ID, gwen.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Gwen Stacy: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(held) {
		t.Fatal("the top card was not exiled when Gwen entered")
	}
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Fatal("the exiled card is not playable while Gwen is in play")
	}

	if err := g.ActivateCatalogAbility(me.ID, gwen.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, gwen.InstanceID)
	if card == nil || card.ActiveFace != 1 || card.Name != "Ghost-Spider" {
		t.Fatalf("after the transform: %+v, want Ghost-Spider in place", card)
	}
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Error("the window closed when Gwen became Ghost-Spider; she is the same creature")
	}

	// Next turn too: "for as long as you control this creature".
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Error("the window closed at end of turn")
	}
}

// The transform is sorcery-speed: not with something on the stack.
func TestGwenStacyTransformsOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gwen := gwenStacyCard(me.ID)
	g.Battlefield.PushTop(gwen)
	stackLibrary(me, "Filler", "Sorcery", "")
	// Something on the stack: a spell of ours.
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Pending Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, gwen.InstanceID, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with a spell on the stack")
	}
}

// Ghost-Spider's trigger counts a spell cast from exile, and her ability
// takes two counters of DIFFERENT kinds when the activator says so.
func TestGhostSpiderGrowsFromExileAndSpendsTwoCountersOfAnyKinds(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spider := gwenStacyCard(me.ID)
	spider.SetFace(1)
	g.Battlefield.PushTop(spider)
	id := spider.InstanceID

	// A spell cast from exile through an impulse grant: +1/+1.
	bear := topCard(me, "Exiled Bear", "Creature — Bear", "{G}")
	var exileErr error
	g.WithWriteLock(func() {
		_, exileErr = g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{})
	})
	if exileErr != nil {
		t.Fatalf("exile: %v", exileErr)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Fatalf("Ghost-Spider has %d +1/+1 counters after a spell from exile, want 1", got)
	}

	// One counter is not two: the ability cannot be paid.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{CounterKind: game.CounterPlusOne}); err == nil {
		t.Fatal("paid two counters with one")
	}

	// A +1/+1 and a shield counter: one legal payment.
	setCountersOn(g, id, map[string]int{game.CounterPlusOne: 1, game.CounterShield: 1})
	top := topCard(me, "Next Card", "Sorcery", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		CounterCounts: []int{1, 1},
		CounterKinds:  []string{game.CounterPlusOne, game.CounterShield},
	}); err != nil {
		t.Fatalf("activate with a +1/+1 and a shield counter: %v", err)
	}
	if countersOn(g, id, game.CounterPlusOne) != 0 || countersOn(g, id, game.CounterShield) != 0 {
		t.Fatal("the counters did not come off as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
	nextTurnOfMine(t, g)
	if permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the card is still playable a turn later; the grant is for this turn")
	}
}

// --- Gnarlbark Elm ----------------------------------------------------

func TestGnarlbarkElmSpendsItsMinusCountersForMinusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elm := castCreatureCard(t, g, "Gnarlbark Elm", "Creature — Treefolk Warlock", gnarlbarkElmOracle, 3, 4)
	if got := countersOn(g, elm, game.CounterMinusOne); got != 2 {
		t.Fatalf("Gnarlbark Elm entered with %d -1/-1 counters, want 2", got)
	}
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)

	if err := g.ActivateCatalogAbility(me.ID, elm, 0, game.ActivateAbilityParams{
		CounterKind: game.CounterMinusOne,
		Targets:     []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, elm); p != 3 || tough != 4 {
		t.Errorf("Gnarlbark Elm is %d/%d with its counters spent, want 3/4", p, tough)
	}
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}

	// No counters left: the cost cannot be paid again.
	err := g.ActivateCatalogAbility(me.ID, elm, 0, game.ActivateAbilityParams{
		CounterKind: game.CounterMinusOne,
		Targets:     []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err == nil {
		t.Fatal("activated with no counters to remove")
	}
}

// --- Loch Mare ----------------------------------------------------------

func TestLochMareDrawsForOneCounterAndStunsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mare := castCreatureCard(t, g, "Loch Mare", "Creature — Horse Serpent", lochMareOracle, 4, 5)
	if got := countersOn(g, mare, game.CounterMinusOne); got != 3 {
		t.Fatalf("Loch Mare entered with %d -1/-1 counters, want 3", got)
	}
	stackLibrary(me, "Drawn", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mare, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}

	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, mare, 1, game.ActivateAbilityParams{
		CounterKind: game.CounterMinusOne,
		Targets:     []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("stun ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := battlefieldCardFor(g, victim); c == nil || !c.Tapped {
		t.Error("the target is not tapped")
	}
	if got := countersOn(g, victim, game.CounterStun); got != 1 {
		t.Errorf("the target has %d stun counters, want 1", got)
	}
	if got := countersOn(g, mare, game.CounterMinusOne); got != 0 {
		t.Errorf("Loch Mare has %d -1/-1 counters left, want 0", got)
	}
}

// --- Reaping Willow -----------------------------------------------------

func TestReapingWillowReturnsACreatureCardOfManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	willow := castCreatureCard(t, g, "Reaping Willow", "Creature — Treefolk Cleric", reapingWillowOracle, 3, 6)
	if got := countersOn(g, willow, game.CounterMinusOne); got != 2 {
		t.Fatalf("Reaping Willow entered with %d -1/-1 counters, want 2", got)
	}
	small := uuid.New()
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})

	err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{
		CounterKind: game.CounterMinusOne,
		Targets:     []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a mana value 4 card: %v, want ErrIllegalTarget", err)
	}
	if got := countersOn(g, willow, game.CounterMinusOne); got != 2 {
		t.Fatalf("a refused activation removed counters (%d left)", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, willow, 0, game.ActivateAbilityParams{
		CounterKind: game.CounterMinusOne,
		Targets:     []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldCardFor(g, small) == nil {
		t.Error("the mana value 3 creature card did not return to the battlefield")
	}
}
