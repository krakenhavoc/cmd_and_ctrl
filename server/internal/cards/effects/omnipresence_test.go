package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// omnipresence_test.go — #3503, CR 118.9. A free cast from hand whose
// mana value cap is the number of creatures the caster controls, read
// each time the offer is derived.

const omnipresenceOracle = "8dad1fa8-3533-4e3c-b9b3-fcd57b20a33d"

func pushOmniBear(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Owner: controller, Controller: controller,
	})
}

func freeOfferKey(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) bool {
	return offerWithKey(grantedCardOffers(g, seat, card, zone), GrantedAltCostFreeUpToCreatureCount) != nil
}

// The cap follows the board between snapshots: a three-drop is out of
// reach with two creatures, in reach with a third, and out again when
// one dies. The boundary (mana value equal to the count) is covered, and
// the cast gate agrees with the offer list every time.
func TestOmnipresenceCapFollowsTheCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omnipresence", TypeLine: "Enchantment",
		OracleID: omnipresenceOracle, Owner: me.ID, Controller: me.ID,
	})
	three := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: three, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Owner: me.ID, Controller: me.ID})

	if freeOfferKey(g, me.ID, three, game.ZoneHand) {
		t.Fatal("a mana value 3 spell was offered for free with no creatures")
	}
	pushOmniBear(g, me.ID)
	second := pushOmniBear(g, me.ID)
	if freeOfferKey(g, me.ID, three, game.ZoneHand) {
		t.Fatal("a mana value 3 spell was offered for free with two creatures")
	}
	err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFreeUpToCreatureCount})
	if err == nil {
		t.Fatal("the cast gate accepted the claim with two creatures against a mana value 3 spell")
	}
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Logf("refused with %v", err)
	}

	pushOmniBear(g, me.ID)
	if !freeOfferKey(g, me.ID, three, game.ZoneHand) {
		t.Fatal("a mana value 3 spell was not offered for free with three creatures (the cap is not following the board)")
	}

	// A creature dies: the cap falls with it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(second); err != nil {
			t.Fatalf("destroy a creature: %v", err)
		}
	})
	if freeOfferKey(g, me.ID, three, game.ZoneHand) {
		t.Fatal("the offer outlived the creature that paid for it")
	}
	pushOmniBear(g, me.ID)

	// Back to three: cast it, strict, with an empty pool.
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFreeUpToCreatureCount}); err != nil {
		t.Fatalf("free cast of a mana value 3 spell with three creatures: %v", err)
	}
	if me.Hand.Contains(three) {
		t.Error("the spell is still in hand")
	}
}

// "Creatures you control": an opponent's creatures do not count, and
// the grant is for the controller's own hand.
func TestOmnipresenceCountsOnlyYourCreaturesAndReachesOnlyYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omnipresence", TypeLine: "Enchantment",
		OracleID: omnipresenceOracle, Owner: me.ID, Controller: me.ID,
	})
	pushOmniBear(g, me.ID)
	pushOmniBear(g, opp.ID)
	pushOmniBear(g, opp.ID)
	pushOmniBear(g, opp.ID)

	two := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: two, Name: "Two Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}",
		Owner: me.ID, Controller: me.ID})
	one := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: one, Name: "One Sorcery", TypeLine: "Sorcery", ManaCost: "{U}",
		Owner: me.ID, Controller: me.ID})
	if freeOfferKey(g, me.ID, two, game.ZoneHand) {
		t.Error("the opponent's creatures were counted toward the cap")
	}
	if !freeOfferKey(g, me.ID, one, game.ZoneHand) {
		t.Error("a mana value 1 spell was not covered by one creature")
	}
	inYard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: inYard, Name: "Yard Sorcery", TypeLine: "Sorcery", ManaCost: "{U}",
		Owner: me.ID, Controller: me.ID})
	if freeOfferKey(g, me.ID, inYard, game.ZoneGraveyard) {
		t.Error("Omnipresence offered a free cast from a graveyard")
	}
	// The opponent has Omnipresence on my side of the table: nothing for them.
	theirs := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{U}",
		Owner: opp.ID, Controller: opp.ID})
	if freeOfferKey(g, opp.ID, theirs, game.ZoneHand) {
		t.Error("the opponent was offered my Omnipresence's free cast")
	}
}

// With no creatures a spell with mana value 0 is still covered (0 <= 0),
// the boundary of "less than or equal to".
func TestOmnipresenceCoversManaValueZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omnipresence", TypeLine: "Enchantment",
		OracleID: omnipresenceOracle, Owner: me.ID, Controller: me.ID,
	})
	zero := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: zero, Name: "Ancestral Test", TypeLine: "Sorcery", ManaCost: "", Layout: "normal",
		Owner: me.ID, Controller: me.ID})
	if !freeOfferKey(g, me.ID, zero, game.ZoneHand) {
		t.Error("a mana value 0 spell was not covered by a cap of 0")
	}
}
