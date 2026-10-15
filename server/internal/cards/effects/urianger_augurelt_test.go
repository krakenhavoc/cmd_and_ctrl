package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// urianger_augurelt_test.go — #3787: cards exiled face down with a
// source, a temporary play grant over them that takes {2} off the
// spells cast through it, and the play-from-exile trigger. Kayla's Music
// Box rides the same two primitives and is tested at the end.

const (
	uriangerOracle = "7ca584e7-4621-4b4e-a381-a0051acbf66a"
	kaylasOracle   = "8110f15d-82db-4299-a959-09c22475daec"

	uriangerDraw = 0
	uriangerPlay = 1
)

func uriangerTable(t *testing.T) (g *game.Game, me, opp *game.Player, uri uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	uri = pushCatalogPermanent(g, me.ID, "Urianger Augurelt", "Legendary Creature — Elf Advisor", uriangerOracle, false)
	return
}

func objectRefOf(t *testing.T, g *game.Game, id uuid.UUID) game.PermissionCardRef {
	t.Helper()
	var ref game.PermissionCardRef
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s not found", id)
		}
		ref = game.ObjectRefOf(c)
	})
	return ref
}

// topCard puts one named card of the given type on top of p's library.
func topCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

func activateArcanum(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
	passPriorityAroundTable(t, g)
}

func untapSource(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
}

// drawArcanum runs Draw Arcanum over a fresh top card and exiles it.
func drawArcanum(t *testing.T, g *game.Game, me *game.Player, uri uuid.UUID, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := topCard(me, name, typeLine, cost)
	activateArcanum(t, g, me, uri, uriangerDraw)
	answerChooseCards(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	untapSource(g, uri)
	return id
}

// Draw Arcanum: a "look", then a "you may exile it face down". The card
// leaves the library, lands in exile hidden by and linked to THIS
// Urianger, and only its controller may look at it.
func TestUriangerDrawArcanumExilesTheTopCardFaceDownLinked(t *testing.T) {
	g, me, opp, uri := uriangerTable(t)
	size := me.Library.Size()
	id := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{2}{G}")

	c, ok := exileCard(g, id)
	if !ok {
		t.Fatal("the card was not exiled")
	}
	want := objectRefOf(t, g, uri)
	if c.ExiledWith != want {
		t.Errorf("ExiledWith = %+v, want Urianger the object %+v", c.ExiledWith, want)
	}
	if c.HiddenBy != want || c.FaceDownKind != game.FaceDownHidden || !c.FaceDown {
		t.Errorf("not exiled face down under Urianger: hiddenBy %+v kind %q down %v", c.HiddenBy, c.FaceDownKind, c.FaceDown)
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("knowers: me %v, opponent %v — want the controller only", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
	}
	if got := me.Library.Size(); got != size {
		t.Errorf("library = %d, want %d (one pushed, one exiled)", got, size)
	}
}

// Hidden information: the opponent's wire view of the exile zone names
// nothing about the face-down card, and the controller's does.
func TestUriangerFaceDownCardIsHiddenFromOpponents(t *testing.T) {
	g, me, opp, uri := uriangerTable(t)
	id := drawArcanum(t, g, me, uri, "Secret Dragon", "Creature — Dragon", "{4}{R}{R}")

	find := func(viewer uuid.UUID) protocol.CardView {
		v := protocol.ViewOfGameFor(g, viewer.String())
		for _, c := range v.Exile.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("card missing from %s's exile view", viewer)
		return protocol.CardView{}
	}
	if own := find(me.ID); own.Name != "Secret Dragon" {
		t.Errorf("the controller sees %q, want the card's name", own.Name)
	}
	if other := find(opp.ID); other.Name != "" || other.CastableHere {
		t.Errorf("the opponent sees name %q castable %v, want a blank face-down card", other.Name, other.CastableHere)
	}
}

// "You MAY exile it": declining leaves the card on top of the library
// and exiles nothing.
func TestUriangerDrawArcanumMayDecline(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	id := topCard(me, "Kept Card", "Sorcery", "{1}")
	activateArcanum(t, g, me, uri, uriangerDraw)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(id) {
		t.Error("a declined card was exiled")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != id {
		t.Error("the declined card is no longer on top of the library")
	}
}

// Play Arcanum: the exiled creature is castable, and {2} cheaper. A
// {2}{G} bear is paid with {G}; without the grant {G} is not enough.
func TestUriangerPlayArcanumCastsAnExiledSpellForTwoLess(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	bear := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{2}{G}")

	if enumeratedFor(g, me.ID, bear) {
		t.Fatal("non-vacuity: the card is castable before Play Arcanum")
	}
	activateArcanum(t, g, me, uri, uriangerPlay)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if !enumeratedFor(g, me.ID, bear) || !viewCastable(t, g, me.ID, bear) {
		t.Fatal("the enumerator and the view must offer the exiled card with only {G} in the pool")
	}
	var price game.CastPrice
	var err error
	g.ReadSnapshot(func() {
		c, _ := exileCard(g, bear)
		price, err = g.PriceCastForEffect(me.ID, c, game.CastSpellParams{FromZone: "exile"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := price.Total.String(); got != "{G}" {
		t.Errorf("price = %s, want {G}", got)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Fatal("the creature did not resolve")
	}
}

// The reduction is generic only and never below zero: a {1} spell costs
// nothing, a {G}{G} spell is untouched.
func TestUriangerReductionIsGenericAndStopsAtZero(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	cheap := drawArcanum(t, g, me, uri, "Cheap Spell", "Instant", "{1}")
	coloured := drawArcanum(t, g, me, uri, "Green Spell", "Instant", "{G}{G}")
	activateArcanum(t, g, me, uri, uriangerPlay)

	priceOf := func(id uuid.UUID) string {
		var price game.CastPrice
		var err error
		g.ReadSnapshot(func() {
			c, _ := exileCard(g, id)
			price, err = g.PriceCastForEffect(me.ID, c, game.CastSpellParams{FromZone: "exile"})
		})
		if err != nil {
			t.Fatal(err)
		}
		return price.Total.String()
	}
	if got := priceOf(cheap); got != "{0}" && got != "" {
		t.Errorf("{1} spell costs %q, want nothing", got)
	}
	if got := priceOf(coloured); got != "{G}{G}" {
		t.Errorf("{G}{G} spell costs %s, want {G}{G}", got)
	}
}

// CR 601.2f: the board's increases are added before the permission's
// reduction comes off. A {1}{G} spell under a Sphere of Resistance is
// {2}{G}, and {2} less leaves {G}, not {G} + the tax.
func TestUriangerReductionComesOffAfterTheIncreases(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	pushCatalogPermanent(g, me.ID, "Sphere of Resistance", "Artifact", "09c96077-3804-4f12-a613-5bebc5e0413f", false)
	bear := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{1}{G}")
	activateArcanum(t, g, me, uri, uriangerPlay)

	var price game.CastPrice
	var err error
	g.ReadSnapshot(func() {
		c, _ := exileCard(g, bear)
		price, err = g.PriceCastForEffect(me.ID, c, game.CastSpellParams{FromZone: "exile"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := price.Total.String(); got != "{G}" {
		t.Errorf("price = %s, want {G} ({1}{G} + {1} tax - {2})", got)
	}
}

// "Play": a land card is played with the turn's land drop, and
// "whenever you play a land from exile" gains 2 life.
func TestUriangerPlaysAnExiledLandAndGainsTwoLife(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	forest := drawArcanum(t, g, me, uri, "Forest", "Basic Land — Forest", "")
	activateArcanum(t, g, me, uri, uriangerPlay)

	life := me.Life
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if _, ok := battlefieldCard(g, forest); !ok {
		t.Fatal("the land is not on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("life gained = %d, want 2", got)
	}
}

// "...or cast a spell from exile" gains 2 life too, and a spell cast
// from hand does not.
func TestUriangerGainsTwoLifeForASpellCastFromExileOnly(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	bear := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{G}")
	activateArcanum(t, g, me, uri, uriangerPlay)

	handBear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: handBear, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, handBear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("a spell cast from hand gained %d life", me.Life-life)
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("life gained = %d, want 2", got)
	}
}

// The permission is the controller's alone, only over cards exiled WITH
// Urianger, and only for the turn.
func TestUriangerPermissionReachesOnlyItsOwnCardsForTheControllerThisTurn(t *testing.T) {
	g, me, opp, uri := uriangerTable(t)
	mine := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{G}")

	// A card exiled face up by something else is not "exiled with Urianger".
	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Other Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(other); err != nil {
			t.Fatal(err)
		}
	})
	activateArcanum(t, g, me, uri, uriangerPlay)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}

	if !enumeratedFor(g, me.ID, mine) {
		t.Fatal("non-vacuity: the exiled-with card is not offered")
	}
	if enumeratedFor(g, me.ID, other) {
		t.Error("a card exiled without the link was offered")
	}
	if enumeratedFor(g, opp.ID, mine) {
		t.Error("the opponent was offered the controller's face-down card")
	}
	if err := g.CastSpell(opp.ID, mine, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("the opponent cast the controller's face-down card")
	}

	advanceToNextSeatsTurn(t, g)
	if enumeratedFor(g, me.ID, mine) {
		t.Error("the permission outlived the turn")
	}
}

// CR 607.2a / 400.7: a Urianger that left and came back is a new object
// and has no claim on what its earlier self exiled.
func TestUriangerThatReturnedHasNoClaimOnTheOldCards(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	old := drawArcanum(t, g, me, uri, "Old Card", "Creature — Bear", "{G}")

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(uri); err != nil {
			t.Fatal(err)
		}
	})
	uri2 := pushCatalogPermanent(g, me.ID, "Urianger Augurelt", "Legendary Creature — Elf Advisor", uriangerOracle, false)
	activateArcanum(t, g, me, uri2, uriangerPlay)

	if enumeratedFor(g, me.ID, old) {
		t.Error("a new Urianger was granted play over the old one's card")
	}
}

// The grant is read when the card is played, so a card exiled AFTER
// Play Arcanum resolved (an untap effect) is playable the same turn.
func TestUriangerPlaysACardExiledAfterPlayArcanum(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	activateArcanum(t, g, me, uri, uriangerPlay)
	untapSource(g, uri)
	late := drawArcanum(t, g, me, uri, "Late Card", "Creature — Bear", "{G}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if !enumeratedFor(g, me.ID, late) {
		t.Fatal("a card exiled after Play Arcanum is not playable")
	}
}

// The bot's enumeration prices the discount: a bot with {G} offers the
// {2}{G} creature, and without the permission it does not.
func TestUriangerEnumerationPricesTheDiscount(t *testing.T) {
	g, me, _, uri := uriangerTable(t)
	bear := drawArcanum(t, g, me, uri, "Hidden Bear", "Creature — Bear", "{2}{G}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == bear {
			t.Fatal("offered before Play Arcanum with one mana")
		}
	}
	activateArcanum(t, g, me, uri, uriangerPlay)
	found := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == bear {
			found = true
		}
	}
	if !found {
		t.Fatal("not offered after Play Arcanum with {G} in the pool")
	}
}

// --- Kayla's Music Box -------------------------------------------------

func kaylasTable(t *testing.T) (g *game.Game, me, opp *game.Player, box uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	box = pushCatalogPermanent(g, me.ID, "Kayla's Music Box", "Legendary Artifact", kaylasOracle, false)
	return
}

// "Look at the top card of your library, then exile it face down": no
// prompt, the card is exiled linked and hidden.
func TestKaylasMusicBoxExilesTheTopCardWithoutAsking(t *testing.T) {
	g, me, opp, box := kaylasTable(t)
	id := topCard(me, "Boxed", "Creature — Bear", "{G}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	activateArcanum(t, g, me, box, 0)

	c, ok := exileCard(g, id)
	if !ok {
		t.Fatal("the top card was not exiled")
	}
	if c.ExiledWith != objectRefOf(t, g, box) || c.FaceDownKind != game.FaceDownHidden {
		t.Errorf("not linked and hidden: %+v kind %q", c.ExiledWith, c.FaceDownKind)
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("only the controller may look at it")
	}
}

// "Cards YOU OWN exiled with Kayla's Music Box": a card the Box exiled
// that someone else owns (it changed hands) is not playable.
func TestKaylasMusicBoxPlaysOnlyCardsTheHolderOwns(t *testing.T) {
	g, me, opp, box := kaylasTable(t)
	mine := topCard(me, "Mine", "Creature — Bear", "{G}")
	theirs := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Creature — Bear", ManaCost: "{G}", Owner: opp.ID, Controller: opp.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	activateArcanum(t, g, me, box, 0)
	untapSource(g, box)
	ref := objectRefOf(t, g, box)
	g.WithWriteLock(func() {
		if err := g.ExileLinkedForEffect(game.ObjectRef(ref), theirs); err != nil {
			t.Fatal(err)
		}
	})

	activateArcanum(t, g, me, box, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if !enumeratedFor(g, me.ID, mine) {
		t.Fatal("non-vacuity: the Box's own card is not offered")
	}
	if enumeratedFor(g, me.ID, theirs) {
		t.Error("a card the holder does not own was offered")
	}
}
