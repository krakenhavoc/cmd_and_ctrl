package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_st_cards_test.go — Bloomburrow cards starting with S or T, second
// batch (#4467). Each test plays the real catalog entry and asserts what
// its text says.

const (
	blbSTScalesOracle      = "655a2b78-6a6c-4b9e-b7c0-48d3768978f5"
	blbSTRedeemerOracle    = "39bb051f-1ca7-4d9b-b33b-7f559f816e3e"
	blbSTMonolithOracle    = "ae8003eb-9f19-4f23-be3b-0ac3ddc03a72"
	blbSTSkyskipperOracle  = "a50db36c-948b-4132-a5e5-84aab1f42612"
	blbSTSpellgyreOracle   = "7db31623-5dea-4165-8593-fbac568bf5b4"
	blbSTLasherOracle      = "0d0c4c91-683a-4563-a0b0-39f3c04f6307"
	blbSTSoothsayerOracle  = "0c0940c6-8126-4d58-809e-871cf8588ef9"
	blbSTClericOracle      = "8d72ddb3-ef6e-4847-b6d4-5d85b1dd5802"
	blbSTStarseerOracle    = "12a165c4-229b-458b-8fc7-c1b7cba35a8e"
	blbSTChargerOracle     = "75237b2c-232b-479a-ab3a-e1566c8a9fe7"
	blbSTSlingerOracle     = "5ba19dbe-9dbb-4cc9-81a3-18c26f27bd0f"
	blbSTWildguideOracle   = "74868bc0-6081-4f46-a2c6-9521cd5465aa"
	blbSTThistledownOracle = "7054c42b-48dd-4887-b878-69c005259d27"
	blbSTThornplateOracle  = "33b6555e-035d-4241-a341-a9b9139ca6f0"
	blbSTTreeguardOracle   = "d7dfb4d5-8510-4904-abb7-834ec65d9bf3"
)

func TestBlbSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbSTScalesOracle:      "Scales of Shale",
		blbSTRedeemerOracle:    "Serra Redeemer",
		blbSTMonolithOracle:    "Sinister Monolith",
		blbSTSkyskipperOracle:  "Skyskipper Duo",
		blbSTSpellgyreOracle:   "Spellgyre",
		blbSTLasherOracle:      "Splash Lasher",
		blbSTSoothsayerOracle:  "Starlit Soothsayer",
		blbSTClericOracle:      "Starscape Cleric",
		blbSTStarseerOracle:    "Starseer Mentor",
		blbSTChargerOracle:     "Steampath Charger",
		blbSTSlingerOracle:     "Teapot Slinger",
		blbSTWildguideOracle:   "Tender Wildguide",
		blbSTThistledownOracle: "Thistledown Players",
		blbSTThornplateOracle:  "Thornplate Intimidator",
		blbSTTreeguardOracle:   "Treeguard Duo",
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

// blbSTOpponentLives is every opponent's life total, by player.
func blbSTOpponentLives(g *game.Game, me uuid.UUID) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me {
			out[p.ID] = p.Life
		}
	}
	return out
}

// blbSTEachOpponentLost fails unless every opponent lost exactly n life
// since `before`.
func blbSTEachOpponentLost(t *testing.T, g *game.Game, before map[uuid.UUID]int, n int) {
	t.Helper()
	for id, life := range before {
		if got := lifeOf(g, id); got != life-n {
			t.Errorf("opponent %v is at %d, want %d", id, got, life-n)
		}
	}
	if len(before) != len(g.Seats)-1 {
		t.Fatalf("%d opponents, want %d", len(before), len(g.Seats)-1)
	}
}

// blbSTPickOption answers the opponent's punisher question with the
// option labelled `label`.
func blbSTPickOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	pick := latestOptionPickFor(g, chooser)
	if pick == nil {
		t.Fatalf("%v is not asked: %+v", chooser, g.PendingChoices)
	}
	for i, o := range pick.PickOptions {
		if o.Label == label {
			answerOptionPick(t, g, chooser, i)
			return
		}
	}
	t.Fatalf("no option %q among %+v", label, pick.PickOptions)
}

// --- Scales of Shale ---------------------------------------------------------

func TestBlbSTScalesOfShalePumpsAndGrantsLifelinkAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Scales of Shale", "Instant", blbSTScalesOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 4/2", p, tough)
	}
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the bear should have %s", kw)
		}
	}
}

func TestBlbSTScalesOfShaleCostsOneLessPerLizard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Scales of Shale", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: blbSTScalesOracle}
	if err := socB19CastWithPool(g, me, card, "B", 2, game.CastSpellParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("cast for two mana with no Lizard")
	}
	blbB1Push(g, me.ID, "Lizard", "Creature — Lizard Warlock", "", 1, 1)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "B", 2, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast for {1}{B} with a Lizard: %v", err)
	}
}

// --- Serra Redeemer ----------------------------------------------------------

func TestBlbSTSerraRedeemerGrowsOnlySmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	redeemer := blbB1Push(g, me.ID, "Serra Redeemer", "Creature — Angel Soldier", blbSTRedeemerOracle, 2, 4)
	if !hasKeywordOnBattlefield(t, g, redeemer, "flying") {
		t.Error("the Redeemer should fly")
	}
	small := trekB31Enter(t, g, me, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, small); got != 2 {
		t.Errorf("the 2/2 has %d +1/+1 counters, want 2", got)
	}
	big := trekB31Enter(t, g, me, game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, big); got != 0 {
		t.Errorf("the 3/3 has %d +1/+1 counters, want 0", got)
	}
	if got := plusOnes(g, redeemer); got != 0 {
		t.Errorf("the Redeemer has %d +1/+1 counters, want 0", got)
	}
}

// --- Sinister Monolith -------------------------------------------------------

func TestBlbSTSinisterMonolithDrainsAtYourBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Sinister Monolith", "Artifact", blbSTMonolithOracle, false)
	advanceToMain(t, g)
	before, life := blbSTOpponentLives(g, me.ID), lifeOf(g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	blbSTEachOpponentLost(t, g, before, 1)
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("I am at %d, want %d", got, life+1)
	}
}

func TestBlbSTSinisterMonolithSacrificesToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	monolith := pushCatalogPermanent(g, me.ID, "Sinister Monolith", "Artifact", blbSTMonolithOracle, false)
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), lifeOf(g, me.ID)
	b16Activate(t, g, me.ID, monolith, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if got := lifeOf(g, me.ID); got != life-2 {
		t.Errorf("I am at %d, want %d", got, life-2)
	}
	if g.Battlefield.Contains(monolith) {
		t.Error("the Monolith should be sacrificed")
	}
}

// --- Skyskipper Duo ----------------------------------------------------------

func TestBlbSTSkyskipperDuoBlinksAnotherCreatureUntilTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	duo := castHoldCreature(t, g, "Skyskipper Duo", "Creature — Bird Frog", blbSTSkyskipperOracle, 3, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: duo}); err == nil {
		t.Fatal("the Duo itself was accepted; the target is another creature")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || b12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatal("the bear was not exiled")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Bear") {
		t.Error("the bear did not return at the end step")
	}
}

// --- Spellgyre ---------------------------------------------------------------

func TestBlbSTSpellgyreCountersASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	gyre := b13HandCard(me, "Spellgyre", "Instant", blbSTSpellgyreOracle)
	life := me.Life
	if err := g.CastSpell(me.ID, gyre, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("Spellgyre: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt should be countered")
	}
}

func TestBlbSTSpellgyreSurveilsTwoThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	topLibraryCard(me, "Fourth", "Sorcery")
	third := topLibraryCard(me, "Third", "Sorcery")
	second := topLibraryCard(me, "Second", "Sorcery")
	top := topLibraryCard(me, "Top", "Sorcery")
	hand := me.Hand.Size()
	castModal(t, g, "Spellgyre", "Instant", blbSTSpellgyreOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+2 || !me.Hand.Contains(second) || !me.Hand.Contains(third) {
		t.Errorf("hand %d (want %d) should hold the kept card and the next one", me.Hand.Size(), hand+2)
	}
}

// --- Splash Lasher -----------------------------------------------------------

func TestBlbSTSplashLasherTapsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castHoldCreature(t, g, "Splash Lasher", "Creature — Frog Wizard", blbSTLasherOracle, 3, 3)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the target should be tapped")
	}
	if got := countersOn(g, theirs, game.CounterStun); got != 1 {
		t.Errorf("the target has %d stun counters, want 1", got)
	}
}

// --- Starlit Soothsayer ------------------------------------------------------

func TestBlbSTStarlitSoothsayerSurveilsOnlyAfterALifeChange(t *testing.T) {
	for _, tc := range []struct {
		name  string
		delta int
		want  bool
	}{
		{"no change", 0, false},
		{"gained", 2, true},
		{"lost", -2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _, _ := blbB1Seats(g)
			seedLibrary(me, "A", "B")
			blbB1Push(g, me.ID, "Starlit Soothsayer", "Creature — Bat Cleric", blbSTSoothsayerOracle, 2, 2)
			advanceToMain(t, g)
			if tc.delta != 0 {
				b14Gain(t, g, me.ID, tc.delta)
			}
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			c := surveilChoiceFor(g, me.ID)
			if got := c != nil; got != tc.want {
				t.Fatalf("surveil asked = %v, want %v", got, tc.want)
			}
			if c != nil && len(c.ScryCards) != 1 {
				t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
			}
		})
	}
}

// --- Starscape Cleric --------------------------------------------------------

func TestBlbSTStarscapeClericCantBlockAndDrainsOnLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	cleric := blbB1Push(g, me.ID, "Starscape Cleric", "Creature — Bat Cleric", blbSTClericOracle, 2, 1)
	g.ReadSnapshot(func() { g.RecomputeLayersIfStaleLocked() })
	if c, ok := battlefieldCard(g, cleric); !ok || !game.Restricted(&c, game.CantBlock) {
		t.Error("the Cleric can block")
	}
	if !hasKeywordOnBattlefield(t, g, cleric, "flying") {
		t.Error("the Cleric should fly")
	}
	before := blbSTOpponentLives(g, me.ID)
	b14Gain(t, g, me.ID, 3)
	passPriorityAroundTable(t, g)
	blbSTEachOpponentLost(t, g, before, 1)
}

// --- Starseer Mentor ---------------------------------------------------------

func TestBlbSTStarseerMentorPunishesAfterALifeChange(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Starseer Mentor", "Creature — Bat Warlock", blbSTStarseerOracle, 3, 5)
	advanceToMain(t, g)
	b14Gain(t, g, me.ID, 1)
	advanceToEndStep(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("I was accepted; the target is an opponent")
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	life := lifeOf(g, opp.ID)
	blbSTPickOption(t, g, opp.ID, "Lose 3 life")
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("the opponent is at %d, want %d", got, life-3)
	}
}

func TestBlbSTStarseerMentorDoesNothingWithoutALifeChange(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Starseer Mentor", "Creature — Bat Warlock", blbSTStarseerOracle, 3, 5)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		t.Error("the Mentor triggered with no life gained or lost")
	}
}

// --- Steampath Charger -------------------------------------------------------

func TestBlbSTSteampathChargerOffspringMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := castWithOptionalCosts(t, g, "Steampath Charger", "Creature — Lizard Warlock", blbSTChargerOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	passPriorityAroundTable(t, g)
	chargers := blbMP2Named(g, "Steampath Charger")
	if len(chargers) != 2 {
		t.Fatalf("%d Steampath Chargers, want 2 (offspring paid)", len(chargers))
	}
	tokens := 0
	for _, c := range chargers {
		if c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d tokens, want 1", tokens)
	}
}

func TestBlbSTSteampathChargerPingsAPlayerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	charger := blbB1Push(g, me.ID, "Steampath Charger", "Creature — Lizard Warlock", blbSTChargerOracle, 2, 1)
	life := lifeOf(g, opp.ID)
	killCreature(t, g, me.ID, charger)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-1 {
		t.Errorf("the opponent is at %d, want %d", got, life-1)
	}
}

// --- Teapot Slinger ----------------------------------------------------------

func TestBlbSTTeapotSlingerDealsTwoToEachOpponentOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	slinger := blbB1Push(g, me.ID, "Teapot Slinger", "Creature — Raccoon Warrior", blbSTSlingerOracle, 3, 4)
	if !hasKeywordOnBattlefield(t, g, slinger, "menace") {
		t.Error("the Slinger should have menace")
	}
	toMainForCost(t, g)
	before := blbSTOpponentLives(g, me.ID)
	castPaid(t, g, me, "Three", "{2}{R}", "CCR")
	passPriorityAroundTable(t, g)
	blbSTEachOpponentLost(t, g, before, 0)
	castPaid(t, g, me, "One", "{R}", "R")
	passPriorityAroundTable(t, g)
	blbSTEachOpponentLost(t, g, before, 2)
}

// --- Tender Wildguide --------------------------------------------------------

func TestBlbSTTenderWildguideTapsForMana(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _, _ := blbB1Seats(g)
	mana := pushCatalogPermanent(g, me.ID, "Tender Wildguide", "Creature — Possum Druid", blbSTWildguideOracle, false)
	if err := g.ActivateManaAbility(me.ID, mana, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !b16Tapped(t, g, mana) {
		t.Error("the Wildguide should be tapped for mana")
	}
}

func TestBlbSTTenderWildguideTapsForACounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _, _ := blbB1Seats(g)
	grow := blbB1Push(g, me.ID, "Tender Wildguide", "Creature — Possum Druid", blbSTWildguideOracle, 2, 2)
	b16Activate(t, g, me.ID, grow, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, grow); got != 1 {
		t.Errorf("the Wildguide has %d +1/+1 counters, want 1", got)
	}
	sick := pushCatalogPermanent(g, me.ID, "Tender Wildguide", "Creature — Possum Druid", blbSTWildguideOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Wildguide tapped for a counter")
	}
}

// --- Thistledown Players -----------------------------------------------------

func TestBlbSTThistledownPlayersUntapsANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	players := blbB1Push(g, me.ID, "Thistledown Players", "Creature — Mouse Bard", blbSTThistledownOracle, 3, 3)
	tapped := blbB1Push(g, me.ID, "Tapped Bear", "Creature — Bear", "", 2, 2)
	land := b12Permanent(g, me.ID, "Tapped Forest", "Basic Land — Forest")
	b16Tap(g, tapped)
	b16Tap(g, land)
	declareAttack(t, g, opp.ID, players)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was accepted; the target is a nonland permanent")
	}
	answerPickTarget(t, g, tapped)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, tapped) {
		t.Error("the bear should be untapped")
	}
}

// --- Thornplate Intimidator --------------------------------------------------

func TestBlbSTThornplateIntimidatorPunishesTheTargetedOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castHoldCreature(t, g, "Thornplate Intimidator", "Creature — Rat Rogue", blbSTThornplateOracle, 4, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("I was accepted; the target is an opponent")
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, other.ID) != nil {
		t.Fatal("an opponent who was not targeted is asked")
	}
	life := lifeOf(g, opp.ID)
	blbSTPickOption(t, g, opp.ID, punisherSacrificeLabel)
	sac := latestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatal("the sacrifice branch asks nothing")
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the bear was not sacrificed")
	}
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("the opponent sacrificed and is at %d, want %d", got, life)
	}
}

// --- Treeguard Duo -----------------------------------------------------------

func TestBlbSTTreeguardDuoPumpsByTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, me.ID, "Other", "Creature — Rabbit", "", 1, 1)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castHoldCreature(t, g, "Treeguard Duo", "Creature — Frog Rabbit", blbSTTreeguardOracle, 3, 4)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	// The bear, the other creature and the Duo: X is 3.
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the bear is %d/%d, want 5/5", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("the bear should have vigilance")
	}
}
