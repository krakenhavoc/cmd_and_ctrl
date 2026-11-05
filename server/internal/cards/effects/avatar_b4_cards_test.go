package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b4_cards_test.go — Avatar: The Last Airbender, batch 4 (#4459).
// Each test plays the real catalog entry and asserts what its text says.

const (
	avB4RiverSpiritOracle  = "12d91743-1ccb-4f2b-b2a9-1c2b34f7c8eb"
	avB4CrashingWaveOracle = "c8d3936f-4c08-4ffe-83c9-7ed26fb8f9cd"
	avB4EmberIslandOracle  = "0b786547-67ed-49fa-9981-35c4c2175f62"
	avB4FlyerOracle        = "cbbea4e2-4b36-4965-89b6-00d2a564f146"
	avB4FlexibleOracle     = "8cf88fa8-2f47-4125-a59d-9c2b2d4765a7"
	avB4FortuneOracle      = "a5a05edd-2a30-414b-b766-f0651d5783f5"
	avB4GeyserOracle       = "dfdcb451-a597-46fe-893c-49e6dbda5f71"
	avB4KoiOracle          = "c93bb65a-a74a-499b-a538-65272079aab8"
	avB4GranGranOracle     = "0259a9ed-0558-4183-b737-871ed5573c32"
	avB4HonestWorkOracle   = "12956c07-e679-471b-8dcb-2a2d15b0be55"
)

func TestAvB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB4RiverSpiritOracle:  "Benevolent River Spirit",
		avB4CrashingWaveOracle: "Crashing Wave",
		avB4EmberIslandOracle:  "Ember Island Production",
		avB4FlyerOracle:        "First-Time Flyer",
		avB4FlexibleOracle:     "Flexible Waterbender",
		avB4FortuneOracle:      "Forecasting Fortune Teller",
		avB4GeyserOracle:       "Geyser Leaper",
		avB4KoiOracle:          "Giant Koi",
		avB4GranGranOracle:     "Gran-Gran",
		avB4HonestWorkOracle:   "Honest Work",
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

// --- fixtures ------------------------------------------------------------------

// avB4Lesson puts a Lesson sorcery card into p's graveyard.
func avB4Lesson(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "A Lesson", TypeLine: "Sorcery — Lesson", Owner: p.ID, Controller: p.ID})
	})
	return id
}

// avB4Tapped reports whether a battlefield card is tapped.
func avB4Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// avB4SetTapped taps a battlefield card directly.
func avB4SetTapped(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

// avB4Activate activates a catalog ability strictly, waterbending with
// the listed permanents and paying the rest from the pool.
func avB4Activate(g *game.Game, p uuid.UUID, card uuid.UUID, idx int, waterbend []uuid.UUID) error {
	return g.ActivateCatalogAbility(p, card, idx, game.ActivateAbilityParams{Strict: true, WaterbendIDs: waterbend})
}

// --- Benevolent River Spirit -------------------------------------------------------

func TestAvB4RiverSpiritWaterbendsFiveOnTopOfItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tappers := pushTapCostSoldiers(g, me.ID, 5)
	fillPoolColored(me, "U", 2)
	spirit := castFromHandForTest(t, g, me, "Benevolent River Spirit", "Creature — Spirit", "{U}{U}", avB4RiverSpiritOracle,
		game.CastSpellParams{Strict: true, TapIDs: tappers})
	for _, id := range tappers {
		if !avB4Tapped(t, g, id) {
			t.Fatal("a creature tapped to waterbend is untapped")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spirit) {
		t.Fatal("Benevolent River Spirit did not resolve")
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 2 when Benevolent River Spirit entered")
	}
	answerScryKeepAll(t, g, me.ID)
	if !hasEffectiveKeyword(t, g, spirit, "flying") {
		t.Error("Benevolent River Spirit has no flying")
	}
}

func TestAvB4RiverSpiritCannotBeCastWithoutTheWholeWaterbend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tappers := pushTapCostSoldiers(g, me.ID, 4)
	fillPoolColored(me, "U", 2)
	if _, err := castFromHandErr(g, me, "Benevolent River Spirit", "Creature — Spirit", "{U}{U}", avB4RiverSpiritOracle,
		game.CastSpellParams{Strict: true, TapIDs: tappers}); err == nil {
		t.Fatal("cast for {U}{U} and four taps; waterbend {5} wants five")
	}
}

func TestAvB4RiverSpiritHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := pushCatalogPermanent(g, me.ID, "Benevolent River Spirit", "Creature — Spirit", avB4RiverSpiritOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, spirit)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting Benevolent River Spirit")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spirit) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Crashing Wave -----------------------------------------------------------------

func TestAvB4CrashingWaveTapsThenDistributesAmongAnyTappedOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	target := b12Creature(g, opp.ID, "Untapped Foe", "Creature — Bear", 2, 2)
	already := b12Creature(g, opp.ID, "Tapped Foe", "Creature — Bear", 2, 2)
	avB4SetTapped(g, already)
	untouched := b12Creature(g, opp.ID, "Other Foe", "Creature — Bear", 2, 2)
	tappers := pushTapCostSoldiers(g, me.ID, 1)
	fillPoolColored(me, "U", 2)
	castFromHandForTest(t, g, me, "Crashing Wave", "Sorcery", "{U}{U}", avB4CrashingWaveOracle, game.CastSpellParams{
		Strict: true, XValue: 1, TapIDs: tappers,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	passPriorityAroundTable(t, g)
	if !avB4Tapped(t, g, target) {
		t.Fatal("the target was not tapped")
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to distribute the stun counters")
	}
	for _, id := range pick.ChooseCards {
		if id == untouched || id == tappers[0] {
			t.Fatalf("an untapped creature or the caster's own creature is offered: %v", pick.ChooseCards)
		}
	}
	answerChooseCards(t, g, me.ID, target, already)
	answerChooseCards(t, g, me.ID, already)
	if got := countersOn(g, already, game.CounterStun); got != 2 {
		t.Errorf("the creature named for two has %d stun counters", got)
	}
	if got := countersOn(g, target, game.CounterStun); got != 1 {
		t.Errorf("the other chosen creature has %d stun counters, want 1", got)
	}
	if got := countersOn(g, untouched, game.CounterStun); got != 0 {
		t.Errorf("an unchosen creature has %d stun counters", got)
	}
}

func TestAvB4CrashingWaveWithOneTappedFoePutsAllThreeOnIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	foe := b12Creature(g, opp.ID, "Tapped Foe", "Creature — Bear", 2, 2)
	avB4SetTapped(g, foe)
	fillPoolColored(me, "U", 2)
	castFromHandForTest(t, g, me, "Crashing Wave", "Sorcery", "{U}{U}", avB4CrashingWaveOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, foe, game.CounterStun); got != 3 {
		t.Errorf("the only tapped foe has %d stun counters, want 3", got)
	}
}

func TestAvB4CrashingWaveSplitsOneEachOverThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	var foes []uuid.UUID
	for i := 0; i < 3; i++ {
		id := b12Creature(g, opp.ID, "Tapped Foe", "Creature — Bear", 2, 2)
		avB4SetTapped(g, id)
		foes = append(foes, id)
	}
	fillPoolColored(me, "U", 2)
	castFromHandForTest(t, g, me, "Crashing Wave", "Sorcery", "{U}{U}", avB4CrashingWaveOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, foes...)
	for _, id := range foes {
		if got := countersOn(g, id, game.CounterStun); got != 1 {
			t.Errorf("a chosen creature has %d stun counters, want 1", got)
		}
	}
}

func TestAvB4CrashingWaveRefusesMoreTargetsThanX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	tappers := pushTapCostSoldiers(g, me.ID, 1)
	fillPoolColored(me, "U", 2)
	if _, err := castFromHandErr(g, me, "Crashing Wave", "Sorcery", "{U}{U}", avB4CrashingWaveOracle, game.CastSpellParams{
		Strict: true, XValue: 1, TapIDs: tappers,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err == nil {
		t.Fatal("two targets were accepted for X = 1")
	}
}

// --- Ember Island Production -------------------------------------------------------

func avB4TokenNamed(g *game.Game, controller uuid.UUID, name string) (game.Card, bool) {
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Controller == controller && c.Name == name {
				out, found = c, true
			}
		}
	})
	return out, found
}

func TestAvB4EmberIslandCopiesYourCreatureAsANonlegendaryFourFourHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sokka := b12Creature(g, me.ID, "Sokka", "Legendary Creature — Human Warrior Ally", 1, 3)
	castFromHandForTest(t, g, me, "Ember Island Production", "Sorcery", "{3}{U}{U}", avB4EmberIslandOracle, game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sokka}},
	})
	passPriorityAroundTable(t, g)
	tok, ok := avB4TokenNamed(g, me.ID, "Sokka")
	if !ok {
		t.Fatal("no token copy of Sokka")
	}
	eff := tok.Effective()
	if eff.Power != 4 || eff.Toughness != 4 {
		t.Errorf("token is %d/%d, want 4/4", eff.Power, eff.Toughness)
	}
	if tok.HasSupertype("Legendary") {
		t.Error("the token copy is legendary")
	}
	if !tok.HasSubtype("Hero") || !tok.HasSubtype("Warrior") {
		t.Errorf("token type line %q, want Hero in addition to its other types", tok.TypeLine)
	}
}

func TestAvB4EmberIslandCopiesAnOpposingCreatureAsATwoTwoCowardForYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	ozai := b12Creature(g, opp.ID, "Ozai", "Legendary Creature — Human Noble", 7, 7)
	castFromHandForTest(t, g, me, "Ember Island Production", "Sorcery", "{3}{U}{U}", avB4EmberIslandOracle, game.CastSpellParams{
		Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ozai}},
	})
	passPriorityAroundTable(t, g)
	tok, ok := avB4TokenNamed(g, me.ID, "Ozai")
	if !ok {
		t.Fatal("no token copy of the opponent's creature under your control")
	}
	eff := tok.Effective()
	if eff.Power != 2 || eff.Toughness != 2 {
		t.Errorf("token is %d/%d, want 2/2", eff.Power, eff.Toughness)
	}
	if tok.HasSupertype("Legendary") || !tok.HasSubtype("Coward") || !tok.HasSubtype("Noble") {
		t.Errorf("token type line %q, want a nonlegendary Coward in addition", tok.TypeLine)
	}
}

func TestAvB4EmberIslandFirstModeRefusesAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	fillPoolColored(me, "U", 5)
	if _, err := castFromHandErr(g, me, "Ember Island Production", "Sorcery", "{3}{U}{U}", avB4EmberIslandOracle, game.CastSpellParams{
		Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}},
	}); err == nil {
		t.Fatal("the you-control mode accepted an opponent's creature")
	}
}

// --- First-Time Flyer --------------------------------------------------------------

func TestAvB4FirstTimeFlyerGrowsWithALessonInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flyer := b12Push(g, me.ID, "First-Time Flyer", "Creature — Human Pilot Ally", avB4FlyerOracle, 1, 2)
	if p, tough := effectivePTOf(t, g, flyer); p != 1 || tough != 2 {
		t.Fatalf("First-Time Flyer is %d/%d with no Lesson, want 1/2", p, tough)
	}
	avB4Lesson(g, opp)
	seedCreature(g, "Nudge", me.ID) // a battlefield change refreshes the layer cache
	if p, tough := effectivePTOf(t, g, flyer); p != 1 || tough != 2 {
		t.Errorf("an opponent's Lesson pumped First-Time Flyer to %d/%d", p, tough)
	}
	// A Lesson cast and resolved reaches the graveyard the real way.
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "A Lesson", "Sorcery — Lesson", "{U}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, flyer); p != 2 || tough != 3 {
		t.Errorf("First-Time Flyer is %d/%d with a Lesson in your graveyard, want 2/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, flyer, "flying") {
		t.Error("First-Time Flyer has no flying")
	}
}

// --- Flexible Waterbender ----------------------------------------------------------

func TestAvB4FlexibleWaterbenderBecomesFiveTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bender := b12Push(g, me.ID, "Flexible Waterbender", "Creature — Human Warrior Ally", avB4FlexibleOracle, 2, 5)
	tappers := pushTapCostSoldiers(g, me.ID, 3)
	if err := avB4Activate(g, me.ID, bender, 0, tappers); err != nil {
		t.Fatalf("waterbend {3} by tapping three creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bender); p != 5 || tough != 2 {
		t.Errorf("Flexible Waterbender is %d/%d, want 5/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bender, "vigilance") {
		t.Error("Flexible Waterbender has no vigilance")
	}
}

func TestAvB4FlexibleWaterbenderCannotPayWithTwoTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bender := b12Push(g, me.ID, "Flexible Waterbender", "Creature — Human Warrior Ally", avB4FlexibleOracle, 2, 5)
	tappers := pushTapCostSoldiers(g, me.ID, 2)
	if err := avB4Activate(g, me.ID, bender, 0, tappers); err == nil {
		t.Fatal("waterbend {3} was paid with two taps and no mana")
	}
}

// --- Forecasting Fortune Teller ----------------------------------------------------

func TestAvB4FortuneTellerMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Forecasting Fortune Teller", "Creature — Human Advisor Ally", avB4FortuneOracle, 1, 3)
	passPriorityAroundTable(t, g)
	clue, ok := avB4TokenNamed(g, me.ID, "Clue")
	if !ok {
		t.Fatal("no Clue token")
	}
	if !clue.IsArtifact() || !clue.HasSubtype("Clue") {
		t.Errorf("Clue type line %q", clue.TypeLine)
	}
}

// --- Geyser Leaper -----------------------------------------------------------------

func TestAvB4GeyserLeaperLootsForWaterbendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	leaper := b12Push(g, me.ID, "Geyser Leaper", "Creature — Human Warrior Ally", avB4GeyserOracle, 4, 3)
	tappers := pushTapCostSoldiers(g, me.ID, 3)
	fillPool(me, 1)
	hand := me.Hand.Size()
	if err := avB4Activate(g, me.ID, leaper, 0, tappers); err != nil {
		t.Fatalf("waterbend {4} with three taps and one mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the draw")
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after a loot", hand, me.Hand.Size())
	}
	if !hasEffectiveKeyword(t, g, leaper, "flying") {
		t.Error("Geyser Leaper has no flying")
	}
}

// --- Giant Koi ---------------------------------------------------------------------

func TestAvB4GiantKoiCantBeBlockedAfterWaterbending(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	koi := b12Push(g, me.ID, "Giant Koi", "Creature — Fish", avB4KoiOracle, 5, 7)
	if restrictionsOf(t, g, koi)&game.CantBeBlocked != 0 {
		t.Fatal("Giant Koi is unblockable before the ability")
	}
	tappers := pushTapCostSoldiers(g, me.ID, 3)
	if err := avB4Activate(g, me.ID, koi, 0, tappers); err != nil {
		t.Fatalf("waterbend {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, koi)&game.CantBeBlocked == 0 {
		t.Error("Giant Koi can still be blocked")
	}
}

func TestAvB4GiantKoiIslandcyclesForAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	koi := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: koi, Name: "Giant Koi", TypeLine: "Creature — Fish", OracleID: avB4KoiOracle,
			ManaCost: "{4}{U}{U}", Power: 5, Toughness: 7, Owner: me.ID, Controller: me.ID})
	})
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: avB4KoiOracle})
	if len(abilities) != 2 {
		t.Fatalf("%d activated abilities, want waterbend and islandcycling", len(abilities))
	}
	if abilities[1].Label == "" || abilities[1].Cost.Mana != "{2}" {
		t.Errorf("islandcycling ability %+v, want a {2} cost", abilities[1])
	}
}

// --- Gran-Gran ---------------------------------------------------------------------

func TestAvB4GranGranLootsWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gran := b12Push(g, me.ID, "Gran-Gran", "Legendary Creature — Human Peasant Ally", avB4GranGranOracle, 1, 2)
	hand := me.Hand.Size()
	trekB18AAttack(t, g, opp.ID, gran)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no loot after Gran-Gran became tapped")
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after a loot", hand, me.Hand.Size())
	}
}

func TestAvB4GranGranDiscountsNoncreatureSpellsWithThreeLessons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Gran-Gran", "Legendary Creature — Human Peasant Ally", avB4GranGranOracle, 1, 2)
	avB4Lesson(g, me)
	avB4Lesson(g, me)
	fillPoolColored(me, "U", 1)
	if _, err := castFromHandErr(g, me, "Two Drop", "Sorcery", "{1}{U}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a {1}{U} sorcery was cast for {U} with only two Lessons in the graveyard")
	}
	avB4Lesson(g, me)
	castFromHandForTest(t, g, me, "Two Drop", "Sorcery", "{1}{U}", "", game.CastSpellParams{Strict: true})
}

func TestAvB4GranGranDoesNotDiscountCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Gran-Gran", "Legendary Creature — Human Peasant Ally", avB4GranGranOracle, 1, 2)
	for i := 0; i < 3; i++ {
		avB4Lesson(g, me)
	}
	fillPoolColored(me, "U", 1)
	if _, err := castFromHandErr(g, me, "Two Drop Bear", "Creature — Bear", "{1}{U}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was discounted")
	}
}

// --- Honest Work -------------------------------------------------------------------

func TestAvB4HonestWorkMakesAHumbleMerchant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Dragon", TypeLine: "Creature — Dragon",
		Keywords: []string{"flying"}, Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	auraCast(t, g, "Honest Work", avB4HonestWorkOracle, dragon)
	passPriorityAroundTable(t, g)

	c := effectiveCardForTest(t, g, dragon)
	if !c.Tapped {
		t.Error("the enchanted creature was not tapped")
	}
	if n := countersOn(g, dragon, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters left, want all removed", n)
	}
	eff := c.Effective()
	if eff.Name != "Humble Merchant" {
		t.Errorf("name %q, want Humble Merchant", eff.Name)
	}
	if eff.Power != 1 || eff.Toughness != 1 {
		t.Errorf("P/T %d/%d, want 1/1", eff.Power, eff.Toughness)
	}
	if subs := effectiveSubtypes(t, g, dragon); len(subs) != 1 || subs[0] != "Citizen" {
		t.Errorf("subtypes %v, want only Citizen", subs)
	}
	if !c.IsCreature() {
		t.Error("the enchanted creature stopped being a creature")
	}
	if containsString(effectiveAbilities(t, g, dragon), "flying") {
		t.Error("the enchanted creature kept flying")
	}
	if n := grantedManaRowsOf(g, dragon); n != 1 {
		t.Fatalf("%d granted mana abilities, want {T}: Add {C}", n)
	}
	_ = me
}

func TestAvB4HonestWorkRefusesYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "U", 1)
	if _, err := castFromHandErr(g, me, "Honest Work", "Enchantment — Aura", "{U}", avB4HonestWorkOracle, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("Honest Work enchanted a creature you control")
	}
}
