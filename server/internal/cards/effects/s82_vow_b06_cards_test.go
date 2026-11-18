package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b06_cards_test.go — Innistrad: Crimson Vow, S82 batch 6 (#5495).
// Each test plays the real catalog entries through the engine.

const (
	vowB06Gown      = "63e6c80a-a7c7-4c98-b760-271c36425dd1"
	vowB06Finery    = "b7d2feda-1a0e-4b62-be6d-a8dec7d8ea31"
	vowB06Cradle    = "e0717287-9bff-446b-838d-6fc0adccc768"
	vowB06Hearse    = "87925405-072b-49d6-97c0-fc3a13c10d07"
	vowB06Window    = "3e159175-78e5-40df-8fe9-e694938e7222"
	vowB06Journal   = "22e0e822-c4aa-4ec1-b7e4-0c1c869ef71b"
	vowB06Lantern   = "f0c4414f-f153-40f6-a3b3-94bcbf1d3e29"
	vowB06Heron     = "f297648b-0dda-4c58-a3eb-fdabdfb42177"
	vowB06Voice     = "01e742ed-660c-43c6-b393-e589a54cbe1b"
	vowB06Purifier  = "d28eac97-f38a-4d2a-aeae-e758bd3ecf4e"
	vowB06Spirit    = "bc79eb3c-33e2-4820-a59c-92a6271fd725"
	vowB06Puppeteer = "7bb3c3b8-25b7-4c48-aa41-dcdcebbc3d5e"
	vowB06Waltzer   = "bc9ac394-15b6-4d95-acca-5dc247235d2b"
)

func TestVowB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB06Gown: "Bride's Gown", vowB06Finery: "Groom's Finery", vowB06Cradle: "Cradle of Safety",
		vowB06Hearse: "Honeymoon Hearse", vowB06Window: "Boarded Window", vowB06Journal: "Investigator's Journal",
		vowB06Lantern: "Lantern of the Lost", vowB06Heron: "Heron of Hope", vowB06Voice: "Voice of the Blessed",
		vowB06Purifier: "Markov Purifier", vowB06Spirit: "Fleeting Spirit", vowB06Puppeteer: "Creepy Puppeteer",
		vowB06Waltzer: "Markov Waltzer",
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

// vowB06CastStrictErr casts a card from hand under strict timing and legality,
// returning the announce error instead of failing the test.
func vowB06CastStrictErr(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func vowB06AddCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(id, kind, n) })
}

// --- Voice of the Blessed -------------------------------------------------

func TestVowB06VoiceGrowsOncePerLifeGainEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	voice := pushCatalogPermanent(g, me.ID, "Voice of the Blessed", "Creature — Spirit Cleric", vowB06Voice, false)
	gainLifeNow(t, g, voice, me, 3)
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, voice); n != 1 {
		t.Fatalf("a 3-life gain left %d counters, want 1", n)
	}
	gainLifeNow(t, g, voice, g.Seats[1], 2)
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, voice); n != 1 {
		t.Errorf("an opponent's life gain left %d counters, want 1", n)
	}
}

func TestVowB06VoiceKeywordsFollowTheCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	voice := pushCatalogPermanent(g, me.ID, "Voice of the Blessed", "Creature — Spirit Cleric", vowB06Voice, false)
	vowB06AddCounters(g, voice, game.CounterPlusOne, 3)
	for _, kw := range []string{"flying", "vigilance", "indestructible"} {
		if effectiveAbilitiesContain(t, g, voice, kw) {
			t.Errorf("three counters give %s", kw)
		}
	}
	vowB06AddCounters(g, voice, game.CounterPlusOne, 1)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectiveAbilitiesContain(t, g, voice, kw) {
			t.Errorf("four counters lack %s", kw)
		}
	}
	if effectiveAbilitiesContain(t, g, voice, "indestructible") {
		t.Error("four counters give indestructible")
	}
	vowB06AddCounters(g, voice, game.CounterPlusOne, 5)
	if effectiveAbilitiesContain(t, g, voice, "indestructible") {
		t.Error("nine counters give indestructible")
	}
	vowB06AddCounters(g, voice, game.CounterPlusOne, 1)
	if !effectiveAbilitiesContain(t, g, voice, "indestructible") {
		t.Error("ten counters lack indestructible")
	}
	vowB06AddCounters(g, voice, game.CounterPlusOne, -7)
	for _, kw := range []string{"flying", "vigilance", "indestructible"} {
		if kw != "flying" && kw != "vigilance" && effectiveAbilitiesContain(t, g, voice, kw) {
			t.Errorf("three counters left give %s", kw)
		}
	}
	if effectiveAbilitiesContain(t, g, voice, "flying") {
		t.Error("removing counters below four left flying on")
	}
}

// --- Heron of Hope --------------------------------------------------------

func TestVowB06HeronAddsOneToEveryGainOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heron := pushCatalogPermanent(g, me.ID, "Heron of Hope", "Creature — Bird", vowB06Heron, false)
	before, oppBefore := me.Life, opp.Life
	gainLifeNow(t, g, heron, me, 2)
	if me.Life != before+3 {
		t.Errorf("a 2-life gain made %d, want 3", me.Life-before)
	}
	gainLifeNow(t, g, heron, opp, 2)
	if opp.Life != oppBefore+2 {
		t.Errorf("an opponent's 2-life gain made %d, want 2", opp.Life-oppBefore)
	}
}

func TestVowB06HeronLifelinkIsActivatedForOneTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heron := pushCatalogPermanent(g, me.ID, "Heron of Hope", "Creature — Bird", vowB06Heron, true)
	if effectiveAbilitiesContain(t, g, heron, "lifelink") {
		t.Fatal("the Heron starts with lifelink")
	}
	if err := g.ActivateCatalogAbility(me.ID, heron, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "W", "C")
	b16Activate(t, g, me.ID, heron, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, heron, "lifelink") {
		t.Error("the Heron lacks lifelink after {1}{W}")
	}
}

// --- Markov Purifier ------------------------------------------------------

func TestVowB06PurifierDrawsOnlyAfterGainingLifeAndPaying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pur := pushCatalogPermanent(g, me.ID, "Markov Purifier", "Creature — Vampire Cleric", vowB06Purifier, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil || pendingOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("asked to pay with no life gained")
	}
	_ = pur
}

func TestVowB06PurifierPaysTwoToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pur := pushCatalogPermanent(g, me.ID, "Markov Purifier", "Creature — Vampire Cleric", vowB06Purifier, false)
	advanceToMain(t, g)
	gainLifeNow(t, g, pur, me, 1)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	b06AddMana(me, "C", "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestVowB06PurifierDeclinedDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pur := pushCatalogPermanent(g, me.ID, "Markov Purifier", "Creature — Vampire Cleric", vowB06Purifier, false)
	advanceToMain(t, g)
	gainLifeNow(t, g, pur, me, 1)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	b06AddMana(me, "C", "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("declining drew: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Cradle of Safety -----------------------------------------------------

func TestVowB06CradleGivesPlusOneAndHexproofForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "U", "C")
	castCatalogSpell(t, g, "Cradle of Safety", "Enchantment — Aura", vowB06Cradle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 3/3", p, tt)
	}
	if !effectiveAbilitiesContain(t, g, bear, "hexproof") {
		t.Error("the enchanted creature lacks hexproof")
	}
}

func TestVowB06CradleCantEnchantAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "U", "C")
	err := vowB06CastStrictErr(t, g, "Cradle of Safety", "Enchantment — Aura", vowB06Cradle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if err == nil {
		t.Error("cast on an opponent's creature")
	}
}

// --- Markov Waltzer -------------------------------------------------------

func TestVowB06WaltzerPumpsUpToTwoAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Markov Waltzer", "Creature — Vampire", vowB06Waltzer, false)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	b17PickCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	got := effectivePower(t, g, a) + effectivePower(t, g, b)
	if got != 6 {
		t.Errorf("two Bears total %d power, want 6 (+1/+0 each)", got)
	}
}

// --- Creepy Puppeteer -----------------------------------------------------

func vowB06Attack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	opp := g.Seats[1]
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

func TestVowB06PuppeteerResetsTheOnlyOtherAttackerToFourThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pup := pushCatalogPermanent(g, me.ID, "Creepy Puppeteer", "Creature — Human Rogue", vowB06Puppeteer, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	vowB06Attack(t, g, pup, bear)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tt != 3 {
		t.Errorf("the other attacker is %d/%d, want 4/3", p, tt)
	}
}

func TestVowB06PuppeteerDeclinedLeavesTheOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pup := pushCatalogPermanent(g, me.ID, "Creepy Puppeteer", "Creature — Human Rogue", vowB06Puppeteer, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	vowB06Attack(t, g, pup, bear)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 1 {
		t.Errorf("a declined trigger left power %d, want 1", p)
	}
}

func TestVowB06PuppeteerNeedsExactlyOneOther(t *testing.T) {
	for name, extra := range map[string]int{"alone": 0, "two others": 2} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pup := pushCatalogPermanent(g, me.ID, "Creepy Puppeteer", "Creature — Human Rogue", vowB06Puppeteer, false)
			attackers := []uuid.UUID{pup}
			var others []uuid.UUID
			for i := 0; i < extra; i++ {
				id := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
				attackers = append(attackers, id)
				others = append(others, id)
			}
			vowB06Attack(t, g, attackers...)
			if latestTriggerPromptFor(g, me.ID) != nil {
				t.Error("the trigger was offered")
			}
			passPriorityAroundTable(t, g)
			for _, id := range others {
				if p := effectivePower(t, g, id); p != 1 {
					t.Errorf("an attacker is %d power, want 1", p)
				}
			}
		})
	}
}

// --- Investigator's Journal ----------------------------------------------

func TestVowB06JournalEntersWithTheGreatestCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 2; i++ {
		b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	}
	for i := 0; i < 3; i++ {
		b12Creature(g, g.Seats[2].ID, "Theirs", "Creature — Bear", 1, 1)
	}
	b06AddMana(me, "C", "C")
	id := castCatalogSpell(t, g, "Investigator's Journal", "Artifact — Book Clue", vowB06Journal, nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, "suspect"); n != 3 {
		t.Errorf("%d suspect counters, want 3 (the most creatures any player controls)", n)
	}
}

func TestVowB06JournalDrawsByRemovingACounterOrBySacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	j := pushCatalogPermanent(g, me.ID, "Investigator's Journal", "Artifact — Book Clue", vowB06Journal, false)
	vowB06AddCounters(g, j, "suspect", 1)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, j, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || countersOn(g, j, "suspect") != 0 || !g.Battlefield.Contains(j) {
		t.Fatalf("hand %d→%d, counters %d: want a draw and the counter gone", hand, me.Hand.Size(), countersOn(g, j, "suspect"))
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(j) })
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, j, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated the counter ability with no suspect counter left")
	}
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, j, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || g.Battlefield.Contains(j) {
		t.Errorf("sacrifice: hand %d, still on battlefield %v", me.Hand.Size(), g.Battlefield.Contains(j))
	}
}

// --- Lantern of the Lost --------------------------------------------------

func vowB06Gravecard(owner *game.Player, name string) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Owner: owner.ID, Controller: owner.ID})
	return id
}

func TestVowB06LanternEntersAndExilesOneTargetedCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := vowB06Gravecard(me, "Mine")
	theirs := vowB06Gravecard(opp, "Theirs")
	b06AddMana(me, "C")
	castCatalogSpell(t, g, "Lantern of the Lost", "Artifact", vowB06Lantern, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || !me.Graveyard.Contains(mine) {
		t.Error("only the chosen card should be exiled")
	}
}

func TestVowB06LanternExilesEveryGraveyardAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lantern := pushCatalogPermanent(g, me.ID, "Lantern of the Lost", "Artifact", vowB06Lantern, false)
	a, b := vowB06Gravecard(me, "Mine"), vowB06Gravecard(opp, "Theirs")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, lantern, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || opp.Graveyard.Contains(b) {
		t.Error("a graveyard still holds a card")
	}
	if g.Battlefield.Contains(lantern) {
		t.Error("the Lantern was not exiled as a cost")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Fleeting Spirit ------------------------------------------------------

func TestVowB06SpiritDiscardsToBlinkBackAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sp := pushCatalogPermanent(g, me.ID, "Fleeting Spirit", "Creature — Spirit", vowB06Spirit, false)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sp, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no discard named")
	}
	if err := g.ActivateCatalogAbility(me.ID, sp, 1, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{me.Hand.Cards[0].InstanceID}}); err != nil {
		t.Fatalf("discard to blink: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sp) {
		t.Fatal("the Spirit is still on the battlefield")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Fleeting Spirit")) != 1 {
		t.Error("the Spirit did not return at the next end step")
	}
}

func TestVowB06SpiritNeedsThreeGraveyardCardsForFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sp := pushCatalogPermanent(g, me.ID, "Fleeting Spirit", "Creature — Spirit", vowB06Spirit, false)
	advanceToMain(t, g)
	b06AddMana(me, "W")
	one, two := vowB06Gravecard(me, "One"), vowB06Gravecard(me, "Two")
	if err := g.ActivateCatalogAbility(me.ID, sp, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{one, two}}); err == nil {
		t.Fatal("activated with only two cards exiled")
	}
	three, _ := vowB06Gravecard(me, "Three"), vowB06Gravecard(me, "Four")
	b16Activate(t, g, me.ID, sp, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{one, two, three}})
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, sp, "first strike") {
		t.Error("no first strike after paying")
	}
	if n := me.Graveyard.Size(); n != 1 {
		t.Errorf("graveyard holds %d cards, want 1 left of four", n)
	}
}

// --- Honeymoon Hearse -----------------------------------------------------

func TestVowB06HearseTapsTwoCreaturesToBecomeACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hearse := pushCatalogPermanent(g, me.ID, "Honeymoon Hearse", "Artifact — Vehicle", vowB06Hearse, false)
	g.WithWriteLock(func() {
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Power, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Toughness = 5, 5
	})
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	// One untapped creature is not enough.
	tapForTest(t, g, b)
	if err := g.ActivateCatalogAbility(me.ID, hearse, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated tapping one creature")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(b) })
	b16Activate(t, g, me.ID, hearse, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}})
	passPriorityAroundTable(t, g)
	card, _ := battlefieldCard(g, hearse)
	if !card.IsCreature() {
		t.Error("the Hearse is not a creature after the ability")
	}
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("the two creatures were not tapped")
	}
	if p := effectivePower(t, g, hearse); p != 5 {
		t.Errorf("the Hearse is %d power, want 5", p)
	}
}

// --- Boarded Window -------------------------------------------------------

func TestVowB06WindowShrinksOnlyCreaturesAttackingItsController(t *testing.T) {
	g := newCatalogGame(t)
	opp, third := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, opp.ID, "Boarded Window", "Artifact", vowB06Window, false)
	atMe := b12Creature(g, g.Seats[0].ID, "Raider", "Creature — Bear", 3, 3)
	atThird := b12Creature(g, g.Seats[0].ID, "Other Raider", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atMe, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(atThird, third.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if p := effectivePower(t, g, atMe); p != 2 {
		t.Errorf("a creature attacking the Window's controller has %d power, want 2", p)
	}
	if p := effectivePower(t, g, atThird); p != 3 {
		t.Errorf("a creature attacking someone else has %d power, want 3", p)
	}
}

func TestVowB06WindowIsExiledAtEndStepAfterFourDamageCombatAndNoncombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	window := pushCatalogPermanent(g, opp.ID, "Boarded Window", "Artifact", vowB06Window, false)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 3, 3)
	vowB06Attack(t, g, raider)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := g.DamageTakenThisTurn(opp.ID); got != 2 {
		t.Fatalf("combat damage taken %d, want 2 (3 power, -1 from the Window)", got)
	}
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(raider, opp.ID, 2) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(window) {
		t.Error("the Window is still on the battlefield after 2 combat and 2 noncombat damage")
	}
}

func TestVowB06WindowStaysAfterThreeDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	window := pushCatalogPermanent(g, opp.ID, "Boarded Window", "Artifact", vowB06Window, false)
	src := b12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 3) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(window) {
		t.Fatal("exiled after 3 damage")
	}
}

func TestVowB06WindowIsExiledAfterFourNoncombatDamageAtAnotherPlayersEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	window := pushCatalogPermanent(g, opp.ID, "Boarded Window", "Artifact", vowB06Window, false)
	src := b12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 2)
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 2)
	})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(window) {
		t.Error("the Window is still on the battlefield after 4 damage (it triggers on every player's end step)")
	}
}

// --- Bride's Gown and Groom's Finery -------------------------------------

func vowB06Wedding(t *testing.T) (g *game.Game, me *game.Player, gown, finery, a, b uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	gown = pushCatalogPermanent(g, me.ID, "Bride's Gown", "Artifact — Equipment", vowB06Gown, false)
	finery = pushCatalogPermanent(g, me.ID, "Groom's Finery", "Artifact — Equipment", vowB06Finery, false)
	a = b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b = b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	return
}

func TestVowB06GownAloneIsPlusTwoPlusZero(t *testing.T) {
	g, me, gown, _, a, _ := vowB06Wedding(t)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, gown, a)
	if p, tt := effectivePower(t, g, a), effectiveToughness(t, g, a); p != 4 || tt != 2 {
		t.Errorf("Gown alone: %d/%d, want 4/2", p, tt)
	}
	if effectiveAbilitiesContain(t, g, a, "first strike") {
		t.Error("the Gown alone gives first strike")
	}
}

func TestVowB06TheWeddingPairGivesTheBonusToBothWearers(t *testing.T) {
	g, me, gown, finery, a, b := vowB06Wedding(t)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, gown, a)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, finery, b)
	if p, tt := effectivePower(t, g, a), effectiveToughness(t, g, a); p != 4 || tt != 4 {
		t.Errorf("Gown wearer: %d/%d, want 4/4", p, tt)
	}
	if !effectiveAbilitiesContain(t, g, a, "first strike") || effectiveAbilitiesContain(t, g, a, "deathtouch") {
		t.Error("the Gown wearer should have first strike and not deathtouch")
	}
	if p, tt := effectivePower(t, g, b), effectiveToughness(t, g, b); p != 4 || tt != 4 {
		t.Errorf("Finery wearer: %d/%d, want 4/4", p, tt)
	}
	if !effectiveAbilitiesContain(t, g, b, "deathtouch") || effectiveAbilitiesContain(t, g, b, "first strike") {
		t.Error("the Finery wearer should have deathtouch and not first strike")
	}
	// Moving the Finery onto the Gown's wearer keeps the bonus on it.
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, finery, a)
	if p, tt := effectivePower(t, g, a), effectiveToughness(t, g, a); p != 6 || tt != 6 {
		t.Errorf("both on one creature: %d/%d, want 6/6 (each adds +2/+0 and +0/+2)", p, tt)
	}
	if tt := effectiveToughness(t, g, b); tt != 2 {
		t.Errorf("the creature that lost the Finery has toughness %d, want 2", tt)
	}
}

func TestVowB06TheBonusEndsWhenThePartnerLeaves(t *testing.T) {
	g, me, gown, finery, a, b := vowB06Wedding(t)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, gown, a)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, finery, b)
	g.WithWriteLock(func() { _ = g.UnattachForEffect(finery) })
	if tt := effectiveToughness(t, g, a); tt != 2 {
		t.Errorf("after the Finery came off: toughness %d, want 2", tt)
	}
	if effectiveAbilitiesContain(t, g, a, "first strike") {
		t.Error("first strike stayed after the Finery came off")
	}
}

func TestVowB06AnOpponentsCreatureWearingThePartnerDoesNotCount(t *testing.T) {
	g, me, gown, finery, a, _ := vowB06Wedding(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, gown, a)
	g.WithWriteLock(func() { _ = g.AttachForEffect(finery, game.TargetRef{Kind: game.TargetCard, ID: theirs}) })
	if tt := effectiveToughness(t, g, a); tt != 2 {
		t.Errorf("the partner on an opponent's creature still gave toughness %d, want 2", tt)
	}
}
