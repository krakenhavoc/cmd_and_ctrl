package testsds

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// a109p6CastGate asks the cast gate whether p may cast `card` from hand.
func a109p6CastGate(g *game.Game, p *game.Player, card game.Card) error {
	var err error
	g.WithWriteLock(func() { err = g.CastGateLocked(p.ID, card, game.ZoneHand, game.CastSpellParams{}) })
	return err
}

// --- §5: Narset Transcendent ----------------------------------------

// The −9's emblem refuses an opponent's noncreature spell with its
// clause, and nothing else: an opponent's creature spell, and Narset's
// controller's own spells, are cast as normal. The emblem outlives
// Narset.
func TestNarsetEmblemStopsOpponentsNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	narset := effectstest.A109p6Walker(t, g, me, "Narset Transcendent", effectstest.NarsetOracle, 9)
	effectstest.B16Activate(t, g, me.ID, narset, 2, game.ActivateAbilityParams{})
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.A109p6Seats(g)
	narset := effectstest.A109p6Walker(t, g, me, "Narset Transcendent", effectstest.NarsetOracle, 6)
	top := effectstest.B02TopOfLibrary(me, "Ponder", "Sorcery")
	effectstest.B16Activate(t, g, me.ID, narset, 0, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	if !me.Hand.Contains(top) {
		t.Error("the sorcery on top was not put into the hand")
	}
	if got := effectstest.LoyaltyCount(g, narset); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

// A creature or a land on top is looked at and left there: nothing is
// offered.
func TestNarsetPlusOneLeavesACreatureOrLandOnTop(t *testing.T) {
	for _, typeLine := range []string{"Creature — Bear", "Basic Land — Forest", "Artifact Creature — Golem"} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.A109p6Seats(g)
		narset := effectstest.A109p6Walker(t, g, me, "Narset Transcendent", effectstest.NarsetOracle, 6)
		top := effectstest.B02TopOfLibrary(me, "Top Card", typeLine)
		effectstest.B16Activate(t, g, me.ID, narset, 0, game.ActivateAbilityParams{})
		if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.A109p6Seats(g)
	narset := effectstest.A109p6Walker(t, g, me, "Narset Transcendent", effectstest.NarsetOracle, 6)
	effectstest.B16Activate(t, g, me.ID, narset, 1, game.ActivateAbilityParams{})

	effectstest.CastPlainSpell(t, g, me, "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.CastPlainSpell(t, g, me, "Instant")
	if n := effectstest.TriggersOnStackFrom(g, narset); n != 1 {
		t.Fatalf("Narset's delayed triggers on the stack = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || me.Graveyard.Contains(first) {
		t.Fatal("the instant Narset gave rebound was not exiled as it resolved")
	}
	if n := effectstest.ReboundTriggersFor(g, first); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}
	second := effectstest.CastPlainSpell(t, g, me, "Sorcery")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(second) {
		t.Error("the second spell of the turn rebounded too (the trigger fires once, CR 603.7b)")
	}
}

// "From your hand": the condition refuses a cast from anywhere else, and
// a spell of another player.
func TestNarsetMinusTwoWantsACastFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
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
			ok = effects.YouNextCastFromHand(game.Event{Kind: game.EventCast, Actor: actor, CardID: instant,
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	dovin := effectstest.A109p6Walker(t, g, me, "Dovin Baan", effectstest.DovinBaanOracle, 3)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", "", false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
		}
	}
	effectstest.B16Activate(t, g, me.ID, dovin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature left the battlefield")
	}
	if eff := c.Effective(); eff.Power != 1 || eff.Toughness != 4 {
		t.Errorf("P/T = %d/%d, want 1/4", eff.Power, eff.Toughness)
	}
	if !game.Restricted(&c, game.CantActivate) || !game.Restricted(&c, game.CantActivateMana) {
		t.Error("the creature's activated abilities, mana abilities included, can still be activated")
	}
	if got := effectstest.LoyaltyCount(g, dovin); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
}

// The −1: gain 2 life and draw a card.
func TestDovinBaanMinusOneGainsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.A109p6Seats(g)
	dovin := effectstest.A109p6Walker(t, g, me, "Dovin Baan", effectstest.DovinBaanOracle, 3)
	life, hand := me.Life, me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, dovin, 1, game.ActivateAbilityParams{})
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +2 and +1", life, me.Life, hand, me.Hand.Size())
	}
}

// The −7's emblem caps an opponent's untap step at two permanents, and
// its owner's own untap step not at all.
func TestDovinBaanEmblemCapsOpponentsUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	dovin := effectstest.A109p6Walker(t, g, me, "Dovin Baan", effectstest.DovinBaanOracle, 7)
	effectstest.B16Activate(t, g, me.ID, dovin, 2, game.ActivateAbilityParams{})
	if me.Emblems == nil || len(me.Emblems.Cards) != 1 {
		t.Fatal("the −7 made no emblem")
	}
	var theirs, mine []uuid.UUID
	for i := 0; i < 3; i++ {
		theirs = append(theirs, effectstest.PushTappedForTest(g, opp.ID, "Forest", "", "Basic Land — Forest"))
		mine = append(mine, effectstest.PushTappedForTest(g, me.ID, "Island", "", "Basic Land — Island"))
	}
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	choice := effectstest.AdvanceToUntapChoiceOf(t, g, oppSeat)
	if choice.ChooseMax != 2 {
		t.Fatalf("the opponent may untap %d, want 2", choice.ChooseMax)
	}
	if err := g.ResolveUntapChoice(choice.ID, opp.ID, theirs[:2]); err != nil {
		t.Fatalf("answer the cap: %v", err)
	}
	if !effectstest.TappedForTest(t, g, theirs[2]) {
		t.Error("the third permanent untapped under the emblem")
	}
	mySeat := (oppSeat + len(g.Seats) - 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, mySeat)
	for _, id := range mine {
		if effectstest.TappedForTest(t, g, id) {
			t.Error("the emblem capped its owner's own untap step")
		}
	}
}

// --- §6: the graveyard restriction ------------------------------------

// Ground Seal draws a card as it enters, and then no spell may target a
// card in a graveyard: Regrowth is refused at announce (CR 601.2c).
func TestGroundSealDrawsAndStopsGraveyardTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.A109p6Seats(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Ground Seal", "Enchantment", effectstest.GroundSealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want the enters draw", hand, me.Hand.Size())
	}
	dead := effectstest.PushGraveyardCardForTest(me, "Old Body")
	err := effectstest.CastCatalogSpellErr(t, g, "Regrowth", "Sorcery", effectstest.RegrowthOracle,
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
		{"Ground Seal", "Enchantment", effectstest.GroundSealOracle},
		{"Silent Gravestone", "Artifact", effectstest.SilentGravestoneOracle},
		{"Underworld Cerberus", "Creature — Dog", effectstest.UnderworldCerberusOracle},
		{"Dennick, Pious Apprentice", "Legendary Creature — Human Soldier", effects.DennickOracleID},
	} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.A109p6Seats(g)
		dead := effectstest.PushGraveyardCardForTest(me, "Old Body")
		effectstest.CastCatalogSpell(t, g, "Regrowth", "Sorcery", effectstest.RegrowthOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
		effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(dead) || !me.Graveyard.Contains(dead) {
			t.Errorf("%s: Regrowth returned a card it can no longer target", tc.name)
		}
		legal := effectstest.A109p6LegalCards(g, me.ID, effects.TargetCardInGraveyard("target card in a graveyard"))
		if len(legal) != 0 {
			t.Errorf("%s: graveyard cards are still legal targets: %v", tc.name, legal)
		}
	}
}

// Silent Gravestone's activation exiles itself and every card in every
// graveyard, then draws.
func TestSilentGravestoneExilesItselfAndAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	effectstest.ToMain(t, g)
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Silent Gravestone", "Artifact", effectstest.SilentGravestoneOracle, false)
	mine := effectstest.PushGraveyardCardForTest(me, "My Dead")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{})
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	effectstest.ToMain(t, g)
	dog := effectstest.PushCatalogPermanent(g, me.ID, "Underworld Cerberus", "Creature — Dog", effectstest.UnderworldCerberusOracle, false)
	mine := effectstest.PushGraveyardCardForTest(me, "My Dead")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dog) })
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	effectstest.ToMain(t, g)
	row := effectstest.DisturbRow(effects.DennickOracleID, []string{"W", "U"},
		effectstest.DisturbFace{Name: "Dennick, Pious Apprentice", TypeLine: "Legendary Creature — Human Soldier", Cost: "{W}{U}", Power: "2", Toughness: "3"},
		effectstest.DisturbFace{Name: "Dennick, Pious Apparition", TypeLine: "Legendary Creature — Spirit Soldier", Cost: "", Power: "3", Toughness: "2"})
	dennick := effectstest.BackFaceOnBattlefield(g, row, me)
	clues := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && strings.Contains(c.Name, "Clue") {
				n++
			}
		}
		return n
	}

	effectstest.B02TopOfLibrary(opp, "Old Bolt", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := clues(); n != 0 {
		t.Fatalf("a noncreature card made %d Clues", n)
	}

	effectstest.B02TopOfLibrary(opp, "Bear A", "Creature — Bear")
	effectstest.B02TopOfLibrary(opp, "Bear B", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 2) })
	if n := effectstest.TriggersOnStackFrom(g, dennick); n > 1 {
		t.Errorf("two creature cards milled at once made %d triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := clues(); n != 1 {
		t.Fatalf("Clues after the first creature cards = %d, want 1", n)
	}

	effectstest.B02TopOfLibrary(me, "Bear C", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := clues(); n != 1 {
		t.Errorf("Clues after a second batch this turn = %d, want still 1", n)
	}
}

// The trigger condition: a creature card arriving in a graveyard from
// anywhere, never a token, never a move between graveyards.
func TestCreatureCardPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.A109p6Seats(g)
	bear := effectstest.PushGraveyardCardForTest(me, "Bear")
	token := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: token, Name: "Bear Token", TypeLine: "Token Creature — Bear", Owner: me.ID, Controller: me.ID})
	ask := func(ev game.Event) bool {
		var ok bool
		g.WithWriteLock(func() { ok = effects.CreatureCardPutIntoAGraveyard(ev, g) })
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	effectstest.ToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tomik, Distinguished Advokist", "Legendary Creature — Human Advisor", effectstest.TomikOracle, false)
	myLand := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirLand := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	deadLand := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: deadLand, Name: "Dead Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	deadBear := effectstest.PushGraveyardCardForTest(opp, "Dead Bear")

	lands := effects.TargetPermanent("target land", effects.Land())
	cards := effects.TargetCardInGraveyard("target card in a graveyard")
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
		got := effectstest.A109p6LegalCards(g, tc.who.ID, lands)
		if effectstest.A109p6Has(got, tc.land) != tc.wantLand {
			t.Errorf("%s targeting land %s: legal = %v, want %v", tc.who.Name, tc.land, effectstest.A109p6Has(got, tc.land), tc.wantLand)
		}
		gy := effectstest.A109p6LegalCards(g, tc.who.ID, cards)
		if effectstest.A109p6Has(gy, deadLand) != tc.wantDead {
			t.Errorf("%s targeting a land card in a graveyard: legal = %v, want %v", tc.who.Name, effectstest.A109p6Has(gy, deadLand), tc.wantDead)
		}
		if !effectstest.A109p6Has(gy, deadBear) {
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

// --- §5: Gideon of the Trials -----------------------------------------

// The +1: all damage the target would deal is prevented, until Gideon's
// controller's next turn; another source's damage is not.
func TestGideonOfTheTrialsPlusOnePreventsTheTargetsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	gideon := effectstest.A109p6Walker(t, g, me, "Gideon of the Trials", effectstest.GideonOfTheTrialsOracle, 3)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	wolf := effectstest.PushCatalogPermanent(g, opp.ID, "Wolf", "Creature — Wolf", "", false)
	effectstest.B16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, me.ID, 3) })
	if me.Life != life {
		t.Errorf("the targeted permanent's damage was dealt: life %d → %d", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(wolf, me.ID, 2) })
	if me.Life != life-2 {
		t.Errorf("another source's damage: life %d → %d, want -2", life, me.Life)
	}
	if got := effectstest.LoyaltyCount(g, gideon); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
	// Until your next turn (CR 611.2b): still there on the opponent's
	// turn, gone once Gideon's controller's next turn has begun.
	mySeat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, (mySeat+1)%len(g.Seats))
	before := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, me.ID, 3) })
	if me.Life != before {
		t.Errorf("the shield ended before Gideon's controller's next turn: life %d → %d", before, me.Life)
	}
	effectstest.AdvanceToUpkeepOf(t, g, mySeat)
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, me.ID, 3) })
	if me.Life != before-3 {
		t.Errorf("the shield outlived Gideon's controller's next turn: life %d → %d", before, me.Life)
	}
}

// The first 0: Gideon is a 4/4 Human Soldier creature with
// indestructible that is still a planeswalker, and damage to him is
// prevented, so he loses no loyalty.
func TestGideonOfTheTrialsAnimatesAndIsShielded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	gideon := effectstest.A109p6Walker(t, g, me, "Gideon of the Trials", effectstest.GideonOfTheTrialsOracle, 3)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	effectstest.B16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})
	c, ok := effectstest.BattlefieldCard(g, gideon)
	if !ok {
		t.Fatal("Gideon left the battlefield")
	}
	eff := c.Effective()
	if !c.IsCreature() || !c.IsPlaneswalker() || eff.Power != 4 || eff.Toughness != 4 ||
		!c.HasSubtype("Human") || !c.HasSubtype("Soldier") || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("Gideon after the 0 = %v %v %d/%d %v, want a 4/4 Human Soldier creature planeswalker with indestructible",
			eff.Types, eff.Subtypes, eff.Power, eff.Toughness, eff.Abilities)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, gideon, 3) })
	if got := effectstest.LoyaltyCount(g, gideon); got != 3 {
		t.Errorf("loyalty after 3 damage = %d, want 3 (prevented)", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, gideon); c.DamageMarked != 0 {
		t.Errorf("Gideon has %d damage marked, want 0", c.DamageMarked)
	}
}

// The emblem: while its owner controls a Gideon planeswalker they can't
// lose and their opponents can't win; without one, both again can.
func TestGideonOfTheTrialsEmblemHoldsWhileYouControlAGideon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.A109p6Seats(g)
	gideon := effectstest.A109p6Walker(t, g, me, "Gideon of the Trials", effectstest.GideonOfTheTrialsOracle, 3)
	effectstest.B16Activate(t, g, me.ID, gideon, 2, game.ActivateAbilityParams{})
	if me.Emblems == nil || len(me.Emblems.Cards) != 1 {
		t.Fatal("the second 0 made no emblem")
	}
	gated := func() (cantLose []game.LossCause, oppCantWin bool) {
		g.WithWriteLock(func() {
			cantLose = g.CantLoseCausesForEffect(me)
			oppCantWin = g.CantWinForEffect(opp)
		})
		return
	}
	if cantLose, oppCantWin := gated(); len(cantLose) == 0 || !oppCantWin {
		t.Errorf("with a Gideon: can't lose to %v, opponent can't win = %v; want both gated", cantLose, oppCantWin)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(gideon) })
	if cantLose, oppCantWin := gated(); len(cantLose) != 0 || oppCantWin {
		t.Errorf("with no Gideon: can't lose to %v, opponent can't win = %v; want neither", cantLose, oppCantWin)
	}
}
