package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b31_cards_test.go — the Bloomburrow leftovers of batch 31 (Festival of
// Embers, Keen-Eyed Curator, Glarb, Hugs, Ral, Alania, Kastral, Wishing Well,
// Fecund Greenshell), each against the engine seam it needed.

const (
	festivalOfEmbersOracle = "2de138b0-3159-4a0c-a474-a4e5600b2e51"
	keenEyedCuratorOracle  = "46ce3698-e4e6-44e9-8b99-2b12a703e91e"
	glarbOracle            = "df9409e6-bd6d-4b5b-b807-fcf347ee940a"
	hugsOracle             = "3c5c227a-335b-4995-9821-d4269a0a233a"
	ralCracklingWitOracle  = "3c970f43-17da-4187-a084-efd53b579067"
	alaniaOracle           = "63d50d5e-56de-44c4-b906-b2a21a56e3f3"
	kastralOracle          = "ced5fe9a-4237-4c89-9dd6-b133b966ebcc"
	wishingWellOracle      = "78dbbc15-9304-4178-b9b2-1db6c64ca11a"
	fecundGreenshellOracle = "a090be5a-b0f2-4c17-8f7d-e742f9de8f6b"
)

func TestStdB31FestivalOfEmbersCastsAnInstantFromTheGraveyardForOneLifeThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Festival of Embers", "Enchantment", festivalOfEmbersOracle, 0, 0)
	bolt := finB6GraveyardArtifact(me, "Shock", "Instant", "{1}")
	advanceToMain(t, g)
	floatMana(t, g, me, "{R}")
	life := me.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "festival_of_embers"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want 1 paid", life, me.Life)
	}
	if me.Graveyard.Contains(bolt) {
		t.Error("the spell should be exiled, not returned to the graveyard")
	}
}

func TestStdB31FestivalOfEmbersNeedsYourTurnAndAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Festival of Embers", "Enchantment", festivalOfEmbersOracle, 0, 0)
	bear := finB6GraveyardArtifact(me, "Bear", "Creature — Bear", "{1}")
	advanceToMain(t, g)
	floatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "festival_of_embers"}); err == nil {
		t.Error("a creature card was cast from the graveyard")
	}
	// During an opponent's turn the permission is not live.
	bolt := finB6GraveyardArtifact(me, "Shock", "Instant", "{1}")
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	floatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "festival_of_embers"}); err == nil {
		t.Error("the permission worked on an opponent's turn")
	}
}

func TestStdB31FestivalOfEmbersExilesYourDyingCreaturesAndTokensButNotTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Festival of Embers", "Enchantment", festivalOfEmbersOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	avB18Destroy(t, g, mine)
	avB18Destroy(t, g, theirs)
	if me.Graveyard.Contains(mine) {
		t.Error("your creature should be exiled instead of dying")
	}
	if !g.Exile.Contains(mine) {
		t.Error("your creature should be in exile")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature goes to their graveyard as usual")
	}
}

func TestStdB31KeenEyedCuratorGrowsWithFourCardTypesExiledWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	curator := b12Push(g, me.ID, "Keen-Eyed Curator", "Creature — Raccoon Scout", keenEyedCuratorOracle, 3, 3)
	var cards []uuid.UUID
	for _, tl := range []string{"Creature — Bear", "Land", "Instant", "Artifact"} {
		cards = append(cards, finB6GraveyardArtifact(opp, "Card", tl, "{1}"))
	}
	advanceToMain(t, g)
	for i, id := range cards {
		floatMana(t, g, me, "{G}")
		if err := g.ActivateCatalogAbility(me.ID, curator, 0, game.ActivateAbilityParams{Targets: twTarget(id)}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		wantPower := 3
		if i == 3 {
			wantPower = 7
		}
		if got := effectivePower(t, g, curator); got != wantPower {
			t.Errorf("after %d exiles power is %d, want %d", i+1, got, wantPower)
		}
	}
	if !hasEffectiveKeyword(t, g, curator, "trample") {
		t.Error("four card types: trample")
	}
}

func TestStdB31KeenEyedCuratorCountsOnlyItsOwnExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	curator := b12Push(g, me.ID, "Keen-Eyed Curator", "Creature — Raccoon Scout", keenEyedCuratorOracle, 3, 3)
	for _, tl := range []string{"Creature — Bear", "Land", "Instant", "Artifact"} {
		id := uuid.New()
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Elsewhere", TypeLine: tl, Owner: opp.ID, Controller: opp.ID})
	}
	if got := effectivePower(t, g, curator); got != 3 {
		t.Errorf("cards exiled by something else: power %d, want 3", got)
	}
}

func TestStdB31GlarbPlaysLandsAndCastsBigSpellsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Glarb, Calamity's Augur", "Legendary Creature — Frog Rogue", glarbOracle, 2, 4)
	big := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a mana value 4 spell from the top: %v", err)
	}
}

func TestStdB31GlarbRefusesASmallSpellFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Glarb, Calamity's Augur", "Legendary Creature — Frog Rogue", glarbOracle, 2, 4)
	small := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("a mana value 3 spell was cast from the top of the library")
	}
}

func TestStdB31GlarbPlaysALandFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Glarb, Calamity's Augur", "Legendary Creature — Frog Rogue", glarbOracle, 2, 4)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("play a land from the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land should be on the battlefield")
	}
}

func TestStdB31HugsExilesTheTopXCardsAndAddsALandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hugs, Grisly Guardian", TypeLine: "Legendary Creature — Badger Warrior", ManaCost: "{X}", OracleID: hugsOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	lib := me.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-3 {
		t.Errorf("library %d, want %d (X = 3 exiled)", me.Library.Size(), lib-3)
	}
	if g.Exile.Size() != 3 {
		t.Errorf("exile holds %d cards, want 3", g.Exile.Size())
	}
	first := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	second := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if !g.Battlefield.Contains(first) || !g.Battlefield.Contains(second) {
		t.Error("Hugs allows a second land play each turn")
	}
}

func TestStdB31RalPlusOneMakesAnOtterAndNoncreatureSpellsAddLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ral := b12Push(g, me.ID, "Ral, Crackling Wit", "Legendary Planeswalker — Ral", ralCracklingWitOracle, 0, 0)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, ral).Counters = map[string]int{game.CounterLoyalty: 4} })
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ral, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Otter"); n != 1 {
		t.Errorf("%d Otters, want 1", n)
	}
	if got := counterCount(g, ral, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d after the +1, want 5", got)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ral, game.CounterLoyalty); got != 6 {
		t.Errorf("loyalty %d after a noncreature spell, want 6", got)
	}
}

func TestStdB31RalMinusTenDrawsAndGivesAStormEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ral := b12Push(g, me.ID, "Ral, Crackling Wit", "Legendary Planeswalker — Ral", ralCracklingWitOracle, 0, 0)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, ral).Counters = map[string]int{game.CounterLoyalty: 10} })
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, ral, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d (drew three)", me.Hand.Size(), hand+3)
	}
	if me.Emblems == nil || me.Emblems.Size() != 1 {
		t.Fatalf("emblems: %+v", me.Emblems)
	}
}

func TestStdB31AlaniaCopiesTheFirstInstantWhenAnOpponentDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Alania, Divergent Storm", "Legendary Creature — Otter Wizard", alaniaOracle, 3, 5)
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the opponent's hand %d, want %d (they drew)", opp.Hand.Size(), hand+1)
	}
}

func TestStdB31AlaniaTriggersOnlyOnTheFirstInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Alania, Divergent Storm", "Legendary Creature — Otter Wizard", alaniaOracle, 3, 5)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, false)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a second instant this turn must not trigger Alania")
	}
}

func TestStdB31KastralDrawsWhenABirdConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kastral := b12Push(g, me.ID, "Kastral, the Windcrested", "Legendary Creature — Bird Scout", kastralOracle, 4, 5)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, kastral)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt after a Bird dealt combat damage to a player")
	}
	answerMode(t, g, c, me.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestStdB31KastralPutsABirdFromHandOntoTheBattlefieldWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kastral := b12Push(g, me.ID, "Kastral, the Windcrested", "Legendary Creature — Bird Scout", kastralOracle, 4, 5)
	bird := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bird, Name: "Sparrow", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	attackWith(t, g, opp.ID, kastral)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, bird)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bird) {
		t.Fatal("the Bird should be on the battlefield")
	}
	if got := countersOn(g, bird, game.CounterFinality); got != 1 {
		t.Errorf("finality counters %d, want 1", got)
	}
}

func TestStdB31WishingWellPutsACoinCounterAndOffersAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	well := b12Push(g, me.ID, "Wishing Well", "Artifact", wishingWellOracle, 0, 0)
	spell := finB6GraveyardArtifact(me, "Shock", "Instant", "{1}")
	finB6GraveyardArtifact(me, "Big Spell", "Sorcery", "{3}{R}")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, well, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, well, stdB31CoinCounter); got != 1 {
		t.Fatalf("coin counters %d, want 1", got)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, spell)
		passPriorityAroundTable(t, g)
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell)
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(spell) {
		t.Error("the spell was cast, and it is exiled instead of returning to the graveyard")
	}
}

func TestStdB31FecundGreenshellPumpsWithTenLandsAndLooksOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shell := b12Push(g, me.ID, "Fecund Greenshell", "Creature — Elemental Turtle", fecundGreenshellOracle, 4, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("fewer than ten lands: power %d, want 2", got)
	}
	for i := 0; i < 10; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("ten lands: power %d, want 4", got)
	}
	if got := effectivePower(t, g, shell); got != 6 {
		t.Errorf("Greenshell itself: power %d, want 6", got)
	}
}

func TestStdB31FecundGreenshellPutsANonlandIntoHandWhenAToughCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Fecund Greenshell", "Creature — Elemental Turtle", fecundGreenshellOracle, 4, 6)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Wall", "Creature — Wall", "", 0, 4)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
		t.Errorf("the top nonland card should be in hand: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestStdB31FecundGreenshellIgnoresAnEnteringCreatureWithPowerAtLeastItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Fecund Greenshell", "Creature — Elemental Turtle", fecundGreenshellOracle, 4, 6)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	castHoldCreature(t, g, "Brute", "Creature — Ogre", "", 4, 4)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) {
		t.Error("a 4/4 does not have toughness greater than its power")
	}
}
