package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// room_granted_alt_cost_test.go — #3174. Warped Space and Access Maze
// are granted alternative costs behind a Room door: offered only while
// the door is unlocked, once each turn per Room, and (Access Maze) for
// life equal to the spell's mana value during its controller's turn.

// pushRoom puts a Room on the battlefield under `owner` with the given
// doors unlocked.
func pushRoom(g *game.Game, owner uuid.UUID, oracle, leftName, leftCost, rightName, rightCost string, unlocked game.DoorMask) uuid.UUID {
	c := roomsBCard(owner, oracle, leftName, leftCost, rightName, rightCost)
	c.Unlocked = unlocked
	return pushBattlefieldCardWithTimestamp(g, c)
}

func pushWarpedSpace(g *game.Game, owner uuid.UUID, unlocked game.DoorMask) uuid.UUID {
	return pushRoom(g, owner, charredFoyerOracle, "Charred Foyer", "{3}{R}", "Warped Space", "{4}{R}{R}", unlocked)
}

func pushAccessMaze(g *game.Game, owner uuid.UUID, unlocked game.DoorMask) uuid.UUID {
	return pushRoom(g, owner, crampedVentsOracle, "Cramped Vents", "{3}{B}", "Access Maze", "{5}{B}{B}", unlocked)
}

func unlockDoorForTest(t *testing.T, g *game.Game, room uuid.UUID, door game.DoorSide, actor uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.UnlockDoorForEffect(room, door, actor); err != nil {
			t.Fatalf("unlock: %v", err)
		}
	})
}

// impulseExileCost exiles a sorcery with the given mana cost off the
// active seat's library with "you may play it this turn".
func impulseExileCost(t *testing.T, g *game.Game, name, cost string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the card is not in exile")
	}
	return me, id
}

func handSorcery(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}

func TestWarpedSpaceOffersAFreeCastFromExileOnlyWhileUnlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, first := impulseExileCost(t, g, "Test Bolt", "{2}{R}")
	advanceToMainForTest(t, g)
	room := pushWarpedSpace(g, me.ID, game.DoorLeftUnlocked)

	key := GrantedAltCostZeroFromExileOnceEachTurn
	if offerWithKey(grantedCardOffers(g, me.ID, first, game.ZoneExile), key) != nil {
		t.Fatal("a locked Warped Space offered its free cast")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err == nil {
		t.Fatal("the cast gate accepted a claim of a locked Warped Space")
	}

	unlockDoorForTest(t, g, room, game.DoorRight, me.ID)
	offer := offerWithKey(grantedCardOffers(g, me.ID, first, game.ZoneExile), key)
	if offer == nil {
		t.Fatal("an unlocked Warped Space did not offer its free cast from exile")
	}
	if offer.ManaCost != "" || offer.Life != 0 {
		t.Errorf("offer charges %q and %d life, want nothing", offer.ManaCost, offer.Life)
	}
	// A hand card is not cast from exile.
	inHand := handSorcery(me, "Hand Sorcery", "{1}{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, inHand, game.ZoneHand), key) != nil {
		t.Error("Warped Space offered its free cast to a spell from hand")
	}

	// Strict, with an empty pool: the free cast is the only way to pay.
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	if g.Exile.Contains(first) {
		t.Fatal("the spell is still in exile")
	}
	passPriorityAroundTable(t, g)

	// Once each turn: a second exiled spell this turn is not offered it.
	_, second := impulseExileCost(t, g, "Test Shock", "{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, second, game.ZoneExile), key) != nil {
		t.Fatal("Warped Space offered a second free cast in one turn")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err == nil {
		t.Fatal("the cast gate accepted a second claim in one turn")
	}

	// The limit is per Warped Space: a second one has its own claim.
	pushWarpedSpace(g, me.ID, game.DoorLeftUnlocked|game.DoorRightUnlocked)
	if offerWithKey(grantedCardOffers(g, me.ID, second, game.ZoneExile), key) == nil {
		t.Fatal("a second Warped Space did not offer its own free cast")
	}
}

func TestWarpedSpaceClaimComesBackNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, first := impulseExileCost(t, g, "Test Bolt", "{2}{R}")
	advanceToMainForTest(t, g)
	pushWarpedSpace(g, me.ID, game.DoorRightUnlocked)
	key := GrantedAltCostZeroFromExileOnceEachTurn
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Round the table to my next turn: the claim is fresh.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
	_, next := impulseExileCost(t, g, "Test Shock", "{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, next, game.ZoneExile), key) == nil {
		t.Fatal("Warped Space's once-each-turn claim did not come back on a new turn")
	}
}

func TestAccessMazeCastsFromHandForLifeEqualToManaValue(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	toMain(t, g)
	room := pushAccessMaze(g, me.ID, game.DoorLeftUnlocked)
	key := GrantedAltCostLifeFromHandOnceYourTurn
	three := handSorcery(me, "Three Sorcery", "{1}{B}{B}")
	if offerWithKey(grantedCardOffers(g, me.ID, three, game.ZoneHand), key) != nil {
		t.Fatal("a locked Access Maze offered its life cast")
	}

	unlockDoorForTest(t, g, room, game.DoorRight, me.ID)
	offer := offerWithKey(grantedCardOffers(g, me.ID, three, game.ZoneHand), key)
	if offer == nil {
		t.Fatal("an unlocked Access Maze did not offer its life cast")
	}
	if offer.Life != 3 || offer.ManaCost != "" {
		t.Fatalf("offer charges %q and %d life, want 3 life and no mana", offer.ManaCost, offer.Life)
	}
	if !strings.Contains(offer.Label, "3 life") {
		t.Errorf("label %q does not name the life", offer.Label)
	}

	life := me.Life
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("life cast from hand: %v", err)
	}
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	passPriorityAroundTable(t, g)

	// Once during each of your turns.
	two := handSorcery(me, "Two Sorcery", "{1}{B}")
	if offerWithKey(grantedCardOffers(g, me.ID, two, game.ZoneHand), key) != nil {
		t.Fatal("Access Maze offered a second life cast this turn")
	}
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Fatal("the cast gate accepted a second life cast this turn")
	}

	// Not during an opponent's turn, though the claim is fresh there.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if offerWithKey(grantedCardOffers(g, me.ID, two, game.ZoneHand), key) != nil {
		t.Fatal("Access Maze offered its life cast during an opponent's turn")
	}
	// Back on my turn, once again.
	advanceToMainOf(t, g, seat)
	if offerWithKey(grantedCardOffers(g, me.ID, two, game.ZoneHand), key) == nil {
		t.Fatal("Access Maze's claim did not come back on my next turn")
	}
}

func TestAccessMazeNeedsTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushAccessMaze(g, me.ID, game.DoorRightUnlocked)
	key := GrantedAltCostLifeFromHandOnceYourTurn
	five := handSorcery(me, "Five Sorcery", "{3}{B}{B}")
	me.Life = 4
	if err := g.CastSpell(me.ID, five, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Fatal("a cast for 5 life was accepted at 4 life")
	}
	if me.Life != 4 || !me.Hand.Contains(five) {
		t.Fatalf("the refused cast changed the table: life %d, in hand %v", me.Life, me.Hand.Contains(five))
	}
	// The refused cast spent nothing: at 5 life it may be paid (CR
	// 119.4 lets a player pay life down to 0).
	me.Life = 5
	if err := g.CastSpell(me.ID, five, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("a cast for 5 life at 5 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life %d, want 0", me.Life)
	}
}

func TestRoomGrantedAltCostRowsFollowTheDoor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	room := pushAccessMaze(g, me.ID, game.DoorLeftUnlocked)
	row := func() bool {
		c := findBattlefieldCardForTest(g, room)
		for _, r := range game.AbilityRowsOnBattlefieldPtr(c) {
			if strings.HasPrefix(r.Label, "Once during each of your turns, you may cast spells from your hand by paying life") {
				return true
			}
		}
		return false
	}
	if row() {
		t.Fatal("a locked Access Maze shows its row")
	}
	unlockDoorForTest(t, g, room, game.DoorRight, me.ID)
	if !row() {
		t.Fatalf("an unlocked Access Maze does not show its row: %+v", game.AbilityRowsOnBattlefieldPtr(findBattlefieldCardForTest(g, room)))
	}
}
