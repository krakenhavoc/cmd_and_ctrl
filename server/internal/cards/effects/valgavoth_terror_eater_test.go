package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// valgavoth_terror_eater_test.go — #2530: "during your turn, you may
// play cards exiled with Valgavoth. If you cast a spell this way, pay
// life equal to its mana value rather than pay its mana cost."

const valgavothAltKey = "valgavoth_terror_eater"

// valgavothTable is seat 0's main phase with Valgavoth on the
// battlefield under seat 0.
func valgavothTable(t *testing.T) (g *game.Game, me, opp *game.Player, val uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	val = pushCatalogPermanent(g, me.ID, "Valgavoth, Terror Eater", "Legendary Creature — Elder Demon", valgavothOracle, false)
	return
}

// valgavothEat puts an opponent's creature on the battlefield and
// destroys it, so Valgavoth's replacement exiles it and links it.
func valgavothEat(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Exile.Contains(id) {
		t.Fatalf("%s was not exiled by Valgavoth's replacement", name)
	}
	return id
}

func castForLife(g *game.Game, who *game.Player, id uuid.UUID) error {
	return g.CastSpell(who.ID, id, game.CastSpellParams{FromZone: "exile", AlternativeCost: valgavothAltKey})
}

func cardInExile(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := exileCard(g, id)
	if !ok {
		t.Fatalf("card %s is not in exile", id)
	}
	return c
}

// The replacement links what it exiles to the Valgavoth that did it, as
// that object, and the controller casts it for life equal to its mana
// value — no mana.
func TestValgavothPlaysAnExiledCardForLife(t *testing.T) {
	g, me, opp, val := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")

	var want game.PermissionCardRef
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(val)
		want = game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	})
	if got := cardInExile(t, g, bear).ExiledWith; got != want {
		t.Fatalf("ExiledWith = %+v, want Valgavoth the object %+v", got, want)
	}
	if !enumeratedFor(g, me.ID, bear) || !viewCastable(t, g, me.ID, bear) {
		t.Fatal("non-vacuity: the enumerator and the view must offer the eaten card")
	}

	lifeBefore := me.Life
	if err := castForLife(g, me, bear); err != nil {
		t.Fatalf("cast the exiled creature for life: %v", err)
	}
	if got := lifeBefore - me.Life; got != 2 {
		t.Errorf("life paid = %d, want 2 (its mana value)", got)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature did not resolve onto the battlefield")
	}
	if got.Controller != me.ID {
		t.Error("the cast creature should be under the caster's control")
	}
	if got.ExiledWith.ID != uuid.Nil {
		t.Errorf("the link followed the card out of exile: %+v", got.ExiledWith)
	}
}

// "Rather than pay its mana cost": a cast that does not claim the life
// offer is not on offer at all, even with the mana in the pool.
func TestValgavothCannotPayTheManaCostInstead(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a card exiled with Valgavoth was cast for its printed mana cost")
	}
}

// A card the same replacement exiles from a hand (a discard) is linked
// too — the route is a different function from a battlefield exit, and
// both have to stamp it.
func TestValgavothLinksADiscardedCard(t *testing.T) {
	g, me, opp, val := valgavothTable(t)
	opp.Hand.Cards = nil
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Discarded Sorcery", TypeLine: "Sorcery",
		ManaCost: "{2}{B}", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Exile.Contains(id) || opp.Graveyard.Contains(id) {
		t.Fatal("the discard was not exiled instead")
	}
	if got := cardInExile(t, g, id).ExiledWith.ID; got != val {
		t.Fatalf("a discarded card's ExiledWith = %s, want Valgavoth %s", got, val)
	}
	g.Turn.Step = game.StepPrecombatMain
	lifeBefore := me.Life
	if err := castForLife(g, me, id); err != nil {
		t.Fatalf("cast the discarded sorcery for life: %v", err)
	}
	if got := lifeBefore - me.Life; got != 3 {
		t.Errorf("life paid = %d, want 3", got)
	}
}

// A card that is in exile for any OTHER reason is not exiled with
// Valgavoth, and a card Valgavoth's controller did not lose to the
// replacement is never linked.
func TestValgavothIgnoresCardsExiledByAnythingElse(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	elsewhere := stashed(g, opp.ID, "Plain Exile", "Creature — Bear", "{1}{G}", 0)
	if enumeratedFor(g, me.ID, elsewhere) || viewCastable(t, g, me.ID, elsewhere) {
		t.Error("enumerator or view offers a card nothing linked to Valgavoth")
	}
	if err := castForLife(g, me, elsewhere); err == nil {
		t.Fatal("a card not exiled with Valgavoth was cast through it")
	}
}

// Lands are played, not cast: no life, and the turn's land drop is spent.
func TestValgavothPlaysAnExiledLandForFree(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	land := valgavothEat(t, g, opp.ID, "Eaten Land", "Land", "")
	lifeBefore := me.Life
	dropsBefore := g.LandDropsRemainingFor(me.ID)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if me.Life != lifeBefore {
		t.Errorf("a land cost %d life", lifeBefore-me.Life)
	}
	if got, ok := battlefieldCard(g, land); !ok || got.Controller != me.ID {
		t.Fatal("the land did not enter under the player's control")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != dropsBefore-1 {
		t.Errorf("land drops %d -> %d, want one spent", dropsBefore, got)
	}
}

// CR 119.4: life is a cost, so a player who cannot pay it cannot claim.
func TestValgavothRefusesACastYouCannotPayLifeFor(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	big := valgavothEat(t, g, opp.ID, "Eaten Colossus", "Creature — Giant", "{9}{G}")
	g.WithWriteLock(func() { me.Life = 4 })
	err := castForLife(g, me, big)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast at 4 life for mana value 10: got %v, want ErrInvalidParam", err)
	}
	if me.Life != 4 || !g.Exile.Contains(big) {
		t.Errorf("a refused cast changed the game: life %d, still in exile %v", me.Life, g.Exile.Contains(big))
	}
}

// "During your turn": off the holder's turn nothing is playable, not
// even an instant.
func TestValgavothOnlyWorksDuringYourTurn(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	bolt := valgavothEat(t, g, opp.ID, "Eaten Bolt", "Instant", "{R}")
	if !enumeratedFor(g, me.ID, bolt) {
		t.Fatal("non-vacuity: an instant is offered on your own turn")
	}
	g.Turn.ActiveSeat = 1
	if err := castForLife(g, me, bolt); err == nil {
		t.Fatal("cast succeeded off the holder's turn")
	}
	if enumeratedFor(g, me.ID, bolt) || viewCastable(t, g, me.ID, bolt) {
		t.Error("enumerator or view offers a cast off the holder's turn")
	}
}

// Another seat cannot play what Valgavoth's controller's Valgavoth ate.
func TestValgavothGrantsOnlyItsController(t *testing.T) {
	g, _, opp, _ := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")
	// The opponent owns the card but holds no permission: even on their
	// own turn, with mana, it is not castable from exile.
	g.Turn.ActiveSeat = 1
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(opp.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the card's owner cast it from exile with no permission")
	}
	if err := castForLife(g, opp, bear); err == nil {
		t.Fatal("the card's owner claimed Valgavoth's life offer")
	}
}

// The permission is the permanent's, so it ends when Valgavoth does,
// and a Valgavoth that comes back is a new object with no claim on what
// the old one ate (CR 400.7).
func TestValgavothLeavingEndsThePermissionForGood(t *testing.T) {
	g, me, opp, val := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")
	if !enumeratedFor(g, me.ID, bear) {
		t.Fatal("non-vacuity: offered while Valgavoth is in play")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(val); err != nil {
			t.Fatal(err)
		}
	})
	if err := castForLife(g, me, bear); err == nil {
		t.Fatal("cast succeeded after Valgavoth left")
	}
	if enumeratedFor(g, me.ID, bear) || viewCastable(t, g, me.ID, bear) {
		t.Error("enumerator or view still offers the card after Valgavoth left")
	}
}

// The SAME card returning is the case an instance ID alone cannot see:
// bounced to hand and replayed, Valgavoth keeps its InstanceID and is a
// new object (CR 400.7), so the epoch in the link is what ends the claim.
func TestValgavothBouncedAndReplayedHasNoClaimOnOldCards(t *testing.T) {
	g, me, opp, val := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(val); err != nil {
			t.Fatal(err)
		}
	})
	if !me.Hand.Contains(val) {
		t.Fatal("Valgavoth was not bounced to hand")
	}
	card, err := me.Hand.Remove(val)
	if err != nil {
		t.Fatal(err)
	}
	pushBattlefieldCardWithTimestamp(g, card)
	if !g.Battlefield.Contains(val) {
		t.Fatal("non-vacuity: the same Valgavoth card is back on the battlefield")
	}
	if err := castForLife(g, me, bear); err == nil {
		t.Fatal("a replayed Valgavoth reached the cards its earlier object exiled")
	}
	if enumeratedFor(g, me.ID, bear) || viewCastable(t, g, me.ID, bear) {
		t.Error("enumerator or view offers a card exiled with a previous Valgavoth object")
	}
	// And what the NEW object eats is playable through it.
	fresh := valgavothEat(t, g, opp.ID, "Fresh Bear", "Creature — Bear", "{1}{G}")
	if !enumeratedFor(g, me.ID, fresh) {
		t.Error("the new object's own exiled card is not offered")
	}
}

// The cast is on the wire as the one life offer, so the client and the
// bot price it the way the engine does.
func TestValgavothOffersExactlyTheLifeCost(t *testing.T) {
	g, me, opp, _ := valgavothTable(t)
	bear := valgavothEat(t, g, opp.ID, "Eaten Bear", "Creature — Bear", "{1}{G}")
	v := protocol.ViewOfGameFor(g, me.ID.String())
	seen := false
	for _, c := range v.Exile.Cards {
		if c.InstanceID != bear.String() {
			continue
		}
		seen = true
		if len(c.AlternativeCosts) != 1 || c.AlternativeCosts[0].Key != valgavothAltKey || c.AlternativeCosts[0].Life != 2 {
			t.Errorf("view alternative costs = %+v, want the one 2-life offer", c.AlternativeCosts)
		}
	}
	if !seen {
		t.Error("the eaten card is not in the exile view")
	}
	found := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != bear {
			continue
		}
		found++
		var params struct {
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(m.Params, &params); err != nil {
			t.Fatal(err)
		}
		if params.AlternativeCost != valgavothAltKey {
			t.Errorf("enumerated cast claims %q, want %q", params.AlternativeCost, valgavothAltKey)
		}
	}
	if found == 0 {
		t.Error("the enumerator offers no cast of the eaten card")
	}
}
