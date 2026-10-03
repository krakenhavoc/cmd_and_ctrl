package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblem_targeting_cards_test.go — ADR 0109 Delivery PR 6, the card
// half: §5's emblem gates (#1899: Narset Transcendent, Dovin Baan) and
// §6's targeting restrictions (#1885: Ground Seal, Dennick, Silent
// Gravestone, Underworld Cerberus, Tomik). The bot enumerator's and the
// view's agreement is asserted in internal/legal and internal/protocol.

const (
	narsetOracle             = "e1de0c94-0ecc-425a-9b92-27c2745d07e7"
	dovinBaanOracle          = "5b33bfbf-e4e1-43de-9c96-ccdc2916510a"
	tomikOracle              = "9895a33f-9bbd-4440-8c1a-0d401431b77f"
	groundSealOracle         = "13f6f960-ef79-4f2c-8874-90fe9e77099e"
	silentGravestoneOracle   = "803815c5-12be-48b3-a101-f416c1ef9b7b"
	underworldCerberusOracle = "a36c2b76-d595-42ba-b8c2-bb8f02639981"
	regrowthOracle           = "e6e4a8bd-5c40-4654-8de1-0da9afed90fd"
)

// a109p6Seats is the active seat and the next one.
func a109p6Seats(g *game.Game) (*game.Player, *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// a109p6CastGate asks the cast gate whether p may cast `card` from hand.
func a109p6CastGate(g *game.Game, p *game.Player, card game.Card) error {
	var err error
	g.WithWriteLock(func() { err = g.CastGateLocked(p.ID, card, game.ZoneHand, game.CastSpellParams{}) })
	return err
}

// a109p6LegalCards is the legal card targets of `spec` for a spell or
// ability `by` controls.
func a109p6LegalCards(g *game.Game, by uuid.UUID, spec *game.TargetSpec) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() {
		out = g.LegalTargetsForEffect(game.TargetSource{Controller: by}, spec).Cards
	})
	return out
}

func a109p6Has(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// a109p6Walker puts a catalog planeswalker under p at a main phase.
func a109p6Walker(t *testing.T, g *game.Game, p *game.Player, name, oracle string, loyalty int) uuid.UUID {
	t.Helper()
	toMain(t, g)
	return pushCatalogWalker(g, p.ID, name, oracle, loyalty)
}

// --- §5: Narset Transcendent ----------------------------------------

// The −9's emblem refuses an opponent's noncreature spell with its
// clause, and nothing else: an opponent's creature spell, and Narset's
// controller's own spells, are cast as normal. The emblem outlives
// Narset.
func TestNarsetEmblemStopsOpponentsNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	narset := a109p6Walker(t, g, me, "Narset Transcendent", narsetOracle, 9)
	b16Activate(t, g, me.ID, narset, 2, game.ActivateAbilityParams{})
	if me.Emblems == nil || len(me.Emblems.Cards) != 1 {
		t.Fatal("the −9 made no emblem")
	}
	if g.Battlefield.Contains(narset) {
		t.Error("Narset at 0 loyalty is still on the battlefield")
	}

	instant := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	var cant *game.CantCastError
	if err := a109p6CastGate(g, opp, instant); !errors.As(err, &cant) || cant.Reason != "Your opponents can't cast noncreature spells." {
		t.Errorf("an opponent's instant under the emblem: %v, want the emblem's clause", err)
	}
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID}
	if err := a109p6CastGate(g, opp, bear); err != nil {
		t.Errorf("an opponent's creature spell was refused: %v", err)
	}
	mine := game.Card{InstanceID: uuid.New(), Name: "Opt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	if err := a109p6CastGate(g, me, mine); err != nil {
		t.Errorf("the emblem's owner was refused their own instant: %v", err)
	}
}

// The +1 offers the top card when it is a noncreature, nonland card, and
// puts it into the hand on a yes.
func TestNarsetPlusOneTakesANoncreatureNonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	narset := a109p6Walker(t, g, me, "Narset Transcendent", narsetOracle, 6)
	top := b02TopOfLibrary(me, "Ponder", "Sorcery")
	b16Activate(t, g, me.ID, narset, 0, game.ActivateAbilityParams{})
	answerChooseCards(t, g, me.ID, top)
	if !me.Hand.Contains(top) {
		t.Error("the sorcery on top was not put into the hand")
	}
	if got := loyaltyCount(g, narset); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

// A creature or a land on top is looked at and left there: nothing is
// offered.
func TestNarsetPlusOneLeavesACreatureOrLandOnTop(t *testing.T) {
	for _, typeLine := range []string{"Creature — Bear", "Basic Land — Forest", "Artifact Creature — Golem"} {
		g := newCatalogGame(t)
		me, _ := a109p6Seats(g)
		narset := a109p6Walker(t, g, me, "Narset Transcendent", narsetOracle, 6)
		top := b02TopOfLibrary(me, "Top Card", typeLine)
		b16Activate(t, g, me.ID, narset, 0, game.ActivateAbilityParams{})
		if c := chooseCardsChoiceFor(g, me.ID); c != nil {
			t.Errorf("%s: the card on top was offered", typeLine)
		}
		if me.Hand.Contains(top) || !me.Library.Contains(top) {
			t.Errorf("%s: the top card moved", typeLine)
		}
	}
}

// The −2: the next instant or sorcery cast from hand this turn gains
// rebound, a creature spell cast first does not use it up, and the spell
// after the one it fired on gains nothing.
func TestNarsetMinusTwoGivesTheNextSpellFromHandRebound(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	narset := a109p6Walker(t, g, me, "Narset Transcendent", narsetOracle, 6)
	b16Activate(t, g, me.ID, narset, 1, game.ActivateAbilityParams{})

	castPlainSpell(t, g, me, "Creature — Bear")
	passPriorityAroundTable(t, g)
	first := castPlainSpell(t, g, me, "Instant")
	if n := triggersOnStackFrom(g, narset); n != 1 {
		t.Fatalf("Narset's delayed triggers on the stack = %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || me.Graveyard.Contains(first) {
		t.Fatal("the instant Narset gave rebound was not exiled as it resolved")
	}
	if n := reboundTriggersFor(g, first); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}
	second := castPlainSpell(t, g, me, "Sorcery")
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(second) {
		t.Error("the second spell of the turn rebounded too (the trigger fires once, CR 603.7b)")
	}
}

// "From your hand": the condition refuses a cast from anywhere else, and
// a spell of another player.
func TestNarsetMinusTwoWantsACastFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	instant := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: instant, Name: "Some Instant", TypeLine: "Instant",
			ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	})
	dt := &game.DelayedTrigger{Controller: me.ID}
	params := game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}}
	fires := func(actor uuid.UUID, from game.ZoneKind) bool {
		var ok bool
		g.WithWriteLock(func() {
			ok = youNextCastFromHand(game.Event{Kind: game.EventCast, Actor: actor, CardID: instant,
				OldZone: from, NewZone: game.ZoneStack}, dt, g, params)
		})
		return ok
	}
	if !fires(me.ID, game.ZoneHand) {
		t.Fatal("setup: a cast from hand does not fire it")
	}
	for _, from := range []game.ZoneKind{game.ZoneExile, game.ZoneGraveyard, game.ZoneCommand, game.ZoneLibrary} {
		if fires(me.ID, from) {
			t.Errorf("a cast from %s fired it", from)
		}
	}
	if fires(opp.ID, game.ZoneHand) {
		t.Error("an opponent's cast fired it")
	}
}

// --- §5: Dovin Baan ---------------------------------------------------

// The +1: -3/-0 and no activated abilities, mana abilities included,
// until Dovin's controller's next turn.
func TestDovinBaanPlusOneShrinksAndSilences(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	dovin := a109p6Walker(t, g, me, "Dovin Baan", dovinBaanOracle, 3)
	bear := pushCatalogPermanent(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", "", false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
		}
	}
	b16Activate(t, g, me.ID, dovin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature left the battlefield")
	}
	if eff := c.Effective(); eff.Power != 1 || eff.Toughness != 4 {
		t.Errorf("P/T = %d/%d, want 1/4", eff.Power, eff.Toughness)
	}
	if !game.Restricted(&c, game.CantActivate) || !game.Restricted(&c, game.CantActivateMana) {
		t.Error("the creature's activated abilities, mana abilities included, can still be activated")
	}
	if got := loyaltyCount(g, dovin); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
}

// The −1: gain 2 life and draw a card.
func TestDovinBaanMinusOneGainsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	dovin := a109p6Walker(t, g, me, "Dovin Baan", dovinBaanOracle, 3)
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, dovin, 1, game.ActivateAbilityParams{})
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +2 and +1", life, me.Life, hand, me.Hand.Size())
	}
}

// The −7's emblem caps an opponent's untap step at two permanents, and
// its owner's own untap step not at all.
func TestDovinBaanEmblemCapsOpponentsUntapSteps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	dovin := a109p6Walker(t, g, me, "Dovin Baan", dovinBaanOracle, 7)
	b16Activate(t, g, me.ID, dovin, 2, game.ActivateAbilityParams{})
	if me.Emblems == nil || len(me.Emblems.Cards) != 1 {
		t.Fatal("the −7 made no emblem")
	}
	var theirs, mine []uuid.UUID
	for i := 0; i < 3; i++ {
		theirs = append(theirs, pushTappedForTest(g, opp.ID, "Forest", "", "Basic Land — Forest"))
		mine = append(mine, pushTappedForTest(g, me.ID, "Island", "", "Basic Land — Island"))
	}
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	choice := advanceToUntapChoiceOf(t, g, oppSeat)
	if choice.ChooseMax != 2 {
		t.Fatalf("the opponent may untap %d, want 2", choice.ChooseMax)
	}
	if err := g.ResolveUntapChoice(choice.ID, opp.ID, theirs[:2]); err != nil {
		t.Fatalf("answer the cap: %v", err)
	}
	if !tappedForTest(t, g, theirs[2]) {
		t.Error("the third permanent untapped under the emblem")
	}
	mySeat := (oppSeat + len(g.Seats) - 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, mySeat)
	for _, id := range mine {
		if tappedForTest(t, g, id) {
			t.Error("the emblem capped its owner's own untap step")
		}
	}
}

// --- §6: the graveyard restriction ------------------------------------

// Ground Seal draws a card as it enters, and then no spell may target a
// card in a graveyard: Regrowth is refused at announce (CR 601.2c).
func TestGroundSealDrawsAndStopsGraveyardTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Ground Seal", "Enchantment", groundSealOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want the enters draw", hand, me.Hand.Size())
	}
	dead := pushGraveyardCardForTest(me, "Old Body")
	err := castCatalogSpellErr(t, g, "Regrowth", "Sorcery", regrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Regrowth at a graveyard card under Ground Seal: %v, want ErrIllegalTarget", err)
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("the graveyard card moved")
	}
}

// A spell already aimed at a graveyard card when the restriction arrives
// has an illegal target at resolution and does nothing (CR 608.2b).
func TestGraveyardRestrictionMakesARegrowthOnTheStackFizzle(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Ground Seal", "Enchantment", groundSealOracle},
		{"Silent Gravestone", "Artifact", silentGravestoneOracle},
		{"Underworld Cerberus", "Creature — Dog", underworldCerberusOracle},
		{"Dennick, Pious Apprentice", "Legendary Creature — Human Soldier", dennickOracleID},
	} {
		g := newCatalogGame(t)
		me, _ := a109p6Seats(g)
		dead := pushGraveyardCardForTest(me, "Old Body")
		castCatalogSpell(t, g, "Regrowth", "Sorcery", regrowthOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
		pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		passPriorityAroundTable(t, g)
		if me.Hand.Contains(dead) || !me.Graveyard.Contains(dead) {
			t.Errorf("%s: Regrowth returned a card it can no longer target", tc.name)
		}
		var legal []uuid.UUID
		legal = a109p6LegalCards(g, me.ID, TargetCardInGraveyard("target card in a graveyard"))
		if len(legal) != 0 {
			t.Errorf("%s: graveyard cards are still legal targets: %v", tc.name, legal)
		}
	}
}

// Silent Gravestone's activation exiles itself and every card in every
// graveyard, then draws.
func TestSilentGravestoneExilesItselfAndAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	stone := pushCatalogPermanent(g, me.ID, "Silent Gravestone", "Artifact", silentGravestoneOracle, false)
	mine := pushGraveyardCardForTest(me, "My Dead")
	theirs := pushGraveyardCardForTest(opp, "Their Dead")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{stone, mine, theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile", id)
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a draw", hand, me.Hand.Size())
	}
}

// Underworld Cerberus dying exiles it, then every creature card in every
// graveyard goes to its owner's hand; a noncreature card stays.
func TestUnderworldCerberusDiesAndReturnsCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	dog := pushCatalogPermanent(g, me.ID, "Underworld Cerberus", "Creature — Dog", underworldCerberusOracle, false)
	mine := pushGraveyardCardForTest(me, "My Dead")
	theirs := pushGraveyardCardForTest(opp, "Their Dead")
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dog) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dog) {
		t.Error("the Cerberus was not exiled")
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a creature card did not return to its owner's hand")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}
}

// --- §6: Dennick's back face ------------------------------------------

// Dennick, Pious Apparition investigates once for a batch of creature
// cards put into graveyards, and only once each turn; a noncreature card
// and a token do not trigger it.
func TestDennickApparitionInvestigatesOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	row := disturbRow(dennickOracleID, []string{"W", "U"},
		disturbFace{"Dennick, Pious Apprentice", "Legendary Creature — Human Soldier", "{W}{U}", "2", "3"},
		disturbFace{"Dennick, Pious Apparition", "Legendary Creature — Spirit Soldier", "", "3", "2"})
	dennick := backFaceOnBattlefield(g, row, me)
	clues := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && strings.Contains(c.Name, "Clue") {
				n++
			}
		}
		return n
	}

	b02TopOfLibrary(opp, "Old Bolt", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := clues(); n != 0 {
		t.Fatalf("a noncreature card made %d Clues", n)
	}

	b02TopOfLibrary(opp, "Bear A", "Creature — Bear")
	b02TopOfLibrary(opp, "Bear B", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 2) })
	if n := triggersOnStackFrom(g, dennick); n > 1 {
		t.Errorf("two creature cards milled at once made %d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if n := clues(); n != 1 {
		t.Fatalf("Clues after the first creature cards = %d, want 1", n)
	}

	b02TopOfLibrary(me, "Bear C", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := clues(); n != 1 {
		t.Errorf("Clues after a second batch this turn = %d, want still 1", n)
	}
}

// The trigger condition: a creature card arriving in a graveyard from
// anywhere, never a token, never a move between graveyards.
func TestCreatureCardPutIntoAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	bear := pushGraveyardCardForTest(me, "Bear")
	token := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: token, Name: "Bear Token", TypeLine: "Token Creature — Bear", Owner: me.ID, Controller: me.ID})
	ask := func(ev game.Event) bool {
		var ok bool
		g.WithWriteLock(func() { ok = creatureCardPutIntoAGraveyard(ev, g) })
		return ok
	}
	for _, kind := range []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill} {
		if !ask(game.Event{Kind: kind, CardID: bear, OldZone: game.ZoneLibrary, NewZone: game.ZoneGraveyard}) {
			t.Errorf("%s of a creature card into a graveyard did not count", kind)
		}
	}
	if !ask(game.Event{Kind: game.EventCounterSpell, CardID: bear, Target: bear}) {
		t.Error("a countered creature spell now in a graveyard did not count")
	}
	if ask(game.Event{Kind: game.EventZoneMove, CardID: token, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard}) {
		t.Error("a token counted")
	}
	if ask(game.Event{Kind: game.EventZoneMove, CardID: bear, OldZone: game.ZoneGraveyard, NewZone: game.ZoneGraveyard}) {
		t.Error("a move between graveyards counted")
	}
	if ask(game.Event{Kind: game.EventZoneMove, CardID: bear, OldZone: game.ZoneGraveyard, NewZone: game.ZoneExile}) {
		t.Error("a card leaving a graveyard counted")
	}
}

// --- §6 and §4: Tomik -------------------------------------------------

// Tomik: an opponent's spells and abilities can't target his
// controller's lands on the battlefield or land cards in graveyards
// (anyone's), while his controller's can; a nonland card in a graveyard
// is untouched; and an opponent can't play a land card from a graveyard.
func TestTomikProtectsLandsAndGraveyardLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Tomik, Distinguished Advokist", "Legendary Creature — Human Advisor", tomikOracle, false)
	myLand := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirLand := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	deadLand := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: deadLand, Name: "Dead Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	deadBear := pushGraveyardCardForTest(opp, "Dead Bear")

	lands := TargetPermanent("target land", Land())
	cards := TargetCardInGraveyard("target card in a graveyard")
	for _, tc := range []struct {
		who      *game.Player
		land     uuid.UUID
		wantLand bool
		wantDead bool
	}{
		{opp, myLand, false, false},
		{opp, theirLand, false, false},
		{me, myLand, true, true},
	} {
		got := a109p6LegalCards(g, tc.who.ID, lands)
		if a109p6Has(got, tc.land) != tc.wantLand {
			t.Errorf("%s targeting land %s: legal = %v, want %v", tc.who.Name, tc.land, a109p6Has(got, tc.land), tc.wantLand)
		}
		gy := a109p6LegalCards(g, tc.who.ID, cards)
		if a109p6Has(gy, deadLand) != tc.wantDead {
			t.Errorf("%s targeting a land card in a graveyard: legal = %v, want %v", tc.who.Name, a109p6Has(gy, deadLand), tc.wantDead)
		}
		if !a109p6Has(gy, deadBear) {
			t.Errorf("%s may not target a creature card in a graveyard", tc.who.Name)
		}
	}

	var fromGraveyard, fromHand, mine error
	g.WithWriteLock(func() {
		land := game.Card{Name: "Dead Island", TypeLine: "Basic Land — Island"}
		fromGraveyard = g.LandPlayGateLocked(opp.ID, land, game.ZoneGraveyard)
		fromHand = g.LandPlayGateLocked(opp.ID, land, game.ZoneHand)
		mine = g.LandPlayGateLocked(me.ID, land, game.ZoneGraveyard)
	})
	var cant *game.CantPlayLandError
	if !errors.As(fromGraveyard, &cant) || !strings.Contains(cant.Reason, "Your opponents can't play land cards from graveyards.") {
		t.Errorf("an opponent's land play from a graveyard: %v, want Tomik's clause", fromGraveyard)
	}
	if fromHand != nil || mine != nil {
		t.Errorf("Tomik refused a land play from hand (%v) or his controller's (%v)", fromHand, mine)
	}
}
