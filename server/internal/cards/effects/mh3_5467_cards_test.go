package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mh3_5467_cards_test.go — Modern Horizons 3, batch 1 (#5467, S82):
// fifteen cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	mh3b1SerumVisionaryOracle      = "922d9d79-8038-4185-9dfb-2d2d52fd8460"
	mh3b1FaithfulWatchdogOracle    = "139fb542-89fd-4b9e-87e6-ec925bcca93a"
	mh3b1SnappingVoidcrawOracle    = "e495033b-f727-41e3-b7fc-7f5fa8abdd87"
	mh3b1SarpadianSimulacrumOracle = "b563a578-86ce-41a5-a83e-c70c6a4960e2"
	mh3b1BreatheYourLastOracle     = "dc3af7cb-03b9-445f-82eb-18798ec92934"
	mh3b1WingItOracle              = "7b0b05a0-0be4-498f-be54-8d68b818f90a"
	mh3b1GiftOfTheViperOracle      = "7b1a59db-0ffc-490b-a881-c8192f3747db"
	mh3b1AerieAuxiliaryOracle      = "0c97aea7-c997-45b4-a044-84fb5c07a021"
	mh3b1FurnaceHellkiteOracle     = "647980b2-088c-43a4-bdc4-5865f03a4fc4"
	mh3b1ScurrilousSentryOracle    = "52bfcb09-007c-47d8-8038-692e7dd8f922"
	mh3b1MandibularKiteOracle      = "bca9f65e-ea7a-4b8d-ab0f-c2672fa967f0"
	mh3b1ColossalDreadmaskOracle   = "1244902e-1a4e-4d3c-9a55-9be4602e3d2d"
	mh3b1UnfathomableTruthsOracle  = "63765157-e2ba-4cec-9501-8a405259de5b"
	mh3b1FowlStrikeOracle          = "390320f7-11b8-4b61-8316-36c396470a09"
	mh3b1GhostfireSliceOracle      = "986a7ad0-527a-4079-aa1b-c8acdf087020"

	mh3b1BearType   = "Creature — Bear"
	mh3b1SentryType = "Creature — Human Knight Rogue"
)

// mh3b1Creature seeds a non-catalog creature with printed colours and
// keywords.
func mh3b1Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int, colors, keywords []string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: mh3b1BearType,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Colors: colors, Keywords: keywords,
	})
}

// mh3b1HandNames lists the names of the cards in p's hand.
func mh3b1HandNames(p *game.Player) map[string]bool {
	out := map[string]bool{}
	for _, c := range p.Hand.Cards {
		out[c.Name] = true
	}
	return out
}

func TestMh3b1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mh3b1SerumVisionaryOracle:      "Serum Visionary",
		mh3b1FaithfulWatchdogOracle:    "Faithful Watchdog",
		mh3b1SnappingVoidcrawOracle:    "Snapping Voidcraw",
		mh3b1SarpadianSimulacrumOracle: "Sarpadian Simulacrum",
		mh3b1BreatheYourLastOracle:     "Breathe Your Last",
		mh3b1WingItOracle:              "Wing It",
		mh3b1GiftOfTheViperOracle:      "Gift of the Viper",
		mh3b1AerieAuxiliaryOracle:      "Aerie Auxiliary",
		mh3b1FurnaceHellkiteOracle:     "Furnace Hellkite",
		mh3b1ScurrilousSentryOracle:    "Scurrilous Sentry",
		mh3b1MandibularKiteOracle:      "Mandibular Kite",
		mh3b1ColossalDreadmaskOracle:   "Colossal Dreadmask",
		mh3b1UnfathomableTruthsOracle:  "Unfathomable Truths",
		mh3b1FowlStrikeOracle:          "Fowl Strike",
		mh3b1GhostfireSliceOracle:      "Ghostfire Slice",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, %d caveats", name, def.Name, def.Completeness, len(def.Caveats))
		}
	}
}

// The printed keywords reach the battlefield.
func TestMh3b1KeywordBodies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Sarpadian Simulacrum", "Artifact Creature — Goblin", mh3b1SarpadianSimulacrumOracle, "haste"},
		{"Aerie Auxiliary", "Creature — Bird Soldier", mh3b1AerieAuxiliaryOracle, "flying"},
		{"Furnace Hellkite", "Artifact Creature — Dragon", mh3b1FurnaceHellkiteOracle, "flying"},
		{"Scurrilous Sentry", mh3b1SentryType, mh3b1ScurrilousSentryOracle, "menace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !effectiveAbilitiesContain(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// The three devoid cards are colourless whatever their cost says.
func TestMh3b1DevoidCardsAreColorless(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Snapping Voidcraw", mh3b1SnappingVoidcrawOracle, "{1}{G}{U}"},
		{"Unfathomable Truths", mh3b1UnfathomableTruthsOracle, "{4}{U}"},
		{"Ghostfire Slice", mh3b1GhostfireSliceOracle, "{2}{R}"},
	} {
		c := game.Card{Name: tc.name, OracleID: tc.oracle, TypeLine: "Instant", ManaCost: tc.cost}
		if !c.IsColorless() {
			t.Errorf("%s is %v, want colorless", tc.name, c.EffectiveColors())
		}
	}
}

// --- Serum Visionary ---------------------------------------------------------

func TestMh3b1SerumVisionaryDrawsThenScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Top", "Second", "Third")
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Serum Visionary", "Creature — Vedalken Wizard", mh3b1SerumVisionaryOracle, 2, 2)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
	if !mh3b1HandNames(me)["Top"] {
		t.Error("the top card was not drawn")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt after the draw")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
	}
}

// --- Faithful Watchdog -------------------------------------------------------

func TestMh3b1FaithfulWatchdogEntersWithThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	dog := castHoldCreature(t, g, "Faithful Watchdog", "Creature — Dog", mh3b1FaithfulWatchdogOracle, 0, 0)
	if !g.Battlefield.Contains(dog) {
		t.Fatal("the 0/0 Watchdog did not survive its entry")
	}
	if got := countersOn(g, dog, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if c := apaLive(g, dog); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		p, tg := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("the Watchdog is %d/%d, want 3/3", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, dog, "vigilance") {
		t.Error("the Watchdog has no vigilance")
	}
}

// --- Snapping Voidcraw -------------------------------------------------------

func TestMh3b1SnappingVoidcrawTapsForTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	craw := b12Push(g, me.ID, "Snapping Voidcraw", "Creature — Eldrazi Turtle", mh3b1SnappingVoidcrawOracle, 1, 3)
	if err := g.ActivateManaAbility(me.ID, craw, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Errorf("pool holds %d colorless, want 2", got)
	}
	if !hmlTapped(t, g, craw) {
		t.Error("the Voidcraw did not tap for mana")
	}
}

func TestMh3b1SnappingVoidcrawDrawNeedsColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	craw := b12Push(g, me.ID, "Snapping Voidcraw", "Creature — Eldrazi Turtle", mh3b1SnappingVoidcrawOracle, 1, 3)
	apaMana(me, "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, craw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{C} was paid without colorless mana")
	}
	me.ManaPool.EmptyPool()
	apaMana(me, "C", "G", "G", "G")
	hand := me.Hand.Size()
	apaActivate(t, g, me, craw, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
	if !hmlTapped(t, g, craw) {
		t.Error("the Voidcraw did not tap")
	}
}

// --- Sarpadian Simulacrum ----------------------------------------------------

func TestMh3b1SarpadianSimulacrumSacrificesForFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := ons2Opponent(g)
	advanceToMain(t, g)
	sim := b12Push(g, me.ID, "Sarpadian Simulacrum", "Artifact Creature — Goblin", mh3b1SarpadianSimulacrumOracle, 1, 1)
	beast := b12Creature(g, opp.ID, "Their Beast", "Creature — Beast", 4, 4)
	apaMana(me, "R", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sim, 0, game.ActivateAbilityParams{
		Targets: playerRefs(opp.ID), Strict: true,
	}); err == nil {
		t.Fatal("the Simulacrum targeted a player")
	}
	if !g.Battlefield.Contains(sim) {
		t.Fatal("a refused activation sacrificed the Simulacrum")
	}
	apaActivate(t, g, me, sim, 0, game.ActivateAbilityParams{Targets: cardTarget(beast)})
	if g.Battlefield.Contains(sim) {
		t.Error("the Simulacrum was not sacrificed")
	}
	if g.Battlefield.Contains(beast) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- Breathe Your Last -------------------------------------------------------

func TestMh3b1BreatheYourLastGainsALifePerColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := ons2Opponent(g)
	gold := mh3b1Creature(g, opp.ID, "Gold Bear", 2, 2, []string{"B", "G"}, nil)
	life := me.Life
	castCatalogSpell(t, g, "Breathe Your Last", "Instant", mh3b1BreatheYourLastOracle, cardTarget(gold))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gold) {
		t.Error("the creature was not destroyed")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 for a two-colour creature", life, me.Life)
	}
}

func TestMh3b1BreatheYourLastColorlessGivesNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := ons2Opponent(g)
	golem := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Breathe Your Last", "Instant", mh3b1BreatheYourLastOracle, cardTarget(land)); err == nil {
		t.Fatal("Breathe Your Last targeted a land")
	}
	life := me.Life
	castCatalogSpell(t, g, "Breathe Your Last", "Instant", mh3b1BreatheYourLastOracle, cardTarget(golem))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the colorless creature was not destroyed")
	}
	if me.Life != life {
		t.Errorf("life %d -> %d, want no change for a colorless creature", life, me.Life)
	}
}

// --- Wing It -----------------------------------------------------------------

func TestMh3b1WingItPumpsGivesAFlyingCounterAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", mh3b1BearType, 2, 2)
	castCatalogSpell(t, g, "Wing It", "Instant", mh3b1WingItOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tg)
	}
	if got := countersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters, want 1", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
}

// --- Gift of the Viper -------------------------------------------------------

func TestMh3b1GiftOfTheViperPutsThreeCountersAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", mh3b1BearType, 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Tapped = true })
	castCatalogSpell(t, g, "Gift of the Viper", "Instant", mh3b1GiftOfTheViperOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	for _, kind := range []string{game.CounterPlusOne, game.CounterReach, game.CounterDeathtouch} {
		if got := countersOn(g, bear, kind); got != 1 {
			t.Errorf("%d %s counters, want 1", got, kind)
		}
	}
	for _, kw := range []string{"reach", "deathtouch"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("the Bear has no %s", kw)
		}
	}
	if p := apaLive(g, bear).CurrentPower(); p != 3 {
		t.Errorf("the Bear's power is %d, want 3", p)
	}
	if hmlTapped(t, g, bear) {
		t.Error("the Bear was not untapped")
	}
}

// --- Aerie Auxiliary ---------------------------------------------------------

func TestMh3b1AerieAuxiliarySupportsTwoOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := ons2Opponent(g)
	mine := b12Creature(g, me.ID, "My Bear", mh3b1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", mh3b1BearType, 2, 2)
	aux := castHoldCreature(t, g, "Aerie Auxiliary", "Creature — Bird Soldier", mh3b1AerieAuxiliaryOracle, 3, 3)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if hasID(p.PickTargetCards, aux) {
		t.Error("the Auxiliary was offered as its own support target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(mine, theirs)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, mine, game.CounterPlusOne) != 1 || countersOn(g, theirs, game.CounterPlusOne) != 1 {
		t.Error("support did not put a +1/+1 counter on each target")
	}
	if countersOn(g, aux, game.CounterPlusOne) != 0 {
		t.Error("the Auxiliary put a counter on itself")
	}
}

// --- Furnace Hellkite --------------------------------------------------------

func TestMh3b1FurnaceHellkiteHasAffinityForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	const typeLine, cost = "Artifact Creature — Dragon", "{5}{R}{R}"
	if got := selfPricedMV(t, g, me, mh3b1FurnaceHellkiteOracle, typeLine, cost); got != 7 {
		t.Errorf("with no artifacts: %d, want 7", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Bauble", TypeLine: "Artifact"})
	pushPermanent(g, me.ID, game.Card{Name: "Bauble", TypeLine: "Artifact"})
	pushPermanent(g, g.Seats[1].ID, game.Card{Name: "Their Bauble", TypeLine: "Artifact"})
	if got := selfPricedMV(t, g, me, mh3b1FurnaceHellkiteOracle, typeLine, cost); got != 5 {
		t.Errorf("with two artifacts of yours: %d, want 5", got)
	}
}

func TestMh3b1FurnaceHellkiteFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	kite := b12Push(g, me.ID, "Furnace Hellkite", "Artifact Creature — Dragon", mh3b1FurnaceHellkiteOracle, 5, 5)
	apaMana(me, "R")
	apaActivate(t, g, me, kite, 0, game.ActivateAbilityParams{})
	if p, tg := effectivePower(t, g, kite), effectiveToughness(t, g, kite); p != 6 || tg != 5 {
		t.Errorf("the Hellkite is %d/%d, want 6/5", p, tg)
	}
}

// --- Scurrilous Sentry -------------------------------------------------------

func TestMh3b1ScurrilousSentryConnivesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardForTest(me, "Spare Spell", "Sorcery", "")
	sentry := castHoldCreature(t, g, "Scurrilous Sentry", mh3b1SentryType, mh3b1ScurrilousSentryOracle, 2, 3)
	if got := me.Hand.Size(); got != 2 {
		t.Fatalf("hand holds %d, want the spare spell and the drawn card", got)
	}
	answerDiscard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sentry, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after discarding a nonland card, want 1", got)
	}
}

func TestMh3b1ScurrilousSentryConnivesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := ons2Opponent(g)
	advanceToMain(t, g)
	sentry := b12Push(g, me.ID, "Scurrilous Sentry", mh3b1SentryType, mh3b1ScurrilousSentryOracle, 2, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := handCardForTest(me, "Spare Land", "Basic Land — Forest", "")
	declareAttack(t, g, opp.ID, sentry)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Fatalf("hand holds %d, want the spare land and the drawn card", got)
	}
	answerDiscard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sentry, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after discarding a land, want 0", got)
	}
	if b12ZoneOf(g, land) != game.ZoneGraveyard {
		t.Error("the land was not discarded")
	}
}

// --- Mandibular Kite and Colossal Dreadmask ----------------------------------

func TestMh3b1MandibularKiteGermFliesAndReequips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kite, germ := castLivingWeapon(t, g, "Mandibular Kite", mh3b1MandibularKiteOracle)
	if host := attachmentHostOf(t, g, kite); host.ID != germ {
		t.Fatalf("the Kite is attached to %+v, want the Germ", host)
	}
	if p, tg := effectivePower(t, g, germ), effectiveToughness(t, g, germ); p != 1 || tg != 1 {
		t.Errorf("the Germ is %d/%d, want 1/1", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, germ, "flying") {
		t.Error("the Germ has no flying")
	}
	bear := b12Creature(g, me.ID, "Bear", mh3b1BearType, 2, 2)
	apaMana(me, "W", "C", "C", "C")
	apaActivate(t, g, me, kite, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if host := attachmentHostOf(t, g, kite); host.ID != bear {
		t.Fatalf("the Kite is attached to %+v after equip, want the Bear", host)
	}
	if g.Battlefield.Contains(germ) {
		t.Error("the 0/0 Germ survived with nothing attached")
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the equipped Bear has no flying")
	}
}

func TestMh3b1ColossalDreadmaskMakesASixSixTrampler(t *testing.T) {
	g := newCatalogGame(t)
	mask, germ := castLivingWeapon(t, g, "Colossal Dreadmask", mh3b1ColossalDreadmaskOracle)
	if host := attachmentHostOf(t, g, mask); host.ID != germ {
		t.Fatalf("the Dreadmask is attached to %+v, want the Germ", host)
	}
	if p, tg := effectivePower(t, g, germ), effectiveToughness(t, g, germ); p != 6 || tg != 6 {
		t.Errorf("the Germ is %d/%d, want 6/6", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, germ, "trample") {
		t.Error("the Germ has no trample")
	}
	spec, _ := Lookup(mh3b1ColossalDreadmaskOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Cost.Mana != "{3}{G}{G}" {
		t.Errorf("the equip row is %+v, want one Equip {3}{G}{G}", spec.Activated)
	}
}

// --- Unfathomable Truths -----------------------------------------------------

func TestMh3b1UnfathomableTruthsDrawsThreeAndMakesASpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Unfathomable Truths", "Instant", mh3b1UnfathomableTruthsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d, want three cards drawn", hand, got)
	}
	spawn := findBattlefieldByName(g, "Eldrazi Spawn")
	if spawn == uuid.Nil {
		t.Fatal("no Eldrazi Spawn token")
	}
	if c := b12Card(t, g, spawn); c.Controller != me.ID || c.Power != 0 || c.Toughness != 1 {
		t.Errorf("the Spawn is %d/%d under %s, want a 0/1 of yours", c.Power, c.Toughness, c.Controller)
	}
}

// --- Fowl Strike -------------------------------------------------------------

func TestMh3b1FowlStrikeDestroysOnlyAFlier(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	bird := mh3b1Creature(g, opp.ID, "Their Bird", 2, 2, nil, []string{"flying"})
	bear := b12Creature(g, opp.ID, "Their Bear", mh3b1BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Fowl Strike", "Instant", mh3b1FowlStrikeOracle, cardTarget(bear)); err == nil {
		t.Fatal("Fowl Strike targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Fowl Strike", "Instant", mh3b1FowlStrikeOracle, cardTarget(bird))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flier survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func TestMh3b1FowlStrikeReinforcesFromTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", mh3b1BearType, 2, 2)
	strike := b13HandCard(me, "Fowl Strike", "Instant", mh3b1FowlStrikeOracle)
	apaMana(me, "G", "C", "C")
	apaActivate(t, g, me, strike, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if b12ZoneOf(g, strike) != game.ZoneGraveyard {
		t.Error("Fowl Strike was not discarded to pay for reinforce")
	}
}

// --- Ghostfire Slice ---------------------------------------------------------

func TestMh3b1GhostfireSliceCostsLessAgainstAMulticoloredPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	const cost = "{2}{R}"
	if got := selfPricedMV(t, g, me, mh3b1GhostfireSliceOracle, "Instant", cost); got != 3 {
		t.Errorf("on an empty board: %d, want 3", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "My Gold", TypeLine: mh3b1BearType, Colors: []string{"W", "U"}})
	pushPermanent(g, g.Seats[1].ID, game.Card{Name: "Their Mono", TypeLine: mh3b1BearType, Colors: []string{"R"}})
	if got := selfPricedMV(t, g, me, mh3b1GhostfireSliceOracle, "Instant", cost); got != 3 {
		t.Errorf("with only your own gold permanent: %d, want 3", got)
	}
	pushPermanent(g, g.Seats[len(g.Seats)-1].ID, game.Card{Name: "Their Gold", TypeLine: "Enchantment", Colors: []string{"B", "G"}})
	if got := selfPricedMV(t, g, me, mh3b1GhostfireSliceOracle, "Instant", cost); got != 1 {
		t.Errorf("with an opponent's gold permanent: %d, want 1", got)
	}
}

func TestMh3b1GhostfireSliceDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	life := opp.Life
	castCatalogSpell(t, g, "Ghostfire Slice", "Instant", mh3b1GhostfireSliceOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}
