package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b9_cards_test.go — Teenage Mutant Ninja Turtles, batch 9 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB9LeadersTalentOracle = "134430b7-d566-485a-bd5c-7db8a285f4ca"
	tmtB9LeatherheadOracle   = "ee9b606a-a1fe-43c8-84d0-9bf7b012cc75"
	tmtB9MadameNullOracle    = "73413bd5-8c40-4d45-8f05-c20c39ee4017"
	tmtB9MikeyLeoOracle      = "a86117c1-b83d-4350-9a10-a23b7e67e1bd"
	tmtB9MondoGeckoOracle    = "fa3ae120-756d-4c52-91f9-235767539d67"
	tmtB9NorthWindOracle     = "b408c345-b308-4920-8494-4f204d856aa3"
	tmtB9FarmOracle          = "e05f1a43-16ce-4e88-b0ad-2202efb25516"
	tmtB9NunchakuOracle      = "6d819742-3d02-4d60-9b10-82caa49a9d49"
	tmtB9ParameciaOracle     = "9076a886-efe0-4c99-8661-f0fd7a7b3129"
	tmtB9PartyDudeOracle     = "8df7a58c-053f-4ead-a778-2747718e5f10"
	tmtB9PizzaFaceOracle     = "e71479c9-bb38-4112-bf8a-41b39a3ade51"
	tmtB9PunksOracle         = "d1edb950-a118-4b4d-a0ef-9a6683d74c9f"
	tmtB9KatanaOracle        = "e9685c61-e33d-4ae0-9dd3-7fd0df5eb46f"
)

func TestTmtB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB9LeadersTalentOracle: "Leader's Talent",
		tmtB9LeatherheadOracle:   "Leatherhead, Swamp Stalker",
		tmtB9MadameNullOracle:    "Madame Null, Power Broker",
		tmtB9MikeyLeoOracle:      "Mikey & Leo, Chaos & Order",
		tmtB9MondoGeckoOracle:    "Mondo Gecko",
		tmtB9NorthWindOracle:     "North Wind Avatar",
		tmtB9FarmOracle:          "Northampton Farm",
		tmtB9NunchakuOracle:      "Novel Nunchaku",
		tmtB9ParameciaOracle:     "Paramecia Coloniex",
		tmtB9PartyDudeOracle:     "Party Dude",
		tmtB9PizzaFaceOracle:     "Pizza Face, Gastromancer",
		tmtB9PunksOracle:         "Purple Dragon Punks",
		tmtB9KatanaOracle:        "Quintessential Katana",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// tmtB9Kill destroys a permanent and settles the stack.
func tmtB9Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	effectstest.B27Kill(g, id)
	effectstest.PassPriorityAroundTable(t, g)
}

// tmtB9ZoneOf is the kind of zone a card is in now, "" when it is nowhere.
func tmtB9ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

// tmtB9NamedIn reports whether a card with the name is in p's hand.
func tmtB9InHand(p *game.Player, name string) bool {
	_, ok := effectstest.HandCardNamed(p, name)
	return ok
}

// --- Purple Dragon Punks -------------------------------------------------------

func TestTmtB9PurpleDragonPunksManaPaysArtifactSpellsAndAbilitiesOnly(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	punks := effectstest.PushCatalogPermanent(g, me.ID, "Purple Dragon Punks", "Creature — Human Rogue", tmtB9PunksOracle, false)
	if err := g.ActivateManaAbility(me.ID, punks, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Purple Dragon Punks: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Fatalf("pool = %v, want {R}", got)
	}
	cost, _ := game.ParseCost("{R}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{R}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{R}"}
	shock := game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("refused a creature's ability")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("refused an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(shock)) {
		t.Error("paid for an instant")
	}
}

// --- Madame Null, Power Broker ----------------------------------------------

func TestTmtB9MadameNullPaysLifeEqualToPowerForThatManyCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Madame Null, Power Broker", "Legendary Creature — Demon Advisor", tmtB9MadameNullOracle, "{2}{B}", 1, 3, "B")
	life := me.Life
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	mutant := effectstest.TmtB1Named(g, "Mutant")[0]
	if got := effectstest.PlusOnes(g, mutant); got != 2 {
		t.Errorf("counters on the Mutant = %d, want 2", got)
	}
	if got := life - me.Life; got != 2 {
		t.Errorf("paid %d life, want 2", got)
	}

	// An opponent's creature entering asks nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || effectstest.LatestConfirmFor(g, opp.ID) != nil {
		t.Error("an opponent's creature entering asked to pay life")
	}
}

func TestTmtB9MadameNullDeclinedOrUnaffordablePutsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Madame Null, Power Broker", "Legendary Creature — Demon Advisor", tmtB9MadameNullOracle, "{2}{B}", 1, 3, "B")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.TmtB1Named(g, "Mutant")[0]
	if effectstest.PlusOnes(g, first) != 0 || me.Life != 40 {
		t.Errorf("declined: counters %d, life %d; want 0 and 40", effectstest.PlusOnes(g, first), me.Life)
	}

	// At 1 life a 2-power creature can't be paid for (CR 119.4).
	g.WithWriteLock(func() { me.Life = 1 })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range effectstest.TmtB1Named(g, "Mutant") {
		if effectstest.PlusOnes(g, id) != 0 {
			t.Errorf("a Mutant got %d counters with 1 life to pay 2", effectstest.PlusOnes(g, id))
		}
	}
	if me.Life != 1 {
		t.Errorf("life = %d, want 1", me.Life)
	}
}

// --- Mikey & Leo, Chaos & Order ---------------------------------------------

func TestTmtB9MikeyAndLeoDrawsOnceEachTurnForYourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Mikey & Leo, Chaos & Order", "Legendary Creature — Mutant Ninja Turtle", tmtB9MikeyLeoOracle, "{G/W}{G/W}", 2, 2, "G", "W")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A counter an opponent puts on my creature, and one I put on theirs, draw nothing.
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(opp.ID, bear, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Fatalf("drew %d for counters I did not put on a creature I control", got)
	}

	// Any kind I put on my creature draws one.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, bear, "shield", 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d for my shield counter, want 1", got)
	}
	// Only once each turn.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, bear, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d after a second placement this turn, want 1", got)
	}
}

// --- Leatherhead, Swamp Stalker ----------------------------------------------

func TestTmtB9LeatherheadEntersWithAHexproofCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	lh := effectstest.RfdCast(t, g, "Leatherhead, Swamp Stalker", "Legendary Creature — Crocodile Mutant Rogue", tmtB9LeatherheadOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Counters(g, lh, game.CounterHexproof); got != 1 {
		t.Fatalf("hexproof counters = %d, want 1", got)
	}
	effectstest.AssertKeywords(t, g, lh, "hexproof", "trample")
}

func TestTmtB9LeatherheadRemovesACounterToDestroyThatPlayersArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	lh := effectstest.B27Push(g, me.ID, "Leatherhead, Swamp Stalker", "Legendary Creature — Crocodile Mutant Rogue", tmtB9LeatherheadOracle, "{2}{G}{G}", 5, 4, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(lh, game.CounterHexproof, 1) })
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	charm := effectstest.B12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	elsewhere := effectstest.B12Permanent(g, other.ID, "Other Rock", "Artifact")

	effectstest.AttackWith(t, g, opp.ID, lh)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no reflexive destroy target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(elsewhere)); err == nil {
		t.Fatal("the reflexive trigger targeted an artifact the damaged player does not control")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(rock)...)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the damaged player's artifact survived")
	}
	if !g.Battlefield.Contains(elsewhere) || !g.Battlefield.Contains(charm) {
		t.Error("a permanent that was not the target was destroyed")
	}
	if got := effectstest.TmtB9Counters(g, lh, game.CounterHexproof); got != 0 {
		t.Errorf("hexproof counters after the removal = %d, want 0", got)
	}
}

func TestTmtB9LeatherheadAsksWhichCounterAndDoesNothingWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lh := effectstest.B27Push(g, me.ID, "Leatherhead, Swamp Stalker", "Legendary Creature — Crocodile Mutant Rogue", tmtB9LeatherheadOracle, "{2}{G}{G}", 5, 4, "G")
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(lh, game.CounterHexproof, 1)
		_ = g.AddCounterForEffect(lh, game.CounterPlusOne, 1)
	})
	charm := effectstest.B12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	effectstest.AttackWith(t, g, opp.ID, lh)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil {
		t.Fatalf("no counter-kind question with two kinds on her: %+v", g.PendingChoices)
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Remove a +1/+1 counter" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no +1/+1 option among %+v", pick.PickOptions)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, charm)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(charm) {
		t.Error("the enchantment survived")
	}
	if effectstest.PlusOnes(g, lh) != 0 || effectstest.TmtB9Counters(g, lh, game.CounterHexproof) != 1 {
		t.Errorf("counters: +1/+1 %d, hexproof %d; want 0 and 1", effectstest.PlusOnes(g, lh), effectstest.TmtB9Counters(g, lh, game.CounterHexproof))
	}

	// A Leatherhead with no counter destroys nothing.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	bare := effectstest.B27Push(g2, me2.ID, "Leatherhead, Swamp Stalker", "Legendary Creature — Crocodile Mutant Rogue", tmtB9LeatherheadOracle, "{2}{G}{G}", 5, 4, "G")
	rock := effectstest.B12Permanent(g2, opp2.ID, "Their Rock", "Artifact")
	effectstest.AttackWith(t, g2, opp2.ID, bare)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.LatestPickTarget(g2, me2.ID) != nil || !g2.Battlefield.Contains(rock) {
		t.Error("a Leatherhead with no counter still destroyed something")
	}
}

// --- Mondo Gecko -------------------------------------------------------------

func TestTmtB9MondoGeckoBecomesTheChosenColorWithHexproofFromIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gecko := effectstest.B27Push(g, me.ID, "Mondo Gecko", "Legendary Creature — Lizard Mutant", tmtB9MondoGeckoOracle, "{1}{U}{U}", 2, 3, "U")
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}") })
	if err := g.ActivateCatalogAbility(me.ID, gecko, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(card) {
		t.Fatal("the discard was not paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, gecko); got != "R" {
		t.Errorf("colors = %q, want only red", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, gecko, game.HexproofFromColor("R")) {
		t.Errorf("no hexproof from red: %v", effectstest.EffectiveAbilities(t, g, gecko))
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if got := effectstest.TmtB9Colors(t, g, gecko); got != "U" || effectstest.EffectiveAbilitiesContain(t, g, gecko, game.HexproofFromColor("R")) {
		t.Errorf("after the turn: colors %q, abilities %v; want blue and no hexproof", got, effectstest.EffectiveAbilities(t, g, gecko))
	}
}

func TestTmtB9MondoGeckoDrawsForEachColorAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gecko := effectstest.B27Push(g, me.ID, "Mondo Gecko", "Legendary Creature — Lizard Mutant", tmtB9MondoGeckoOracle, "{1}{U}{U}", 2, 3, "U")
	effectstest.B27Push(g, me.ID, "Gold Thing", "Enchantment", "", "", 0, 0, "B", "G")
	effectstest.B27Push(g, opp.ID, "Their White Thing", "Enchantment", "", "", 0, 0, "W")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, gecko)
	effectstest.PassPriorityAroundTable(t, g)
	// Blue, black and green; the opponent's white does not count.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// --- North Wind Avatar -------------------------------------------------------

func TestTmtB9NorthWindAvatarTakesACardFromOutsideWhenCast(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "North Wind Avatar", "Creature — Dragon Spirit Avatar", tmtB9NorthWindOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no outside-the-game pick: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want a \"may\" of one", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !tmtB9InHand(me, "Sol Ring") {
		t.Error("Sol Ring did not come into the hand")
	}
}

func TestTmtB9NorthWindAvatarDoesNothingWhenNotCast(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	id := effectstest.GyCardWithOracle(me, "North Wind Avatar", "Creature — Dragon Spirit Avatar", tmtB9NorthWindOracle)
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, me.ID, false, game.KeyedTail{})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestCollectionPickFor(g, me.ID) != nil {
		t.Error("a reanimated North Wind Avatar offered a card from outside the game")
	}
}

// --- Northampton Farm --------------------------------------------------------

func TestTmtB9NorthamptonFarmExilesYourCreaturesAndReturnsOneAndTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	farm := effectstest.PushCatalogPermanent(g, me.ID, "Northampton Farm", "Land", tmtB9FarmOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, farm, 0, "{C}", effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Northampton Farm exiled a creature its controller does not own")
	}
	for _, id := range []uuid.UUID{a, b} {
		effectstest.Untap(g, farm)
		if err := effectstest.HobB4Activate(t, g, me.ID, farm, 0, "{C}", effectstest.CardRefs(id)); err != nil {
			t.Fatalf("exile: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tmtB9ZoneOf(g, id) != game.ZoneExile {
			t.Fatalf("%s was not exiled", id)
		}
	}
	// A card exiled some other way is not "exiled with this land".
	stranger := effectstest.B12Creature(g, me.ID, "Stranger", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(stranger) })

	effectstest.Untap(g, farm)
	if err := effectstest.HobB4Activate(t, g, me.ID, farm, 1, "{C}{C}", nil); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if g.Battlefield.Contains(farm) {
		t.Fatal("the Farm was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no choice of creature card with two exiled: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, me.ID, "Bear A") != 1 {
		t.Error("Bear A did not return to the battlefield under my control")
	}
	if !tmtB9InHand(me, "Bear B") || effectstest.TmtB1Count(g, me.ID, "Bear B") != 0 {
		t.Error("Bear B did not return to its owner's hand")
	}
	if tmtB9ZoneOf(g, stranger) != game.ZoneExile {
		t.Error("a card the Farm did not exile left exile")
	}
}

func TestTmtB9NorthamptonFarmWithOneCreatureReturnsItWithoutAQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	farm := effectstest.PushCatalogPermanent(g, me.ID, "Northampton Farm", "Land", tmtB9FarmOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Lone Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, farm, 0, "{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Untap(g, farm)
	if err := effectstest.HobB4Activate(t, g, me.ID, farm, 1, "{C}{C}", nil); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("asked which creature with only one")
	}
	if effectstest.TmtB1Count(g, me.ID, "Lone Bear") != 1 {
		t.Error("the creature did not come back")
	}
}

// --- Novel Nunchaku ----------------------------------------------------------

func TestTmtB9NovelNunchakuAttachesThenTheEquippedCreatureFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	nunchaku := effectstest.CastCatalogSpell(t, g, "Novel Nunchaku", "Artifact — Equipment", tmtB9NunchakuOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TmtB1Card(t, g, nunchaku); c.AttachedTo.ID != bear {
		t.Fatalf("the Nunchaku is attached to %v, want the Bear", c.AttachedTo)
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no reflexive fight target prompt")
	}
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("the fight targeted a creature I control")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(theirs)...)
	effectstest.PassPriorityAroundTable(t, g)
	// The Bear is a 3/3 with the Nunchaku: both deal 3.
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature survived the fight")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the equipped creature took no damage back")
	}
}

func TestTmtB9NovelNunchakuPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Novel Nunchaku", "Artifact — Equipment", tmtB9NunchakuOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	effectstest.PassPriorityAroundTable(t, g)
	// Fight no one: "up to one".
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the fight: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped power = %d, want 3", got)
	}
	effectstest.AssertKeywords(t, g, bear, "trample")
}

// --- Paramecia Coloniex ------------------------------------------------------

func TestTmtB9ParameciaColoniexMillsThreeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Paramecia Coloniex", "Creature — Zombie Worm", tmtB9ParameciaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

func TestTmtB9ParameciaColoniexExilesItselfToTuckACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worm := effectstest.B27Push(g, me.ID, "Paramecia Coloniex", "Creature — Zombie Worm", tmtB9ParameciaOracle, "{1}{B}", 2, 2, "B")
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	tmtB9Kill(t, g, worm)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if tmtB9ZoneOf(g, worm) != game.ZoneExile {
		t.Fatal("Paramecia Coloniex was not exiled")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(dead)...)
	effectstest.PassPriorityAroundTable(t, g)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != dead {
		t.Errorf("library top = %v (%v), want the creature card", top.Name, err)
	}
}

func TestTmtB9ParameciaColoniexDeclinedStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worm := effectstest.B27Push(g, me.ID, "Paramecia Coloniex", "Creature — Zombie Worm", tmtB9ParameciaOracle, "{1}{B}", 2, 2, "B")
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	tmtB9Kill(t, g, worm)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if tmtB9ZoneOf(g, worm) != game.ZoneGraveyard || tmtB9ZoneOf(g, dead) != game.ZoneGraveyard {
		t.Error("declining still moved a card")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the reflexive trigger happened without the exile")
	}
}

// --- Party Dude --------------------------------------------------------------

func TestTmtB9PartyDudeGivesEveryPlayerAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Party Dude", "Enchantment — Class", tmtB9PartyDudeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := effectstest.TmtB1Count(g, p.ID, "Food"); got != 1 {
			t.Errorf("%s has %d Foods, want 1", p.Name, got)
		}
	}
}

func TestTmtB9PartyDudeLevelTwoDrawsForOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dude := effectstest.PushClass(g, me.ID, "Party Dude", "Enchantment — Class", tmtB9PartyDudeOracle)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs2 := effectstest.B12Permanent(g, opp.ID, "Their Other Rock", "Artifact")

	hand := me.Hand.Size()
	tmtB9Kill(t, g, theirs)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Fatalf("drew %d at level 1", got)
	}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, dude, 0, "{1}{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	hand = me.Hand.Size()
	tmtB9Kill(t, g, mine)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d for my own artifact", got)
	}
	tmtB9Kill(t, g, theirs2)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d for an opponent's artifact, want 1", got)
	}
}

func TestTmtB9PartyDudeLevelThreePumpsAnAttackerByYourHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dude := effectstest.PushClass(g, me.ID, "Party Dude", "Enchantment — Class", tmtB9PartyDudeOracle)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, dude, 0, "{1}{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, dude, 1, "{4}{G}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	}
	effectstest.PassPriorityAroundTable(t, g)
	x := me.Hand.Size()
	if got := effectstest.EffectivePower(t, g, bear); got != 2+x {
		t.Errorf("attacker power = %d, want %d (2 + %d cards in hand)", got, 2+x, x)
	}
}

// --- Leader's Talent ---------------------------------------------------------

func TestTmtB9LeadersTalentCountersAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushClass(g, me.ID, "Leader's Talent", "Enchantment — Class", tmtB9LeadersTalentOracle)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, second)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(other)); err == nil {
		t.Fatal("a creature that is not attacking was a legal target")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	effectstest.PassPriorityAroundTable(t, g)
	// One trigger for the declaration, not one per attacker.
	if effectstest.PlusOnes(g, bear) != 1 || effectstest.PlusOnes(g, second) != 0 || effectstest.PlusOnes(g, other) != 0 {
		t.Errorf("counters: target %d, other attacker %d, home %d; want 1, 0, 0", effectstest.PlusOnes(g, bear), effectstest.PlusOnes(g, second), effectstest.PlusOnes(g, other))
	}
}

func TestTmtB9LeadersTalentLevelTwoGainsForACreatureThatHadACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	talent := effectstest.PushClass(g, me.ID, "Leader's Talent", "Enchantment — Class", tmtB9LeadersTalentOracle)
	marked := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	early := effectstest.B12Creature(g, me.ID, "Early Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(marked, "shield", 1)
		_ = g.AddCounterForEffect(early, game.CounterPlusOne, 1)
	})
	tmtB9Kill(t, g, early)
	if me.Life != 40 {
		t.Fatalf("gained life at level 1: %d", me.Life)
	}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 0, "{2}{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	tmtB9Kill(t, g, plain)
	if me.Life != 40 {
		t.Errorf("gained %d for a creature with no counter", me.Life-40)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(marked) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 42 {
		t.Errorf("life = %d after a creature with a counter left, want 42", me.Life)
	}
}

func TestTmtB9LeadersTalentLevelThreeCountersEachCreatureOnAnySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	talent := effectstest.PushClass(g, me.ID, "Leader's Talent", "Enchantment — Class", tmtB9LeadersTalentOracle)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ml := effectstest.B27Push(g, me.ID, "Mikey & Leo, Chaos & Order", "Legendary Creature — Mutant Ninja Turtle", tmtB9MikeyLeoOracle, "{G/W}{G/W}", 2, 2, "G", "W")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 0, "{2}{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 1, "{3}{W}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.CastFree(t, g, me, "Divination", "Sorcery")
	for _, id := range []uuid.UUID{a, b, ml} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%s has %d counters, want 1", id, got)
		}
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	// Mikey & Leo: three creatures got counters at once, one draw.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Mikey & Leo drew %d for one batch of counters, want 1", got)
	}
}

// --- Pizza Face, Gastromancer ------------------------------------------------

func TestTmtB9PizzaFaceMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.RfdCast(t, g, "Pizza Face, Gastromancer", "Legendary Artifact Creature — Food Mutant", tmtB9PizzaFaceOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Fatalf("Foods = %d, want 1", got)
	}
}

func TestTmtB9PizzaFaceSacrificesForFifteenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pizza := effectstest.B27Push(g, me.ID, "Pizza Face, Gastromancer", "Legendary Artifact Creature — Food Mutant", tmtB9PizzaFaceOracle, "{3}{B}{G}", 2, 4, "B", "G")
	if err := effectstest.HobB4Activate(t, g, me.ID, pizza, 0, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 55 || g.Battlefield.Contains(pizza) {
		t.Errorf("life %d, Pizza Face on the battlefield %v; want 55 and sacrificed", me.Life, g.Battlefield.Contains(pizza))
	}
}

func TestTmtB9PizzaFaceDisappearAnimatesANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pizza := effectstest.B27Push(g, me.ID, "Pizza Face, Gastromancer", "Legendary Artifact Creature — Food Mutant", tmtB9PizzaFaceOracle, "{3}{B}{G}", 2, 4, "B", "G")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	tmtB9Kill(t, g, fodder)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no Disappear target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(pizza)); err == nil {
		t.Fatal("Pizza Face targeted itself")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(rock)...)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Mutant") {
		t.Fatalf("the Rock is %q, want an artifact creature Mutant", c.TypeLine)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, rock); got != 3 {
		t.Errorf("the Rock's power = %d, want 3 (0/0 with three counters)", got)
	}
	// The change has no duration.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if c := effectstest.LayeredCard(t, g, rock); !c.IsCreature() {
		t.Error("the Rock stopped being a creature at end of turn")
	}
}

func TestTmtB9PizzaFaceDisappearNeedsAPermanentToHaveLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Pizza Face, Gastromancer", "Legendary Artifact Creature — Food Mutant", tmtB9PizzaFaceOracle, "{3}{B}{G}", 2, 4, "B", "G")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, bear) != 0 {
		t.Error("Disappear fired with no permanent having left the battlefield")
	}
}

// --- Quintessential Katana ---------------------------------------------------

func TestTmtB9QuintessentialKatanaUntapsAndGainsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	katana := effectstest.PushCatalogPermanent(g, me.ID, "Quintessential Katana", "Artifact — Equipment", tmtB9KatanaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, katana, 0, "{C}{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Fatalf("equipped power = %d, want 3", got)
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TmtB1Card(t, g, bear); c.Tapped {
		t.Error("the equipped creature did not untap")
	}
	if me.Life != 42 {
		t.Errorf("life = %d, want 42", me.Life)
	}
	if opp.Life != 37 {
		t.Errorf("opponent life = %d, want 37", opp.Life)
	}
}

func TestTmtB9QuintessentialKatanaMayAttachToAnEnteringNinja(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	katana := effectstest.PushCatalogPermanent(g, me.ID, "Quintessential Katana", "Artifact — Equipment", tmtB9KatanaOracle, false)
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a non-Ninja entering offered the attach")
	}
	effectstest.CastFromHandForTest(t, g, me, "Foot Soldier", "Creature — Human Ninja", "", "", game.CastSpellParams{FromZone: "hand"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	ninja := effectstest.TmtB1Named(g, "Foot Soldier")
	if len(ninja) != 1 {
		t.Fatal("no Ninja on the battlefield")
	}
	if c := effectstest.TmtB1Card(t, g, katana); c.AttachedTo.ID != ninja[0] {
		t.Errorf("the Katana is attached to %v, want the Ninja", c.AttachedTo)
	}
}
