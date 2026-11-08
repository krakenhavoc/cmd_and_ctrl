package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_t_cards_test.go — Outlaws of Thunder Junction (OTJ), the T cards
// (#4468). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjTTakeTheFallOracle   = "279f3aa5-5807-4797-8cd1-e009984a3691"
	otjTThisTownOracle      = "2550c2ae-834d-46e7-9ab4-d7bf7bfa6a78"
	otjTThrowOracle         = "d6780d29-0dd7-4f04-95b8-4ca6a55308f2"
	otjTLassoOracle         = "d98cc18b-6ebb-4f55-9d0d-b8fbe9a92bc1"
	otjTSalvoOracle         = "d3ce7c83-986f-410a-ae00-6c1c1f918a2b"
	otjTTinybonesJoinsUp    = "3d2bd4f6-174d-411a-b930-9161aeefd33a"
	otjTTombTrawlerOracle   = "d22d3934-8437-4a18-81c1-4e9c501846ca"
	otjTTrashTheTownOracle  = "b47f7f1e-2d81-4b1b-870d-4c1a4f020505"
	otjTDredgerOracle       = "9a691e96-fb47-4c23-a063-3b287c51147b"
	otjTTrickShotOracle     = "843c8485-21b8-45ef-9a0c-4967f07278bb"
	otjTTumbleweedOracle    = "ef574346-1d98-447d-acb6-4d8d53c5f651"
	otjTThisTownTypeLine    = "Instant"
	otjTThisTownPrintedCost = "{4}{U}"
)

func TestOtjTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjTTakeTheFallOracle:  "Take the Fall",
		otjTThisTownOracle:     "This Town Ain't Big Enough",
		otjTThrowOracle:        "Throw from the Saddle",
		otjTLassoOracle:        "Thunder Lasso",
		otjTSalvoOracle:        "Thunder Salvo",
		otjTTinybonesJoinsUp:   "Tinybones Joins Up",
		otjTTombTrawlerOracle:  "Tomb Trawler",
		otjTTrashTheTownOracle: "Trash the Town",
		otjTDredgerOracle:      "Treasure Dredger",
		otjTTrickShotOracle:    "Trick Shot",
		otjTTumbleweedOracle:   "Tumbleweed Rising",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if sa := plotOf(otjTTumbleweedOracle); sa == nil || sa.Cost != "{2}{G}" {
		t.Errorf("Tumbleweed Rising: plot %+v, want {2}{G}", sa)
	}
}

// --- Take the Fall -----------------------------------------------------------

func TestOtjTTakeTheFallShrinksByOneOrFourWithAnOutlawAndDraws(t *testing.T) {
	for _, tc := range []struct {
		name   string
		outlaw bool
		want   int
	}{
		{"no outlaw", false, 4},
		{"an outlaw", true, 1},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		seedLibrary(me, "One", "Two")
		if tc.outlaw {
			b12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 1, 1)
		}
		theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		advanceToMain(t, g)
		castCatalogSpell(t, g, "Take the Fall", "Instant", otjTTakeTheFallOracle, b16TargetCard(theirs))
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, theirs); p != tc.want || tough != 5 {
			t.Errorf("%s: the Giant is %d/%d, want %d/5", tc.name, p, tough, tc.want)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("%s: drew %d, want 1", tc.name, got)
		}
	}
}

// --- This Town Ain't Big Enough ----------------------------------------------

func TestOtjTThisTownCostsThreeLessWhenItTargetsYourPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	price := func(targets ...uuid.UUID) int {
		return selfPriced(t, g, me, otjTThisTownOracle, otjTThisTownTypeLine, otjTThisTownPrintedCost, game.ZoneHand, cardRefs(targets...)).ManaValue()
	}
	if got := price(theirs); got != 5 {
		t.Errorf("only an opponent's permanent: %d, want 5", got)
	}
	if got := price(theirs, mine); got != 2 {
		t.Errorf("one of yours among the targets: %d, want 2", got)
	}
}

func TestOtjTThisTownReturnsUpToTwoNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "This Town Ain't Big Enough", otjTThisTownTypeLine, otjTThisTownOracle, cardRefs(land)); err == nil {
		t.Error("a land was accepted as a nonland permanent")
	}
	castCatalogSpell(t, g, "This Town Ain't Big Enough", otjTThisTownTypeLine, otjTThisTownOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("each permanent returns to its owner's hand")
	}
	if _, ok := battlefieldCard(g, land); !ok {
		t.Error("the land stays")
	}
}

// --- Throw from the Saddle ---------------------------------------------------

func TestOtjTThrowFromTheSaddlePumpsOrCountersThenBites(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		counters int
	}{
		{"not a Mount", "Creature — Bear", 0},
		{"a Mount", "Creature — Horse Mount", 1},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "My Beast", tc.typeLine, 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		advanceToMain(t, g)
		castCatalogSpell(t, g, "Throw from the Saddle", "Sorcery", otjTThrowOracle, []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
		})
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
			t.Errorf("%s: the biter is %d/%d, want 3/3", tc.name, p, tough)
		}
		if got := plusOneCounters(g, mine); got != tc.counters {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.name, got, tc.counters)
		}
		if got := otjAC2Damage(t, g, theirs); got != 3 {
			t.Errorf("%s: the Giant has %d damage, want 3", tc.name, got)
		}
		if got := otjAC2Damage(t, g, mine); got != 0 {
			t.Errorf("%s: one-sided, but the biter took %d", tc.name, got)
		}
	}
}

// --- Thunder Lasso -----------------------------------------------------------

func TestOtjTThunderLassoAttachesOnEntryAndTapsAnAttackedPlayersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := seedBear(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	lasso := castCatalogSpell(t, g, "Thunder Lasso", "Artifact — Equipment", otjTLassoOracle, nil)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, lasso); host.ID != bear {
		t.Fatalf("the Lasso did not attach to the Bear: %+v", host)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}

	declareAttack(t, g, opp.ID, bear)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, bystander) || hasID(p.PickTargetCards, bear) {
			t.Errorf("offered a creature the defending player does not control: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("the defending player's creature was not tapped")
	}
	if findBattlefieldCardForTest(g, bystander).Tapped {
		t.Error("another player's creature was tapped")
	}
	spec, _ := Lookup(otjTLassoOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "Equip {2}" {
		t.Errorf("activated rows %+v, want Equip {2}", spec.Activated)
	}
}

// --- Thunder Salvo -----------------------------------------------------------

func TestOtjTThunderSalvoCountsYourOtherSpellsThisTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before int
		want   int
	}{
		{"its own cast only", 0, 2},
		{"after one other spell", 1, 3},
	} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)
		advanceToMain(t, g)
		for range tc.before {
			otjB1BoltOpponent(t, g, opp.ID)
		}
		castCatalogSpell(t, g, "Thunder Salvo", "Instant", otjTSalvoOracle, b16TargetCard(big))
		passPriorityAroundTable(t, g)
		if got := otjAC2Damage(t, g, big); got != tc.want {
			t.Errorf("%s: %d damage, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Tinybones Joins Up ------------------------------------------------------

func TestOtjTTinybonesJoinsUpMakesTheChosenPlayersDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{me, opp, third, fourth} {
		seedLibrary(p, "Card in hand")
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Tinybones Joins Up", "Legendary Enchantment", otjTTinybonesJoinsUp, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: third.ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) == nil || discardChoiceFor(g, third.ID) == nil {
		t.Fatal("each chosen player is asked to discard")
	}
	if discardChoiceFor(g, me.ID) != nil || discardChoiceFor(g, fourth.ID) != nil {
		t.Error("an unchosen player was asked to discard")
	}
}

func TestOtjTTinybonesJoinsUpMillsAndDrainsOnALegendaryCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	seedLibrary(opp, "A", "B", "C")
	seedLibrary(third, "A", "B", "C")
	pushCatalogPermanent(g, me.ID, "Tinybones Joins Up", "Legendary Enchantment", otjTTinybonesJoinsUp, false)
	advanceToMain(t, g)

	msc17Play(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a nonlegendary creature triggered it: %+v", p)
	}

	oppLib, thirdLib, oppLife, thirdLife, myLife := opp.Library.Size(), third.Library.Size(), opp.Life, third.Life, me.Life
	msc17Play(t, g, me, game.Card{Name: "A Legend", TypeLine: "Legendary Creature — Human Rogue", ManaCost: "{0}", Power: 1, Toughness: 1})
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 1 {
		t.Errorf("the chosen player milled %d, want 1", got)
	}
	if got := oppLife - opp.Life; got != 1 {
		t.Errorf("the chosen player lost %d life, want 1", got)
	}
	if third.Library.Size() != thirdLib || third.Life != thirdLife || me.Life != myLife {
		t.Error("an unchosen player milled or lost life")
	}
}

// --- Tomb Trawler ------------------------------------------------------------

func TestOtjTTombTrawlerBottomsACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trawler := b12Push(g, me.ID, "Tomb Trawler", "Artifact Creature — Golem", otjTTombTrawlerOracle, 0, 4)
	seedLibrary(me, "Top")
	dead := pushGraveyardPermanent(me, "Dead Thing", "Creature — Bear", "{1}{G}")
	theirs := pushGraveyardPermanent(opp, "Their Dead Thing", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, trawler, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
	b16Activate(t, g, me.ID, trawler, 0, game.ActivateAbilityParams{Targets: cardRefs(dead)})
	if me.Graveyard.Contains(dead) {
		t.Error("the card leaves the graveyard")
	}
	if got := libraryBottom(me, 1); len(got) != 1 || got[0] != dead {
		t.Errorf("library bottom %v, want the card", got)
	}
}

// --- Trash the Town ----------------------------------------------------------

func TestOtjTTrashTheTownAllThreeBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castModal(t, g, "Trash the Town", "Instant", otjTTrashTheTownOracle, []int{0, 1, 2}, []game.TargetRef{
		modeRef(game.TargetCard, bear, 0, 0),
		modeRef(game.TargetCard, bear, 1, 0),
		modeRef(game.TargetCard, bear, 2, 0),
	})
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature gains trample")
	}
	hand := me.Hand.Size()
	lciB1Attack(t, g, bear, opp.ID)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("combat damage to a player drew %d, want 2", got)
	}
}

func TestOtjTTrashTheTownGrantsTheDrawToItsTargetOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castModal(t, g, "Trash the Town", "Instant", otjTTrashTheTownOracle, []int{2}, []game.TargetRef{
		modeRef(game.TargetCard, bear, 0, 0),
	})
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 0 {
		t.Errorf("an unchosen bullet put %d counters", got)
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("an unchosen bullet granted trample")
	}
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("two creatures connected, only one with the grant: drew %d, want 2", got)
	}
}

// --- Treasure Dredger --------------------------------------------------------

func TestOtjTTreasureDredgerPaysOneLifeForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dredger := b12Push(g, me.ID, "Treasure Dredger", "Creature — Human Rogue", otjTDredgerOracle, 2, 2)
	advanceToMain(t, g)
	life := me.Life
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, dredger, 0, game.ActivateAbilityParams{})
	if got := treasuresOf(g, me.ID); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
	if got := life - me.Life; got != 1 {
		t.Errorf("paid %d life, want 1", got)
	}
	if !b16Tapped(t, g, dredger) {
		t.Error("{T} is part of the cost")
	}
}

// --- Trick Shot --------------------------------------------------------------

func TestOtjTTrickShotHitsACreatureAndAnotherToken(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 7, 7)
	tok := TokenCard("1/1 blue Fish")
	tok.Power, tok.Toughness = 3, 3
	fish := pushToken(g, opp.ID, tok)
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Trick Shot", "Instant", otjTTrickShotOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: fish, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	}); err == nil {
		t.Error("the same token was accepted for both targets; the second is \"other\"")
	}
	castCatalogSpell(t, g, "Trick Shot", "Instant", otjTTrickShotOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: big, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := otjAC2Damage(t, g, big); got != 6 {
		t.Errorf("the creature has %d damage, want 6", got)
	}
	if got := otjAC2Damage(t, g, fish); got != 2 {
		t.Errorf("the token has %d damage, want 2", got)
	}
}

// --- Tumbleweed Rising -------------------------------------------------------

func TestOtjTTumbleweedRisingMakesAnElementalAsBigAsYourBiggest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 6)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 9, 9)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Tumbleweed Rising", "Sorcery", otjTTumbleweedOracle, nil)
	passPriorityAroundTable(t, g)
	var made []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.IsToken() && c.Controller == me.ID {
			made = append(made, c.InstanceID)
		}
	}
	if len(made) != 1 {
		t.Fatalf("%d Elementals, want 1", len(made))
	}
	if p, tough := effectivePTOf(t, g, made[0]); p != 4 || tough != 4 {
		t.Errorf("the Elemental is %d/%d, want 4/4", p, tough)
	}
	if c := layeredCard(t, g, made[0]); len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("the Elemental's colours are %v, want green", c.Colors)
	}
}
