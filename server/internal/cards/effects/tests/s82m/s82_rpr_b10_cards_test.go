package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set RPR (reprints counted in the S82 sets), batch 10 (#5908). Each
// test plays the real catalog entry through the engine.

const (
	rprValiantKnightOracle      = "b5c03abf-4297-4402-9107-539e92d41f2f"
	rprHumiliatorOracle         = "17c3a888-84d2-4c21-8fa3-d2d825dd7601"
	rprLoxodonOracle            = "d157a9df-59bc-40ad-b62b-43f973f14be1"
	rprWeirdingWoodOracle       = "9ce5cc2f-e223-4c3b-adc1-c0109fd31762"
	rprWintermoorOracle         = "b1e4cf31-9efe-4aca-a3ba-cc2681109a2f"
	rprWorkshopEldersOracle     = "9e9164be-3534-4445-bfad-00d16a50575c"
	rprWorthyKnightOracle       = "37733e29-3ade-4bfd-89c9-efd13b548a53"
	rprVoltChargeOracle         = "c57907d1-c5a1-45c7-80a8-38fedf86dd19"
	rprAdrianaOracle            = "de23275e-e9c1-4151-aed8-f79ead03d5d4"
	rprBlightMambaOracle        = "e11a856c-31d3-4f4b-a517-8a5134b88517"
	rprCaressOfPhyrexiaOracle   = "8253fcac-0c07-41e8-afa4-34021f9ef6f4"
	rprKnightTypeLine           = "Creature — Human Knight"
	rprArtifactTypeLine         = "Artifact"
	rprArtifactCreatureTypeLine = "Artifact Creature — Golem"
)

func TestRprB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprValiantKnightOracle:    "Valiant Knight",
		rprHumiliatorOracle:       "Vedalken Humiliator",
		rprLoxodonOracle:          "Venerated Loxodon",
		rprWeirdingWoodOracle:     "Weirding Wood",
		rprWintermoorOracle:       "Wintermoor Commander",
		rprWorkshopEldersOracle:   "Workshop Elders",
		rprWorthyKnightOracle:     "Worthy Knight",
		rprVoltChargeOracle:       "Volt Charge",
		rprAdrianaOracle:          "Adriana, Captain of the Guard",
		rprBlightMambaOracle:      "Blight Mamba",
		rprCaressOfPhyrexiaOracle: "Caress of Phyrexia",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Adriana, Captain of the Guard -------------------------------------

func TestRprB10AdrianaGivesOtherCreaturesYouControlMelee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adriana := effectstest.B27Push(g, me.ID, "Adriana, Captain of the Guard", "Legendary Creature — Human Knight", rprAdrianaOracle, "{3}{R}{W}", 4, 4, "R", "W")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, adriana, game.KeywordMelee) || !effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordMelee) {
		t.Error("Adriana and your other creatures have melee")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, game.KeywordMelee) {
		t.Error("an opponent's creature has melee")
	}
}

func TestRprB10AdrianaMeleeCountsEachOpponentAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adriana := effectstest.B27Push(g, me.ID, "Adriana, Captain of the Guard", "Legendary Creature — Human Knight", rprAdrianaOracle, "{3}{R}{W}", 4, 4, "R", "W")
	effectstest.DeclareAttack(t, g, opp.ID, adriana)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, adriana); p != 5 || tt != 5 {
		t.Errorf("Adriana attacking one opponent is %d/%d, want 5/5", p, tt)
	}
}

// --- Blight Mamba -----------------------------------------------------

func TestRprB10BlightMambaHasInfectAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mamba := effectstest.B12Push(g, me.ID, "Blight Mamba", "Creature — Phyrexian Snake", rprBlightMambaOracle, 1, 1)
	spec, _ := effects.Lookup(rprBlightMambaOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "infect" {
		t.Errorf("printed keywords %v, want infect", spec.PrintedKeywords)
	}
	effectstest.ApaMana(me, "C", "G")
	effectstest.ApaActivate(t, g, me, mamba, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mamba) })
	if !g.Battlefield.Contains(mamba) {
		t.Fatal("the shielded Mamba was destroyed")
	}
	if !effectstest.AlliTapped(t, g, mamba) {
		t.Error("regeneration did not tap it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mamba) })
	if g.Battlefield.Contains(mamba) {
		t.Error("the shield was not used up: the Mamba survived a second destroy")
	}
}

func TestRprB10BlightMambaRegenerationNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mamba := effectstest.B12Push(g, me.ID, "Blight Mamba", "Creature — Phyrexian Snake", rprBlightMambaOracle, 1, 1)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, mamba, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated {1}{G} with one mana")
	}
}

// --- Worthy Knight ----------------------------------------------------

func TestRprB10WorthyKnightMakesAHumanWhenYouCastAKnightOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B27Push(g, me.ID, "Worthy Knight", rprKnightTypeLine, rprWorthyKnightOracle, "{1}{W}", 2, 2, "W")
	effectstest.CastCatalogSpell(t, g, "Squire", rprKnightTypeLine, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("%d Human tokens after casting a Knight, want 1", n)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("%d Human tokens after casting a Bear, want still 1", n)
	}
}

// --- Valiant Knight ---------------------------------------------------

func TestRprB10ValiantKnightPumpsOtherKnightsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valiant := effectstest.B27Push(g, me.ID, "Valiant Knight", rprKnightTypeLine, rprValiantKnightOracle, "{3}{W}", 3, 4, "W")
	squire := effectstest.B12Creature(g, me.ID, "Squire", rprKnightTypeLine, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", rprKnightTypeLine, 2, 2)
	for _, c := range []struct {
		id         uuid.UUID
		name       string
		power, tou int
	}{{valiant, "Valiant Knight", 3, 4}, {squire, "your other Knight", 3, 3}, {bear, "a Bear", 2, 2}, {theirs, "their Knight", 2, 2}} {
		if p, tt := effectstest.EffectivePTOf(t, g, c.id); p != c.power || tt != c.tou {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tt, c.power, c.tou)
		}
	}
}

func TestRprB10ValiantKnightGivesKnightsDoubleStrikeForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), g.Seats[1]
	valiant := effectstest.B12Push(g, me.ID, "Valiant Knight", rprKnightTypeLine, rprValiantKnightOracle, 3, 4)
	squire := effectstest.B12Creature(g, me.ID, "Squire", rprKnightTypeLine, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", rprKnightTypeLine, 2, 2)
	effectstest.ApaMana(me, "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, valiant, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated {3}{W}{W} with four mana")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "W", "W", "W", "W", "W")
	effectstest.ApaActivate(t, g, me, valiant, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{valiant, squire} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
			t.Error("a Knight you control did not gain double strike")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") || effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("a non-Knight or an opponent's Knight gained double strike")
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if effectstest.HasEffectiveKeyword(t, g, valiant, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

// --- Caress of Phyrexia -----------------------------------------------

func TestRprB10CaressOfPhyrexiaDrawsThreeLosesThreeAndPoisonsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 5; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	}
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.CastCatalogSpell(t, g, "Caress of Phyrexia", "Sorcery", rprCaressOfPhyrexiaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
	if got := opp.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("%d poison counters, want 3", got)
	}
	if me.Counters[game.CounterPoison] != 0 {
		t.Error("the caster got poison")
	}
}

// --- Volt Charge ------------------------------------------------------

func TestRprB10VoltChargeDamagesThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Volt Charge", "Instant", rprVoltChargeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("dealt %d, want 3", got)
	}
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 after proliferate", got)
	}
}

// --- Venerated Loxodon ------------------------------------------------

func TestRprB10VeneratedLoxodonCountersItsConvokers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 2)
	bystander := effectstest.PushVanillaCreature(g, me.ID, "Bystander", 1, 1)
	_, loxodon, err := s82Cast(t, g, "Venerated Loxodon", "Creature — Elephant Cleric", "{4}{W}",
		rprLoxodonOracle, []string{"W"}, "WWW", taps, nil)
	if err != nil {
		t.Fatalf("cast with two convokers: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range taps {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("convoker +1/+1 counters: %d, want 1", got)
		}
	}
	if effectstest.CounterCount(g, bystander, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, loxodon, game.CounterPlusOne) != 0 {
		t.Error("a creature that did not convoke got a counter")
	}
}

func TestRprB10VeneratedLoxodonCastNormallyPutsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bystander := effectstest.PushVanillaCreature(g, me.ID, "Bystander", 1, 1)
	_, loxodon, err := s82Cast(t, g, "Venerated Loxodon", "Creature — Elephant Cleric", "{4}{W}",
		rprLoxodonOracle, []string{"W"}, "WWWWW", nil, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, bystander, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, loxodon, game.CounterPlusOne) != 0 {
		t.Error("a Loxodon cast with mana only put counters on something")
	}
}

// --- Wintermoor Commander ---------------------------------------------

func TestRprB10WintermoorToughnessCountsYourKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := effectstest.B27Push(g, me.ID, "Wintermoor Commander", rprKnightTypeLine, rprWintermoorOracle, "{W}{B}", 2, 0, "W", "B")
	if p, tt := effectstest.EffectivePTOf(t, g, cmdr); p != 2 || tt != 1 {
		t.Errorf("alone it is %d/%d, want 2/1", p, tt)
	}
	effectstest.B12Creature(g, me.ID, "Squire", rprKnightTypeLine, 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Knight", rprKnightTypeLine, 2, 2)
	if p, tt := effectstest.EffectivePTOf(t, g, cmdr); p != 2 || tt != 2 {
		t.Errorf("with one more of your Knights it is %d/%d, want 2/2", p, tt)
	}
}

func TestRprB10WintermoorAttackGivesAnotherKnightIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := effectstest.B27Push(g, me.ID, "Wintermoor Commander", rprKnightTypeLine, rprWintermoorOracle, "{W}{B}", 2, 0, "W", "B")
	squire := effectstest.B12Creature(g, me.ID, "Squire", rprKnightTypeLine, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", rprKnightTypeLine, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if !offered[squire] || offered[cmdr] || offered[bear] || offered[theirs] {
		t.Errorf("offered %v: want only the other Knight you control", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, squire)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, squire, "indestructible") {
		t.Error("the targeted Knight did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, cmdr, "indestructible") {
		t.Error("the Commander itself gained indestructible")
	}
}

// --- Vedalken Humiliator ----------------------------------------------

func rprHumiliatorBoard(t *testing.T, artifacts int) (g *game.Game, me, opp *game.Player, humiliator, theirs uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	humiliator = effectstest.B27Push(g, me.ID, "Vedalken Humiliator", "Creature — Vedalken Wizard", rprHumiliatorOracle, "{3}{U}", 3, 4, "U")
	for i := 0; i < artifacts; i++ {
		effectstest.B12Permanent(g, me.ID, "Rock", rprArtifactTypeLine)
	}
	theirs = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Flier", TypeLine: "Creature — Bird",
		Power: 4, Toughness: 4, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	return
}

func TestRprB10HumiliatorWithMetalcraftShrinksAndSilencesOpposingCreatures(t *testing.T) {
	g, me, opp, humiliator, theirs := rprHumiliatorBoard(t, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, humiliator)
	effectstest.SncSettle(t, g, me.ID, uuid.Nil)
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", p, tt)
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("their creature kept flying")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 2 || tt != 2 {
		t.Errorf("your own creature is %d/%d, want untouched 2/2", p, tt)
	}
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 3, 3)
	if p, _ := effectstest.EffectivePTOf(t, g, late); p != 3 {
		t.Errorf("a creature that entered later is %d power, want 3 (the set is locked at resolution)", p)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 4 || tt != 4 || !effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Errorf("their creature is %d/%d after the turn, want 4/4 flying again", p, tt)
	}
}

func TestRprB10HumiliatorWithTwoArtifactsDoesNothing(t *testing.T) {
	g, me, opp, humiliator, theirs := rprHumiliatorBoard(t, 2)
	effectstest.DeclareAttack(t, g, opp.ID, humiliator)
	if effectstest.TriggersOnStackFrom(g, humiliator) != 0 {
		t.Error("the metalcraft trigger went on the stack with two artifacts")
	}
	effectstest.SncSettle(t, g, me.ID, uuid.Nil)
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 4 || tt != 4 {
		t.Errorf("their creature is %d/%d, want untouched 4/4", p, tt)
	}
}

// --- Workshop Elders --------------------------------------------------

func TestRprB10WorkshopEldersGivesYourArtifactCreaturesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elders := effectstest.B27Push(g, me.ID, "Workshop Elders", "Creature — Human Artificer", rprWorkshopEldersOracle, "{6}{U}", 4, 4, "U")
	golem := effectstest.B12Creature(g, me.ID, "Golem", rprArtifactCreatureTypeLine, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", rprArtifactTypeLine)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", rprArtifactCreatureTypeLine, 3, 3)
	for id, want := range map[uuid.UUID]bool{elders: false, golem: true, bear: false, rock: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "flying"); got != want {
			t.Errorf("%s flying = %v, want %v", id, got, want)
		}
	}
}

func rprEldersAtCombat(t *testing.T) (g *game.Game, me *game.Player, rock, golem uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.B27Push(g, me.ID, "Workshop Elders", "Creature — Human Artificer", rprWorkshopEldersOracle, "{6}{U}", 4, 4, "U")
	rock = effectstest.B12Permanent(g, me.ID, "Rock", rprArtifactTypeLine)
	golem = effectstest.B12Creature(g, me.ID, "Golem", rprArtifactCreatureTypeLine, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	return
}

func TestRprB10WorkshopEldersAnimatesANoncreatureArtifactWithFourCounters(t *testing.T) {
	g, me, rock, golem := rprEldersAtCombat(t)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == golem {
			t.Error("an artifact creature was offered")
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, rock); p != 4 || tt != 4 {
		t.Errorf("the artifact is %d/%d, want a 4/4 (0/0 with four counters)", p, tt)
	}
	if got := effectstest.CounterCount(g, rock, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	c, _ := effectstest.BattlefieldCard(g, rock)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Error("the artifact is not an artifact creature")
	}
	if !effectstest.HasEffectiveKeyword(t, g, rock, "flying") {
		t.Error("the new artifact creature lacks flying from the Elders")
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if c, _ := effectstest.BattlefieldCard(g, rock); !c.IsCreature() {
		t.Error("the animation ended with the turn; it has no duration")
	}
}

func TestRprB10WorkshopEldersDeclinedDoesNothing(t *testing.T) {
	g, me, rock, _ := rprEldersAtCombat(t)
	effectstest.PickTriggerTarget(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, rock); c.IsCreature() || effectstest.CounterCount(g, rock, game.CounterPlusOne) != 0 {
		t.Error("declining the may still animated the artifact or put counters")
	}
}

// --- Weirding Wood ----------------------------------------------------

func TestRprB10WeirdingWoodInvestigatesAndTapsTheLandForTwoOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Weirding Wood", effectstest.AuraTypeLine, rprWeirdingWoodOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clue tokens, want 1", n)
	}
	var produced string
	g.ReadSnapshot(func() {
		card, _ := g.LookupCardForEffect(land)
		abs, origins := game.ManaAbilitiesWithOrigins(card)
		for i, o := range origins {
			if o.Granted() {
				produced = abs[i].Produced
			}
		}
	})
	if produced != effects.OneColorOfAmount(2) {
		t.Errorf("granted mana ability produces %q, want two mana of any one color", produced)
	}
	creature := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := neoB12CastStrict(t, g, "Weirding Wood", rprWeirdingWoodOracle, creature); err == nil {
		t.Error("Weirding Wood enchanted a creature")
	}
}
