package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// populate_pool_test.go — the populate cards #2318 left out: one
// test per card for its main behaviour, plus the refused / empty
// cases that make it the printed card rather than a stronger one.

const (
	eyesInTheSkiesOracle     = "9697cd05-474e-46f3-8bcc-d4b6fb2059a1"
	horncallersChantOracle   = "8c069834-ed89-4298-989b-e67036d56196"
	coursersAccordOracle     = "23674648-d1a0-437d-8a0b-4173c75b4507"
	trostanisJudgmentOracle  = "97cf544e-ffbf-4730-8cd2-4f1d5be933f2"
	nestingDovehawkOracle    = "fe8fc442-ed17-40b2-8624-69f2eed3f9be"
	scionOfVituGhaziOracle   = "a34c4d5c-6ecb-4da2-8734-a2f3d35cfe8b"
	trostaniOracle           = "e94ef397-f5c5-4b8d-ae27-528352fa1d1e"
	wayfaringTempleOracle    = "c0525645-9b5a-488d-b4d1-e80f16a1e4e1"
	selesnyaEulogistOracle   = "a0335224-9523-47d5-8944-486f2d4a6b8f"
	lifeFindsAWayOracle      = "fa4f22b2-7fe4-4f7f-8afd-32add76d740d"
	songOfTheWorldsoulOracle = "fbb70f54-8c33-4de0-b05c-f0bff9ffa6ce"
	fullFloweringOracle      = "cb3328b5-a759-4c30-8b1f-202980bb6f6d"
)

func TestPopulatePoolIsRegisteredAsFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eyesInTheSkiesOracle:     "Eyes in the Skies",
		horncallersChantOracle:   "Horncaller's Chant",
		coursersAccordOracle:     "Coursers' Accord",
		trostanisJudgmentOracle:  "Trostani's Judgment",
		nestingDovehawkOracle:    "Nesting Dovehawk",
		scionOfVituGhaziOracle:   "Scion of Vitu-Ghazi",
		trostaniOracle:           "Trostani, Selesnya's Voice",
		wayfaringTempleOracle:    "Wayfaring Temple",
		selesnyaEulogistOracle:   "Selesnya Eulogist",
		lifeFindsAWayOracle:      "Life Finds a Way",
		songOfTheWorldsoulOracle: "Song of the Worldsoul",
		fullFloweringOracle:      "Full Flowering",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v as %q", name, ok, spec.Name)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: Completeness = %s, want full", name, spec.Completeness)
		}
	}
}

func TestEyesInTheSkiesMakesABirdThenCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Eyes in the Skies", "Instant", eyesInTheSkiesOracle, nil)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Bird")); got != 2 {
		t.Fatalf("Birds = %d, want 2 (the new one and its copy)", got)
	}
}

func TestEyesInTheSkiesCanCopyAnotherTokenInstead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castCatalogSpell(t, g, "Eyes in the Skies", "Instant", eyesInTheSkiesOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("a Bird and a Centaur are a real choice")
	}
	answerChooseCards(t, g, me.ID, centaur)

	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Errorf("Centaurs = %d, want 2", got)
	}
	if got := len(populateTokens(g, me.ID, "Bird")); got != 1 {
		t.Errorf("Birds = %d, want 1", got)
	}
}

func TestHorncallersChantMakesATramplingRhinoThenCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Horncaller's Chant", "Sorcery", horncallersChantOracle, nil)
	passPriorityAroundTable(t, g)

	rhinos := populateTokens(g, me.ID, "Rhino")
	if len(rhinos) != 2 {
		t.Fatalf("Rhinos = %d, want 2", len(rhinos))
	}
	for _, id := range rhinos {
		c, _ := g.LookupCardForEffect(id)
		if c.Power != 4 || c.Toughness != 4 || !hasKeywordForTest(c.Keywords, "trample") {
			t.Errorf("Rhino = %d/%d %v, want a 4/4 with trample", c.Power, c.Toughness, c.Keywords)
		}
	}
}

func TestCoursersAccordMakesACentaurThenCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Coursers' Accord", "Sorcery", coursersAccordOracle, nil)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestTrostanisJudgmentExilesThenPopulates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	victim := pushVanillaCreature(g, opp.ID, "Big Bear", 4, 4)

	castCatalogSpell(t, g, "Trostani's Judgment", "Instant", trostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the target should be exiled")
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestTrostanisJudgmentOnYourOnlyTokenLeavesNothingToCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	soldier := pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	castCatalogSpell(t, g, "Trostani's Judgment", "Instant", trostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: soldier}})
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 0 {
		t.Fatalf("Soldiers = %d, want 0: the populate happens after the exile (CR 701.36b)", got)
	}
}

func TestTrostanisJudgmentNeedsACreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushPermanentForTest(g, opp.ID, "Sol Ring", "", "Artifact")

	err := castCatalogSpellErr(t, g, "Trostani's Judgment", "Instant", trostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if err == nil {
		t.Fatal("an artifact is not a legal target")
	}
}

func TestNestingDovehawkPopulatesAtCombatAndGrowsFromTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hawk := pushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", nestingDovehawkOracle, false)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
	if got := currentPowerOf(g, hawk); got != 2 {
		t.Errorf("Dovehawk power = %d, want 2 (the fixture body is 1/1, plus one counter)", got)
	}
}

func TestNestingDovehawkIgnoresNontokensAndOpponentsTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := pushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", nestingDovehawkOracle, false)
	before := currentPowerOf(g, hawk)

	b20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1) })
	passPriorityAroundTable(t, g)
	if got := currentPowerOf(g, hawk); got != before {
		t.Fatalf("power %d -> %d, want no counter", before, got)
	}
}

func TestNestingDovehawkGrowsForEachCreatureTokenYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hawk := pushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", nestingDovehawkOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1) })
	passPriorityAroundTable(t, g)
	if got := currentPowerOf(g, hawk); got != 2 {
		t.Errorf("power = %d, want 2 (the 1/1 fixture body plus one counter)", got)
	}
}

func TestScionOfVituGhaziMakesABirdAndPopulatesOnlyWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	b20CastCreature(t, g, me, "Scion of Vitu-Ghazi", "Creature — Elemental", scionOfVituGhaziOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Bird")); got != 2 {
		t.Fatalf("cast from hand: Birds = %d, want 2", got)
	}

	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Scion of Vitu-Ghazi", TypeLine: "Creature — Elemental",
		OracleID: scionOfVituGhaziOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	b30Reanimate(t, g, dead, me.ID)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Bird")); got != 2 {
		t.Errorf("reanimated: Birds = %d, want still 2", got)
	}
}

func TestTrostaniGainsLifeForOtherCreaturesAndPopulates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trostani := pushCatalogPermanent(g, me.ID, "Trostani, Selesnya's Voice",
		"Legendary Creature — Dryad", trostaniOracle, false)
	life := me.Life

	b20CastCreature(t, g, me, "Wall", "Creature — Wall", "", 0, 5)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Fatalf("life %d -> %d, want +5 (the entering creature's toughness)", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Centaur"), 1) })
	passPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Fatalf("life %d, want %d after a 3/3 token", me.Life, life+8)
	}

	fillPool(me, 1)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, trostani, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
	// The copy entered too, so Trostani gained 3 more.
	if me.Life != life+11 {
		t.Errorf("life %d, want %d (the copy's entry also gains life)", me.Life, life+11)
	}
}

func TestTrostaniIgnoresHerOwnEntryAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life

	b20CastCreature(t, g, me, "Trostani, Selesnya's Voice", "Legendary Creature — Dryad", trostaniOracle, 2, 5)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d -> %d, want no gain for her own entry", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("3/3 green Centaur"), 1) })
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d, want no gain for an opponent's creature", life, me.Life)
	}
}

func TestWayfaringTempleSizeIsYourCreatureCountAndConnectingPopulates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := pushCatalogPermanent(g, me.ID, "Wayfaring Temple", "Creature — Elemental", wayfaringTempleOracle, false)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1) })

	if p, tt := effectivePower(t, g, temple), effectiveToughness(t, g, temple); p != 4 || tt != 4 {
		t.Fatalf("Temple = %d/%d, want 4/4 (Temple, two Soldiers, a Bear)", p, tt)
	}

	before := len(populateTokens(g, me.ID, "Soldier"))
	attackWith(t, g, opp.ID, temple)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Soldier")); got != before+1 {
		t.Errorf("Soldiers = %d, want %d after the Temple connected", got, before+1)
	}
}

func TestWayfaringTempleDoesNotPopulateWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := pushCatalogPermanent(g, me.ID, "Wayfaring Temple", "Creature — Elemental", wayfaringTempleOracle, false)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 9)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(temple, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, temple); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want 1: damage to a creature does not trigger it", got)
	}
}

func TestSelesnyaEulogistExilesACreatureCardThenPopulates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eulogist := pushCatalogPermanent(g, me.ID, "Selesnya Eulogist", "Creature — Centaur Druid", selesnyaEulogistOracle, true)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})

	fillPool(me, 2)
	fillPoolColored(me, "G", 1)
	// A mana-only cost: summoning sickness does not matter.
	if err := g.ActivateCatalogAbility(me.ID, eulogist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(dead) {
		t.Error("the creature card should be exiled")
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestSelesnyaEulogistRefusesANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eulogist := pushCatalogPermanent(g, me.ID, "Selesnya Eulogist", "Creature — Centaur Druid", selesnyaEulogistOracle, false)
	rock := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})

	fillPool(me, 2)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, eulogist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("an artifact card is not a legal target")
	}
}

func TestLifeFindsAWayPopulatesForBigNontokenCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Life Finds a Way", lifeFindsAWayOracle, "Enchantment")
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	b20CastCreature(t, g, me, "Small Bear", "Creature — Bear", "", 3, 3)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Fatalf("a 3-power creature: Soldiers = %d, want 1", got)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("4/4 green Rhino with trample"), 1) })
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Rhino")); got != 1 {
		t.Fatalf("a 4/4 TOKEN must not trigger it: Rhinos = %d, want 1", got)
	}

	b20CastCreature(t, g, me, "Big Bear", "Creature — Bear", "", 4, 4)
	passPriorityAroundTable(t, g)
	// Soldier and Rhino differ, so the populate asks.
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("a 4-power nontoken creature should populate")
	}
	answerChooseCards(t, g, me.ID, populateTokens(g, me.ID, "Soldier")[0])
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Errorf("Soldiers = %d, want 2", got)
	}
}

func TestSongOfTheWorldsoulPopulatesOnEveryCastAndTheSpellStillResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Song of the Worldsoul", songOfTheWorldsoulOracle, "Enchantment")
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	life := opp.Life

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want the Bolt to still resolve", life, opp.Life)
	}
}

func TestSongOfTheWorldsoulIgnoresOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, opp.ID, "Song of the Worldsoul", songOfTheWorldsoulOracle, "Enchantment")
	pushToken(g, opp.ID, TokenCard("1/1 white Soldier"))

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, opp.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want 1: only your own casts trigger it", got)
	}
}

func TestFullFloweringPopulatesXTimes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	castXSpell(t, g, "Full Flowering", "Sorcery", fullFloweringOracle, "{X}{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 4 {
		t.Fatalf("Soldiers = %d, want 4 (one original, three copies)", got)
	}
}

func TestFullFloweringWithXZeroOrNoTokenDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	castXSpell(t, g, "Full Flowering", "Sorcery", fullFloweringOracle, "{X}{X}{G}", 0, nil)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Fatalf("X=0: Soldiers = %d, want 1", got)
	}

	g2 := newCatalogGame(t)
	before := len(g2.Battlefield.Cards)
	castXSpell(t, g2, "Full Flowering", "Sorcery", fullFloweringOracle, "{X}{X}{G}", 2, nil)
	passPriorityAroundTable(t, g2)
	if got := len(g2.Battlefield.Cards); got != before {
		t.Errorf("no token: battlefield %d -> %d, want no change", before, got)
	}
}

func TestFullFloweringAsksEachTimeAndCanCopyTheNewCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castXSpell(t, g, "Full Flowering", "Sorcery", fullFloweringOracle, "{X}{X}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)
	// Second populate: Soldier / Centaur are different again.
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the second populate should ask too")
	}
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Centaur")); got != 3 {
		t.Errorf("Centaurs = %d, want 3", got)
	}
}

// currentPowerOf is power including +1/+1 counters, which the layer
// engine's Effective() deliberately leaves out.
func currentPowerOf(g *game.Game, id uuid.UUID) int {
	c, _ := g.LookupCardForEffect(id)
	return c.CurrentPower()
}
