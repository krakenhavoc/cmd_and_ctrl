package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emet_selch_of_the_third_seat_test.go — Emet-Selch of the Third Seat
// (#3785, deck request #3657).

const (
	emetSelchTestOracle = "634475d9-a1d7-4146-9a37-645d3d162af1"
	cauldronFamiliarOra = "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e"
)

func emetSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	emet := pushDiesCreatureForTest(g, me.ID, "Emet-Selch of the Third Seat", emetSelchTestOracle,
		"Legendary Creature — Elder Wizard", 3, 4)
	toMainForCost(t, g)
	return g, me, opp, emet
}

// emetOpponentLoses makes opp lose one life outside any resolution.
func emetOpponentLoses(t *testing.T, g *game.Game, emet uuid.UUID, opp *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(emet, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
}

// emetOffer drains opp once, picks target for the trigger and resolves
// it, leaving the cast offer open.
func emetOffer(t *testing.T, g *game.Game, me, opp *game.Player, emet, target uuid.UUID) {
	t.Helper()
	emetOpponentLoses(t, g, emet, opp)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
}

func emetCast(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
	})
}

// A loss offers the cast at the card's mana cost less {2}; the spell
// is exiled instead of going to the graveyard.
func TestEmetSelchCastsTheTargetAtItsCostLessTwoAndExilesIt(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	emetOffer(t, g, me, opp, emet, div)

	if err := emetCast(g, me, div); err == nil {
		t.Fatal("the cast is not free: Divination costs {U} after the discount")
	}
	floatForTest(g, me, "U")
	handBefore := len(me.Hand.Cards)
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination for {U}: %v", err)
	}
	// #4250: cast during the trigger's resolution.
	castInsideTheResolution(t, g, div)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(div) || !inExile(g, div) {
		t.Error("Divination goes to exile instead of the graveyard")
	}
	if got := len(me.Hand.Cards) - handBefore; got != 2 {
		t.Errorf("drew %d, want Divination's two", got)
	}
}

// The discount reads the zone the cast comes from: a hand cast pays
// the full price.
func TestEmetSelchDiscountsOnlyGraveyardCasts(t *testing.T) {
	g, me, _, _ := emetSetup(t)
	div := handCardFull(me, "Divination", "Sorcery", "{2}{U}", divinationOracle, []string{"U"})

	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a hand cast of a {2}{U} spell was discounted")
	}
	floatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for the full {2}{U}: %v", err)
	}
}

// "Do this only once each turn" counts the cast: after one is made a
// later loss puts nothing on the stack.
func TestEmetSelchOffersOnlyOnceEachTurn(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	emetOffer(t, g, me, opp, emet, div)
	floatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	emetOpponentLoses(t, g, emet, opp)
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, emet) != nil || latestPickTarget(g, me.ID) != nil {
		t.Fatal("a second trigger fired the turn the first cast was made")
	}
}

// A declined offer spends nothing: the decline (#4250, the resolution_cast's
// empty answer) closes it, and the next loss offers again.
func TestEmetSelchDeclinedOfferLeavesTheTurnsUse(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	emetOffer(t, g, me, opp, emet, div)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, div)
	declineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(div) {
		t.Fatal("the declined card left the graveyard")
	}
	floatForTest(g, me, "U")
	if err := emetCast(g, me, div); err == nil {
		t.Fatal("the lapsed offer still allowed a cast")
	}

	emetOffer(t, g, me, opp, emet, div)
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("a declined offer must not spend the turn's use: %v", err)
	}
}

// A card with its own flashback is cast for its MANA cost here, and
// the cast still spends the turn's use.
func TestEmetSelchCastOfAFlashbackCardCountsAndPaysManaCost(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	souls := gyCard(me, "Lingering Souls", "Sorcery", "{2}{W}", gyThemeLingeringSoulsOracle)
	// A second legal target, so a trigger that wrongly fires has one.
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	emetOffer(t, g, me, opp, emet, souls)
	floatForTest(g, me, "W")
	if err := emetCast(g, me, souls); err != nil {
		t.Fatalf("cast Lingering Souls for {W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, souls) {
		t.Error("the spell is exiled")
	}

	emetOpponentLoses(t, g, emet, opp)
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, emet) != nil || latestPickTarget(g, me.ID) != nil {
		t.Fatal("the cast of a flashback card did not spend the turn's use")
	}
}

// Every opponent losing life at once is ONE trigger (CR 603.2c).
func TestEmetSelchOneOrMoreOpponentsIsOneTrigger(t *testing.T) {
	g, me, _, emet := emetSetup(t)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	castCatalogSpell(t, g, "Cauldron Familiar", "Creature — Bat Skeleton", cauldronFamiliarOra, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 4 && triggerOnStack(g, emet) == nil && latestPickTarget(g, me.ID) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == emet {
			n++
		}
	}
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == me.ID {
			prompts++
		}
	}
	if n+prompts != 1 {
		t.Fatalf("three opponents lost life in one resolution: %d triggers + %d prompts, want exactly one", n, prompts)
	}
}

// The deck's own life gain or loss to the controller does not trigger.
func TestEmetSelchIgnoresItsControllersLifeLoss(t *testing.T) {
	g, me, _, emet := emetSetup(t)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(emet, me.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, emet) != nil || latestPickTarget(g, me.ID) != nil {
		t.Fatal("Emet-Selch triggered on its controller's own life loss")
	}
}

// Only an instant or sorcery card is a legal target.
func TestEmetSelchTargetsOnlyInstantsAndSorceries(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")

	emetOpponentLoses(t, g, emet, opp)
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, emet) != nil || latestPickTarget(g, me.ID) != nil {
		t.Fatal("a trigger with no legal target went on the stack (CR 603.3d)")
	}
}

// #3863: "If that spell would be put into your graveyard, exile it
// instead" replaces a graveyard destination on every path, a counter
// included: a countered spell is exiled.
func TestEmetSelchExilesTheSpellWhenItIsCountered(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	emetOffer(t, g, me, opp, emet, div)
	floatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(div); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if me.Graveyard.Contains(div) || !inExile(g, div) {
		t.Error("the countered spell goes to exile instead of the graveyard")
	}
}

// #3863: and ONLY a graveyard destination. A spell returned to its
// owner's hand goes there; flashback's wider "anywhere else" would
// have exiled it.
func TestEmetSelchLetsABouncedSpellGoToHand(t *testing.T) {
	g, me, opp, emet := emetSetup(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	emetOffer(t, g, me, opp, emet, div)
	floatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ReturnSpellToHandForEffect(div); err != nil {
			t.Fatalf("return to hand: %v", err)
		}
	})
	if inExile(g, div) || !me.Hand.Contains(div) {
		t.Error("the spell returned to its owner's hand was exiled")
	}
}
