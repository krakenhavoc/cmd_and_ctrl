package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPopulatePoolIsRegisteredAsFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.EyesInTheSkiesOracle:     "Eyes in the Skies",
		effectstest.HorncallersChantOracle:   "Horncaller's Chant",
		effectstest.CoursersAccordOracle:     "Coursers' Accord",
		effectstest.TrostanisJudgmentOracle:  "Trostani's Judgment",
		effectstest.NestingDovehawkOracle:    "Nesting Dovehawk",
		effectstest.ScionOfVituGhaziOracle:   "Scion of Vitu-Ghazi",
		effectstest.TrostaniOracle:           "Trostani, Selesnya's Voice",
		effectstest.WayfaringTempleOracle:    "Wayfaring Temple",
		effectstest.SelesnyaEulogistOracle:   "Selesnya Eulogist",
		effectstest.LifeFindsAWayOracle:      "Life Finds a Way",
		effectstest.SongOfTheWorldsoulOracle: "Song of the Worldsoul",
		effectstest.FullFloweringOracle:      "Full Flowering",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v as %q", name, ok, spec.Name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: Completeness = %s, want full", name, spec.Completeness)
		}
	}
}

func TestEyesInTheSkiesMakesABirdThenCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Eyes in the Skies", "Instant", effectstest.EyesInTheSkiesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Bird")); got != 2 {
		t.Fatalf("Birds = %d, want 2 (the new one and its copy)", got)
	}
}

func TestEyesInTheSkiesCanCopyAnotherTokenInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	centaur := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	effectstest.CastCatalogSpell(t, g, "Eyes in the Skies", "Instant", effectstest.EyesInTheSkiesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("a Bird and a Centaur are a real choice")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, centaur)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Errorf("Centaurs = %d, want 2", got)
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Bird")); got != 1 {
		t.Errorf("Birds = %d, want 1", got)
	}
}

func TestHorncallersChantMakesATramplingRhinoThenCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Horncaller's Chant", "Sorcery", effectstest.HorncallersChantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	rhinos := effectstest.PopulateTokens(g, me.ID, "Rhino")
	if len(rhinos) != 2 {
		t.Fatalf("Rhinos = %d, want 2", len(rhinos))
	}
	for _, id := range rhinos {
		c, _ := g.LookupCardForEffect(id)
		if c.Power != 4 || c.Toughness != 4 || !effectstest.HasKeywordForTest(c.Keywords, "trample") {
			t.Errorf("Rhino = %d/%d %v, want a 4/4 with trample", c.Power, c.Toughness, c.Keywords)
		}
	}
}

func TestCoursersAccordMakesACentaurThenCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Coursers' Accord", "Sorcery", effectstest.CoursersAccordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestTrostanisJudgmentExilesThenPopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Big Bear", 4, 4)

	effectstest.CastCatalogSpell(t, g, "Trostani's Judgment", "Instant", effectstest.TrostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the target should be exiled")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestTrostanisJudgmentOnYourOnlyTokenLeavesNothingToCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soldier := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.CastCatalogSpell(t, g, "Trostani's Judgment", "Instant", effectstest.TrostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: soldier}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 0 {
		t.Fatalf("Soldiers = %d, want 0: the populate happens after the exile (CR 701.36b)", got)
	}
}

func TestTrostanisJudgmentNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.PushPermanentForTest(g, opp.ID, "Sol Ring", "", "Artifact")

	err := effectstest.CastCatalogSpellErr(t, g, "Trostani's Judgment", "Instant", effectstest.TrostanisJudgmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if err == nil {
		t.Fatal("an artifact is not a legal target")
	}
}

func TestNestingDovehawkPopulatesAtCombatAndGrowsFromTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.PushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", effectstest.NestingDovehawkOracle, false)
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
	if got := currentPowerOf(g, hawk); got != 2 {
		t.Errorf("Dovehawk power = %d, want 2 (the fixture body is 1/1, plus one counter)", got)
	}
}

func TestNestingDovehawkIgnoresNontokensAndOpponentsTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := effectstest.PushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", effectstest.NestingDovehawkOracle, false)
	before := currentPowerOf(g, hawk)

	effectstest.B20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := currentPowerOf(g, hawk); got != before {
		t.Fatalf("power %d -> %d, want no counter", before, got)
	}
}

func TestNestingDovehawkGrowsForEachCreatureTokenYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.PushCatalogPermanent(g, me.ID, "Nesting Dovehawk", "Creature — Bird", effectstest.NestingDovehawkOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := currentPowerOf(g, hawk); got != 2 {
		t.Errorf("power = %d, want 2 (the 1/1 fixture body plus one counter)", got)
	}
}

func TestScionOfVituGhaziMakesABirdAndPopulatesOnlyWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.B20CastCreature(t, g, me, "Scion of Vitu-Ghazi", "Creature — Elemental", effectstest.ScionOfVituGhaziOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Bird")); got != 2 {
		t.Fatalf("cast from hand: Birds = %d, want 2", got)
	}

	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Scion of Vitu-Ghazi", TypeLine: "Creature — Elemental",
		OracleID: effectstest.ScionOfVituGhaziOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Bird")); got != 2 {
		t.Errorf("reanimated: Birds = %d, want still 2", got)
	}
}

func TestTrostaniGainsLifeForOtherCreaturesAndPopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trostani := effectstest.PushCatalogPermanent(g, me.ID, "Trostani, Selesnya's Voice",
		"Legendary Creature — Dryad", effectstest.TrostaniOracle, false)
	life := me.Life

	effectstest.B20CastCreature(t, g, me, "Wall", "Creature — Wall", "", 0, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Fatalf("life %d -> %d, want +5 (the entering creature's toughness)", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Centaur"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Fatalf("life %d, want %d after a 3/3 token", me.Life, life+8)
	}

	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, trostani, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
	// The copy entered too, so Trostani gained 3 more.
	if me.Life != life+11 {
		t.Errorf("life %d, want %d (the copy's entry also gains life)", me.Life, life+11)
	}
}

func TestTrostaniIgnoresHerOwnEntryAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life

	effectstest.B20CastCreature(t, g, me, "Trostani, Selesnya's Voice", "Legendary Creature — Dryad", effectstest.TrostaniOracle, 2, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d -> %d, want no gain for her own entry", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("3/3 green Centaur"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d, want no gain for an opponent's creature", life, me.Life)
	}
}

func TestWayfaringTempleSizeIsYourCreatureCountAndConnectingPopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := effectstest.PushCatalogPermanent(g, me.ID, "Wayfaring Temple", "Creature — Elemental", effectstest.WayfaringTempleOracle, false)
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })

	if p, tt := effectstest.EffectivePower(t, g, temple), effectstest.EffectiveToughness(t, g, temple); p != 4 || tt != 4 {
		t.Fatalf("Temple = %d/%d, want 4/4 (Temple, two Soldiers, a Bear)", p, tt)
	}

	before := len(effectstest.PopulateTokens(g, me.ID, "Soldier"))
	effectstest.AttackWith(t, g, opp.ID, temple)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != before+1 {
		t.Errorf("Soldiers = %d, want %d after the Temple connected", got, before+1)
	}
}

func TestWayfaringTempleDoesNotPopulateWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := effectstest.PushCatalogPermanent(g, me.ID, "Wayfaring Temple", "Creature — Elemental", effectstest.WayfaringTempleOracle, false)
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(temple, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, temple); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want 1: damage to a creature does not trigger it", got)
	}
}

func TestSelesnyaEulogistExilesACreatureCardThenPopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eulogist := effectstest.PushCatalogPermanent(g, me.ID, "Selesnya Eulogist", "Creature — Centaur Druid", effectstest.SelesnyaEulogistOracle, true)
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})

	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 1)
	// A mana-only cost: summoning sickness does not matter.
	if err := g.ActivateCatalogAbility(me.ID, eulogist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(dead) {
		t.Error("the creature card should be exiled")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestSelesnyaEulogistRefusesANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eulogist := effectstest.PushCatalogPermanent(g, me.ID, "Selesnya Eulogist", "Creature — Centaur Druid", effectstest.SelesnyaEulogistOracle, false)
	rock := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})

	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, eulogist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("an artifact card is not a legal target")
	}
}

func TestLifeFindsAWayPopulatesForBigNontokenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Life Finds a Way", effectstest.LifeFindsAWayOracle, "Enchantment")
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.B20CastCreature(t, g, me, "Small Bear", "Creature — Bear", "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Fatalf("a 3-power creature: Soldiers = %d, want 1", got)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("4/4 green Rhino with trample"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Rhino")); got != 1 {
		t.Fatalf("a 4/4 TOKEN must not trigger it: Rhinos = %d, want 1", got)
	}

	effectstest.B20CastCreature(t, g, me, "Big Bear", "Creature — Bear", "", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	// Soldier and Rhino differ, so the populate asks.
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("a 4-power nontoken creature should populate")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, effectstest.PopulateTokens(g, me.ID, "Soldier")[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Errorf("Soldiers = %d, want 2", got)
	}
}

func TestSongOfTheWorldsoulPopulatesOnEveryCastAndTheSpellStillResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Song of the Worldsoul", effectstest.SongOfTheWorldsoulOracle, "Enchantment")
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	life := opp.Life

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want the Bolt to still resolve", life, opp.Life)
	}
}

func TestSongOfTheWorldsoulIgnoresOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, opp.ID, "Song of the Worldsoul", effectstest.SongOfTheWorldsoulOracle, "Enchantment")
	effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, opp.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want 1: only your own casts trigger it", got)
	}
}

func TestFullFloweringPopulatesXTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.CastXSpell(t, g, "Full Flowering", "Sorcery", effectstest.FullFloweringOracle, "{X}{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 4 {
		t.Fatalf("Soldiers = %d, want 4 (one original, three copies)", got)
	}
}

func TestFullFloweringWithXZeroOrNoTokenDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.CastXSpell(t, g, "Full Flowering", "Sorcery", effectstest.FullFloweringOracle, "{X}{X}{G}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Fatalf("X=0: Soldiers = %d, want 1", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	before := len(g2.Battlefield.Cards)
	effectstest.CastXSpell(t, g2, "Full Flowering", "Sorcery", effectstest.FullFloweringOracle, "{X}{X}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := len(g2.Battlefield.Cards); got != before {
		t.Errorf("no token: battlefield %d -> %d, want no change", before, got)
	}
}

func TestFullFloweringAsksEachTimeAndCanCopyTheNewCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	centaur := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	effectstest.CastXSpell(t, g, "Full Flowering", "Sorcery", effectstest.FullFloweringOracle, "{X}{X}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, centaur)
	effectstest.PassPriorityAroundTable(t, g)
	// Second populate: Soldier / Centaur are different again.
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the second populate should ask too")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, centaur)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 3 {
		t.Errorf("Centaurs = %d, want 3", got)
	}
}

// currentPowerOf is power including +1/+1 counters, which the layer
// engine's Effective() deliberately leaves out.
func currentPowerOf(g *game.Game, id uuid.UUID) int {
	c, _ := g.LookupCardForEffect(id)
	return c.CurrentPower()
}
