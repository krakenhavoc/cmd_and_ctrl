package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std53_cards_test.go — Wilds of Eldraine, Standard batch 53 (#4452).
// Each test plays the real catalog entry and asserts what its text says.

const (
	std53BackForSecondsOracle = "ee3d21b9-58b4-480c-bae3-26de50f1a5c6"
	std53DreamSpoilersOracle  = "07c3e91c-3593-4797-b7be-ef4fb2913cef"
	std53FaerieFencingOracle  = "450a9b73-9024-4103-b9f9-216a2f395cff"
	std53FellHorsemanOracle   = fellHorsemanOracleID
	std53GumdropOracle        = gumdropPoisonerOracleID
	std53RowansOracle         = "f60234e4-9b53-47a5-a21a-82c80b5af072"
	std53IcewroughtOracle     = "0828d891-f7c3-4e51-b42e-6ce88d48ac8b"
	std53PicklockOracle       = picklockPranksterOracleID
	std53TalionsOracle        = "3a721a4c-696b-44d1-858d-20040bd01909"
	std53AshiokOracle         = "67d66678-a8e6-4f80-bade-1f1daf4b0610"
)

func TestStd53EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		std53BackForSecondsOracle:      "Back for Seconds",
		std53DreamSpoilersOracle:       "Dream Spoilers",
		std53FaerieFencingOracle:       "Faerie Fencing",
		std53FellHorsemanOracle:        "Fell Horseman",
		std53FellHorsemanOracle + "#1": "Deathly Ride",
		std53GumdropOracle:             "Gumdrop Poisoner",
		std53GumdropOracle + "#1":      "Tempt with Treats",
		std53RowansOracle:              "Rowan's Grim Search",
		std53IcewroughtOracle:          "Icewrought Sentry",
		std53PicklockOracle:            "Picklock Prankster",
		std53PicklockOracle + "#1":     "Free the Fae",
		std53TalionsOracle:             "Talion's Messenger",
		std53AshiokOracle:              "Ashiok, Wicked Manipulator",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// std53Adventure puts an Adventure card into p's hand.
func std53Adventure(p *game.Player, oracle string, creature, adventure game.Face) game.Card {
	return woeB1Adventure(p, oracle, creature, adventure)
}

// std53Bargain is a token to sacrifice for bargain.
func std53Bargain(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: owner, Controller: owner,
	})
}

// --- Faerie Fencing --------------------------------------------------------------

func TestStd53FaerieFencingShrinksByXWithoutAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 6, 6)
	p1g2Cast(t, g, "Faerie Fencing", "Instant", std53FaerieFencingOracle, "{X}{B}", game.CastSpellParams{
		Targets: cardRefs(foe), XValue: 2,
	})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 4 || tough != 4 {
		t.Errorf("the 6/6 is %d/%d, want 4/4", p, tough)
	}
}

func TestStd53FaerieFencingShrinksThreeMoreWhenYouControlledAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	faerie := b12Creature(g, me.ID, "My Faerie", "Creature — Faerie", 1, 1)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 6, 6)
	p1g2Cast(t, g, "Faerie Fencing", "Instant", std53FaerieFencingOracle, "{X}{B}", game.CastSpellParams{
		Targets: cardRefs(foe), XValue: 1,
	})
	// The Faerie leaving in response changes nothing: the check was made as
	// the spell was cast.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(faerie) })
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("the 6/6 is %d/%d, want 2/2 (-1/-1 and the additional -3/-3)", p, tough)
	}
}

// --- Fell Horseman // Deathly Ride -----------------------------------------------

func TestStd53FellHorsemanGoesToTheBottomOfTheLibraryWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	horseman := pushCatalogPermanent(g, me.ID, "Fell Horseman", "Creature — Zombie Knight", std53FellHorsemanOracle, false)
	woeB1Destroy(t, g, horseman)
	if me.Graveyard.Contains(horseman) {
		t.Error("Fell Horseman stayed in the graveyard")
	}
	if len(me.Library.Cards) == 0 || me.Library.Cards[0].InstanceID != horseman {
		t.Error("Fell Horseman is not on the bottom of its owner's library")
	}
}

func TestStd53DeathlyRideReturnsACreatureCardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	card := std53Adventure(me, std53FellHorsemanOracle,
		game.Face{Name: "Fell Horseman", TypeLine: "Creature — Zombie Knight", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
		game.Face{Name: "Deathly Ride", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{B}", Colors: []string{"B"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(dead))
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to hand")
	}
}

// --- Gumdrop Poisoner // Tempt with Treats ---------------------------------------

func TestStd53GumdropPoisonerShrinksByTheLifeYouGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	castCatalogSpell(t, g, "Gumdrop Poisoner", "Creature — Human Warlock", std53GumdropOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("the 5/5 is %d/%d, want 2/2 (gained 3 life)", p, tough)
	}
}

func TestStd53TemptWithTreatsMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := std53Adventure(me, std53GumdropOracle,
		game.Face{Name: "Gumdrop Poisoner", TypeLine: "Creature — Human Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 2},
		game.Face{Name: "Tempt with Treats", TypeLine: "Instant — Adventure", ManaCost: "{B}", Colors: []string{"B"}})
	woeB1CastAdventure(t, g, me, card, nil)
	food := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Food") {
			food++
		}
	}
	if food != 1 {
		t.Errorf("%d Food tokens, want 1", food)
	}
}

// --- Dream Spoilers --------------------------------------------------------------

func std53CastEvent(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	spell := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: spell, Name: "Some Instant", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: p.ID, CardID: spell})
	})
}

func TestStd53DreamSpoilersShrinksOnASpellCastDuringAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Dream Spoilers", "Creature — Faerie Warlock", std53DreamSpoilersOracle, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 3, 3)

	// On your own turn nothing triggers.
	std53CastEvent(t, g, me)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("Dream Spoilers triggered on a spell cast during your own turn")
	}

	// During an opponent's turn it does.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = (g.Turn.ActiveSeat + 1) % len(g.Seats) })
	std53CastEvent(t, g, me)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("the 3/3 is %d/%d, want 2/2", p, tough)
	}
}

// --- Picklock Prankster // Free the Fae ------------------------------------------

func TestStd53FreeTheFaeMillsFourThenMustTakeAnInstantSorceryOrFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	mk := func(name, typeLine string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID}
	}
	bear, bolt, faerie, forest := mk("Bear", "Creature — Bear"), mk("Bolt", "Instant"), mk("Sprite", "Creature — Faerie Rogue"), mk("Forest", "Basic Land — Forest")
	older := mk("Old Sorcery", "Sorcery")
	me.Library.PushTop(older)
	// Top first: the Bear, then the Bolt, the Faerie and the Forest.
	for _, c := range []game.Card{forest, faerie, bolt, bear} {
		me.Library.PushTop(c)
	}
	me.Graveyard.PushTop(mk("Already There", "Instant"))
	card := std53Adventure(me, std53PicklockOracle,
		game.Face{Name: "Picklock Prankster", TypeLine: "Creature — Faerie Rogue", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 3},
		game.Face{Name: "Free the Fae", TypeLine: "Instant — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Free the Fae: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick over the milled cards")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("pick is %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, bolt.InstanceID) || !hasID(pick.ChooseCards, faerie.InstanceID) {
		t.Errorf("offered %v, want the instant and the Faerie only (not the Bear, the Forest or what was already there)", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, faerie.InstanceID)
	if !me.Hand.Contains(faerie.InstanceID) {
		t.Error("the Faerie is not in hand")
	}
	if !me.Graveyard.Contains(bolt.InstanceID) || !me.Graveyard.Contains(bear.InstanceID) || !me.Graveyard.Contains(forest.InstanceID) {
		t.Error("the rest of the milled cards are not in the graveyard")
	}
	if me.Library.Size() == 0 || me.Library.Cards[me.Library.Size()-1].InstanceID != older.InstanceID {
		t.Error("the fifth card was milled")
	}
}

// --- Back for Seconds ------------------------------------------------------------

func std53GraveCreature(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}

func TestStd53BackForSecondsReturnsUpToTwoCreatureCardsUnbargained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	a, b := std53GraveCreature(me, "Small", "{1}{G}"), std53GraveCreature(me, "Big", "{5}{G}")
	castCatalogSpell(t, g, "Back for Seconds", "Sorcery", std53BackForSecondsOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("both creature cards were not returned to hand")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an unbargained spell offered the battlefield")
	}
}

func TestStd53BackForSecondsBargainedPutsOneSmallCardOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := std53Bargain(g, me.ID)
	small, big := std53GraveCreature(me, "Small", "{3}{G}"), std53GraveCreature(me, "Big", "{4}{G}")
	p1g2Cast(t, g, "Back for Seconds", "Sorcery", std53BackForSecondsOracle, "", game.CastSpellParams{
		Targets: cardRefs(small, big), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no offer to put a card onto the battlefield")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != small || pick.ChooseMin != 0 {
		t.Errorf("offered %v (min %d), want only the mana value 4 card, optional", pick.ChooseCards, pick.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, small)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature did not enter the battlefield")
	}
	if !me.Hand.Contains(big) {
		t.Error("the other creature card did not go to hand")
	}
}

func TestStd53BackForSecondsBargainedMayStillPutEverythingInHand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := std53Bargain(g, me.ID)
	small := std53GraveCreature(me, "Small", "{1}{G}")
	p1g2Cast(t, g, "Back for Seconds", "Sorcery", std53BackForSecondsOracle, "", game.CastSpellParams{
		Targets: cardRefs(small), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	if !me.Hand.Contains(small) || g.Battlefield.Contains(small) {
		t.Error("declining the battlefield did not leave the card in hand")
	}
}

// --- Rowan's Grim Search ---------------------------------------------------------

func TestStd53RowansGrimSearchUnbargainedDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	hand, life := handSize(me), me.Life
	castCatalogSpell(t, g, "Rowan's Grim Search", "Instant", std53RowansOracle, nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+2 {
		t.Errorf("hand %d, want %d (the spell left it, two cards drawn)", got, hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an unbargained spell looked at the library")
	}
}

func TestStd53RowansGrimSearchBargainedKeepsTwoOnTopThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := std53Bargain(g, me.ID)
	ids := seedLibrary(me, "A", "B", "C", "D", "E", "F")
	life := me.Life
	p1g2Cast(t, g, "Rowan's Grim Search", "Instant", std53RowansOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Fatalf("pick %+v, want up to two of the top four", pick)
	}
	answerChooseCards(t, g, me.ID, ids[1], ids[3])
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no prompt to order the cards kept on top")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{ids[3], ids[1]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The kept cards were drawn, in the chosen order: D then B.
	if !me.Hand.Contains(ids[3]) || !me.Hand.Contains(ids[1]) {
		t.Error("the two kept cards were not the two drawn")
	}
	if !me.Graveyard.Contains(ids[0]) || !me.Graveyard.Contains(ids[2]) {
		t.Error("the rest of the top four did not go to the graveyard")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if libraryTopNames(me, 1)[0] != "E" {
		t.Errorf("top of library %v, want E", libraryTopNames(me, 1))
	}
}

// --- Talion's Messenger ----------------------------------------------------------

func TestStd53TalionsMessengerLootsAndGrowsAFaerieWhenYouAttackWithAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	messenger := blbB1Push(g, me.ID, "Talion's Messenger", "Creature — Faerie Noble", std53TalionsOracle, 3, 2)
	other := blbB1Push(g, me.ID, "Other Faerie", "Creature — Faerie", "", 1, 1)
	hand := handSize(me)
	declareAttack(t, g, opp.ID, messenger)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Fatalf("hand %d after the draw, want %d", got, hand+1)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the discard")
	}
	pending := discardChoiceFor(g, me.ID)
	if pending == nil {
		t.Fatal("no discard prompt")
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, other); got != 1 {
		t.Errorf("the other Faerie has %d +1/+1 counters, want 1", got)
	}
}

func TestStd53TalionsMessengerIgnoresAnAttackWithoutAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	blbB1Push(g, me.ID, "Talion's Messenger", "Creature — Faerie Noble", std53TalionsOracle, 3, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := handSize(me)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand {
		t.Errorf("hand %d, want %d: no Faerie attacked", got, hand)
	}
}

// --- Icewrought Sentry -----------------------------------------------------------

func TestStd53IcewroughtSentryPumpsOnlyWhenYourEffectTapsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	sentry := pushCatalogPermanent(g, me.ID, "Icewrought Sentry", "Creature — Elemental Soldier", std53IcewroughtOracle, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other Foe", "Creature — Bear", 2, 2)
	tap := func(id, by uuid.UUID) {
		g.WithWriteLock(func() {
			if by == uuid.Nil {
				_ = g.TapTargetForEffect(id)
			} else {
				_ = g.TapTargetForEffectBy(id, by)
			}
		})
		passPriorityAroundTable(t, g)
	}
	base, _ := effectivePTOf(t, g, sentry)

	tap(mine, me.ID)   // your own creature
	tap(other, opp.ID) // an opponent's instruction
	if p, _ := effectivePTOf(t, g, sentry); p != base {
		t.Fatalf("the Sentry grew to %d from a tap that was not yours on an opponent's creature", p)
	}
	untapForTest(g, other)
	tap(other, uuid.Nil) // a tap with no effect behind it
	if p, _ := effectivePTOf(t, g, sentry); p != base {
		t.Fatalf("the Sentry grew to %d from a tap no effect made", p)
	}
	tap(foe, me.ID)
	if p, tough := effectivePTOf(t, g, sentry); p != base+2 || tough != 2 {
		t.Errorf("the Sentry is %d/%d, want %d/2", p, tough, base+2)
	}
	tap(foe, me.ID) // already tapped: nothing becomes tapped
	if p, _ := effectivePTOf(t, g, sentry); p != base+2 {
		t.Errorf("the Sentry is %d wide after tapping a tapped creature, want %d", p, base+2)
	}
}

func TestStd53IcewroughtSentryPaysToTapOnAttackAndGrowsFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	sentry := pushCatalogPermanent(g, me.ID, "Icewrought Sentry", "Creature — Elemental Soldier", std53IcewroughtOracle, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, sentry)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment")
	}
	b06AddMana(me, "C", "U")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if !std53Tapped(g, foe) {
		t.Error("the opponent's creature was not tapped")
	}
	if p, tough := effectivePTOf(t, g, sentry); p != 3 || tough != 2 {
		t.Errorf("the Sentry is %d/%d, want 3/2 (the fixture's 1/1 plus +2/+1)", p, tough)
	}
}

func std53Tapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}

// --- Ashiok, Wicked Manipulator --------------------------------------------------

func std53Ashiok(t *testing.T, g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	t.Helper()
	return pushCatalogWalkerWithLine(g, owner, "Ashiok, Wicked Manipulator", "Legendary Planeswalker — Ashiok", std53AshiokOracle, loyalty)
}

func TestStd53AshiokExilesFromTheLibraryInsteadOfPayingLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	std53Ashiok(t, g, me.ID, 5)
	ids := seedLibrary(me, "A", "B", "C", "D")
	life, size := me.Life, me.Library.Size()
	g.WithWriteLock(func() {
		if err := g.PayLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatalf("pay life: %v", err)
		}
	})
	if me.Life != life {
		t.Errorf("life %d, want %d: the payment was exiled instead", me.Life, life)
	}
	for _, id := range ids[:3] {
		if !g.Exile.Contains(id) {
			t.Error("a card from the top three was not exiled")
		}
	}
	if g.Exile.Contains(ids[3]) || me.Library.Size() != size-3 {
		t.Error("exactly three cards should have left the library")
	}
	// An opponent's payment is theirs: Ashiok's controller's replacement
	// does not touch it.
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.PayLifeForEffect(uuid.Nil, opp.ID, 2) })
	if opp.Life != oppLife-2 {
		t.Errorf("opponent's life %d, want %d", opp.Life, oppLife-2)
	}
}

func TestStd53AshiokPaysLifeAsNormalWithTooFewCardsInTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	std53Ashiok(t, g, me.ID, 5)
	g.WithWriteLock(func() {
		for me.Library.Size() > 2 {
			me.Library.Cards = me.Library.Cards[1:]
		}
	})
	life := me.Life
	g.WithWriteLock(func() { _ = g.PayLifeForEffect(uuid.Nil, me.ID, 3) })
	if me.Life != life-3 || me.Library.Size() != 2 {
		t.Errorf("life %d library %d, want %d and 2: the library is too short to pay with", me.Life, me.Library.Size(), life-3)
	}
}

func TestStd53AshiokDoesNotReplaceLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	std53Ashiok(t, g, me.ID, 5)
	life, size := me.Life, me.Library.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	if me.Life != life-3 || me.Library.Size() != size {
		t.Errorf("life %d library %d: losing life is not paying it", me.Life, me.Library.Size())
	}
}

func TestStd53AshiokPlusOneExilesOneAndTakesTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	toMain(t, g)
	ashiok := std53Ashiok(t, g, me.ID, 5)
	ids := seedLibrary(me, "A", "B", "C")
	b16Activate(t, g, me.ID, ashiok, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) || !g.Exile.Contains(ids[0]) {
		t.Error("B should be in hand and A exiled")
	}
	if n := loyaltyCount(g, ashiok); n != 6 {
		t.Errorf("loyalty %d, want 6", n)
	}
}

func std53Nightmares(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Nightmare" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestStd53AshiokMinusTwoMakesTwoNightmaresThatStayWithNothingExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	toMain(t, g)
	ashiok := std53Ashiok(t, g, me.ID, 5)
	b16Activate(t, g, me.ID, ashiok, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	nightmares := std53Nightmares(g, me.ID)
	if len(nightmares) != 2 {
		t.Fatalf("%d Nightmares, want 2", len(nightmares))
	}
	if p, tough := effectivePTOf(t, g, nightmares[0]); p != 1 || tough != 1 {
		t.Errorf("the Nightmare is %d/%d, want 1/1", p, tough)
	}
	// Nothing has been exiled this turn: the beginning of combat does nothing.
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if plusOnes(g, nightmares[0]) != 0 {
		t.Fatal("a Nightmare grew with nothing exiled this turn")
	}
}

func TestStd53AshiokNightmaresGrowWhenACardWasExiledThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	toMain(t, g)
	ashiok := std53Ashiok(t, g, me.ID, 5)
	b16Activate(t, g, me.ID, ashiok, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _, _ = g.MillToZoneForEffect(me.ID, 1, game.ZoneExile) })
	advanceTo(t, g, game.StepBeginCombat)
	// Two Nightmares trigger together: the controller orders them (CR 603.3b).
	answerOpenOrderingIfAny(t, g, me.ID)
	passPriorityAroundTable(t, g)
	grown := 0
	for _, id := range std53Nightmares(g, me.ID) {
		if plusOnes(g, id) == 1 {
			grown++
		}
	}
	if grown != 2 {
		t.Errorf("%d Nightmares grew, want both", grown)
	}
}

func TestStd53AshiokMinusSevenExilesTheTotalManaValueOfYourExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	toMain(t, g)
	ashiok := std53Ashiok(t, g, me.ID, 7)
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine 3", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine 2", TypeLine: "Instant", ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs 5", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Owner: opp.ID, Controller: opp.ID})
	size := opp.Library.Size()
	b16Activate(t, g, me.ID, ashiok, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if got := size - opp.Library.Size(); got != 5 {
		t.Errorf("the target player exiled %d cards, want 5 (3 + 2, not their own 5)", got)
	}
}
