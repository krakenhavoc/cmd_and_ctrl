package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b041_cards_test.go — Strixhaven: School of Mages, batch 041 (S82,
// #5502). Each test plays the real catalog entry through the engine.

const (
	oldB41KianneOracle       = "9eab65ee-e8be-4fa6-9634-fb03582bd745"
	oldB41LeylineOracle      = "9f2e713f-c7d5-4c91-9934-9e36888244fd"
	oldB41CoachOracle        = "aae4dd38-3b35-4c92-a0d4-5d37cfd93df8"
	oldB41ShaileOracle       = "b0907441-30fe-436a-b29c-585b57eabfc3"
	oldB41SummoningOracle    = "ce690421-db1d-4177-a2db-338f29aab623"
	oldB41DissectorOracle    = "d5d6a41e-84bd-4b10-9734-50d91d523d2f"
	oldB41EmergentOracle     = "dc859ce2-a267-481f-80d5-2cc38935baa0"
	oldB41SerpentineOracle   = "e07ed462-17b8-46ec-8a3f-3f5c89f4bf78"
	oldB41BodyOracle         = "e0c21b32-f672-4496-91ed-7b7f351eb539"
	oldB41ProfessorsOracle   = "e3c8e552-d315-486d-a176-813b3a300174"
	oldB41FractalTokenName   = "Fractal"
	oldB41FractalTokenColors = "GU"
)

func TestOldB41EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB41KianneOracle:        "Kianne, Dean of Substance",
		oldB41KianneOracle + "#1": "Imbraham, Dean of Theory",
		oldB41LeylineOracle:       "Leyline Invocation",
		oldB41CoachOracle:         "Dueling Coach",
		oldB41ShaileOracle:        "Shaile, Dean of Radiance",
		oldB41ShaileOracle + "#1": "Embrose, Dean of Shadow",
		oldB41SummoningOracle:     "Fractal Summoning",
		oldB41DissectorOracle:     "Novice Dissector",
		oldB41EmergentOracle:      "Emergent Sequence",
		oldB41SerpentineOracle:    "Serpentine Curve",
		oldB41BodyOracle:          "Body of Research",
		oldB41ProfessorsOracle:    "Professor's Warning",
	}
	for oracle, name := range want {
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

// oldB41Fractals are the Fractal tokens `owner` controls.
func oldB41Fractals(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == oldB41FractalTokenName && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

func oldB41OneFractal(t *testing.T, g *game.Game, owner uuid.UUID, size int) {
	t.Helper()
	fr := oldB41Fractals(g, owner)
	if len(fr) != 1 {
		t.Fatalf("%d Fractal tokens, want 1", len(fr))
	}
	if p, tough := fr[0].CurrentPower(), fr[0].CurrentToughness(); p != size || tough != size {
		t.Errorf("the Fractal is %d/%d, want %d/%d", p, tough, size, size)
	}
}

// oldB41Cast casts a catalog spell from the hand, paying `mana` exactly.
func oldB41Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	id := effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB41LeylineInvocationCountsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, g.Seats[1].ID, "Island", "Basic Land — Island")
	oldB41Cast(t, g, me, "Leyline Invocation", "Sorcery", "{5}{G}", oldB41LeylineOracle, "CCCCCG", game.CastSpellParams{})
	oldB41OneFractal(t, g, me.ID, 3)
	fr := oldB41Fractals(g, me.ID)[0]
	if got := effectstest.EffectiveColors(t, g, fr.InstanceID); len(got) != 2 {
		t.Errorf("Fractal colours %v, want green and blue", got)
	}
}

func TestOldB41LeylineInvocationWithNoLandsMakesAFractalThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB41Cast(t, g, me, "Leyline Invocation", "Sorcery", "{5}{G}", oldB41LeylineOracle, "CCCCCG", game.CastSpellParams{})
	if n := len(oldB41Fractals(g, me.ID)); n != 0 {
		t.Errorf("%d Fractals survive a 0/0 with no counters", n)
	}
}

func TestOldB41FractalSummoningPutsXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB41Cast(t, g, me, "Fractal Summoning", "Sorcery — Lesson", "{X}{G/U}{G/U}", oldB41SummoningOracle, "CCGU", game.CastSpellParams{XValue: 2})
	oldB41OneFractal(t, g, me.ID, 2)
}

func TestOldB41SerpentineCurveCountsInstantsAndSorceriesInExileAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine In Exile", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs In Exile", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hidden", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID, FaceDown: true})
	oldB41Cast(t, g, me, "Serpentine Curve", "Sorcery", "{3}{U}", oldB41SerpentineOracle, "CCCU", game.CastSpellParams{})
	// One, plus the instant in the graveyard and the face-up sorcery you own in exile.
	oldB41OneFractal(t, g, me.ID, 3)
}

func TestOldB41BodyOfResearchCountsYourLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	want := me.Library.Size()
	oldB41Cast(t, g, me, "Body of Research", "Sorcery", "{G}{G}{G}{U}{U}{U}", oldB41BodyOracle, "GGGUUU", game.CastSpellParams{})
	oldB41OneFractal(t, g, me.ID, want)
}

func TestOldB41DuelingCoachEntersAndPutsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Dueling Coach", "Creature — Human Monk", oldB41CoachOracle, 2, 2)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
}

func TestOldB41DuelingCoachActivationGrowsEveryCountedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	coach := effectstest.PushCatalogPermanent(g, me.ID, "Dueling Coach", "Creature — Human Monk", oldB41CoachOracle, false)
	withCounter := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{withCounter, theirs} {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
	}
	effectstest.FloatForTest(g, me, "CCCCW")
	effectstest.ApaActivate(t, g, me, coach, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, withCounter, game.CounterPlusOne); got != 2 {
		t.Errorf("a creature with a counter has %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature without one got %d", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature has %d, want 1 (untouched)", got)
	}
	if got := effectstest.CounterCount(g, coach, game.CounterPlusOne); got != 0 {
		t.Errorf("the Coach has %d, it had none", got)
	}
}

func TestOldB41NoviceDissectorSacrificesAnotherCreatureForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dis := effectstest.PushCatalogPermanent(g, me.ID, "Novice Dissector", "Creature — Troll Warlock", oldB41DissectorOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, dis, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

func TestOldB41NoviceDissectorCannotSacrificeItselfOrActAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dis := effectstest.PushCatalogPermanent(g, me.ID, "Novice Dissector", "Creature — Troll Warlock", oldB41DissectorOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	err := g.ActivateCatalogAbility(me.ID, dis, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{dis},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("the Dissector sacrificed itself to its own cost")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "C")
	err = g.ActivateCatalogAbility(me.ID, dis, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("activated outside a main phase")
	}
}

func TestOldB41ProfessorsWarningModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1g3CastModal(t, g, "Professor's Warning", "Instant", oldB41ProfessorsOracle, []int{0}, effectstest.P1g3ModeCard(bear, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("mode one: %d counters, want 1", got)
	}
	effectstest.P1g3CastModal(t, g, "Professor's Warning", "Instant", oldB41ProfessorsOracle, []int{1}, effectstest.P1g3ModeCard(bear, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("mode two: the Bear is not indestructible")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("mode two added a counter: %d", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("indestructible lasts past the turn")
	}
}

func TestOldB41EmergentSequenceFetchesAFractalLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// A land already played this turn counts too.
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: uuid.New(), OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(forest)
	before := g.TurnTallyFor(me.ID).LandsEntered
	oldB41Cast(t, g, me, "Emergent Sequence", "Sorcery", "{1}{G}", oldB41EmergentOracle, "CG", game.CastSpellParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, forest.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, forest.InstanceID)
	if !ok {
		t.Fatal("the Forest did not enter")
	}
	if !c.Tapped {
		t.Error("the land did not enter tapped")
	}
	if !c.IsLand() || !c.IsCreature() {
		t.Errorf("land %v creature %v, want both", c.IsLand(), c.IsCreature())
	}
	want := before + 1
	if got := c.Counters[game.CounterPlusOne]; got != want {
		t.Errorf("%d +1/+1 counters, want %d (lands that entered this turn)", got, want)
	}
	if p := c.CurrentPower(); p != want {
		t.Errorf("power %d, want %d", p, want)
	}
	if !c.HasSubtype("Fractal") {
		t.Error("the land is not a Fractal")
	}
	colors := effectstest.EffectiveColors(t, g, forest.InstanceID)
	if len(colors) != 2 {
		t.Errorf("colours %v, want green and blue", colors)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, forest.InstanceID, "haste") {
		t.Error("the land gained haste, which the card does not say")
	}
}

// --- Kianne, Dean of Substance // Imbraham, Dean of Theory ---------------------

func oldB41Dean(owner uuid.UUID, oracle string, f0, f1 game.Face, face int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner, Faces: []game.Face{f0, f1},
	}
	c.SetFace(face)
	return c
}

func oldB41Kianne(owner uuid.UUID, face int) game.Card {
	return oldB41Dean(owner, oldB41KianneOracle,
		game.Face{Name: "Kianne, Dean of Substance", TypeLine: "Legendary Creature — Elf Druid", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2},
		game.Face{Name: "Imbraham, Dean of Theory", TypeLine: "Legendary Creature — Bird Wizard", ManaCost: "{2}{U}{U}", Colors: []string{"U"}, Power: 3, Toughness: 3},
		face)
}

func oldB41Exiled(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID,
		Counters: map[string]int{effects.OldB41StudyCounter: 1}})
	return id
}

func TestOldB41KianneExilesAndCountersANonlandAndTakesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kianne := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Kianne(me.ID, 0))
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaActivate(t, g, me, kianne, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCountInExile(g, spell, effects.OldB41StudyCounter); got != 1 {
		t.Fatalf("the exiled nonland has %d study counters, want 1", got)
	}
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	effectstest.P1g3Untap(g, kianne)
	effectstest.ApaActivate(t, g, me, kianne, 0, game.ActivateAbilityParams{})
	if !effectstest.InHandOf(me, land) {
		t.Error("an exiled land card did not go to the hand")
	}
	if got := effectstest.CounterCountInExile(g, land, effects.OldB41StudyCounter); got > 0 {
		t.Errorf("the land took %d study counters", got)
	}
}

func TestOldB41KianneFractalCountsDifferentManaValues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kianne := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Kianne(me.ID, 0))
	oldB41Exiled(g, me, "Two A", "Sorcery", "{1}{G}")
	oldB41Exiled(g, me, "Two B", "Instant", "{1}{U}")
	oldB41Exiled(g, me, "Four", "Creature — Bear", "{3}{G}")
	oldB41Exiled(g, me, "Land", "Basic Land — Forest", "")
	oldB41Exiled(g, opp, "Theirs", "Sorcery", "{5}")
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "No Counter", TypeLine: "Sorcery", ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CCCCG")
	effectstest.ApaActivate(t, g, me, kianne, 1, game.ActivateAbilityParams{})
	oldB41OneFractal(t, g, me.ID, 2)
}

func TestOldB41ImbrahamExilesXCardsCountersThemAndTakesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	imb := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Kianne(me.ID, 1))
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	for _, p := range []struct {
		id   uuid.UUID
		name string
		tl   string
	}{{a, "A", "Sorcery"}, {b, "B", "Basic Land — Forest"}, {c, "C", "Creature — Bear"}} {
		me.Library.Cards = append(me.Library.Cards, game.Card{InstanceID: p.id, Name: p.name, TypeLine: p.tl, Owner: me.ID, Controller: me.ID})
	}
	effectstest.FloatForTest(g, me, "CCUU")
	if err := g.ActivateCatalogAbility(me.ID, imb, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	var counted uuid.UUID
	for _, ec := range g.Exile.Cards {
		if ec.Counters[effects.OldB41StudyCounter] == 1 {
			exiled++
			counted = ec.InstanceID
		}
	}
	if exiled != 2 {
		t.Fatalf("%d cards exiled with study counters, want 2 (X = 2)", exiled)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, counted)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, counted) {
		t.Error("the chosen study card did not go to the hand")
	}
}

func TestOldB41ImbrahamAtXZeroStillOffersAStudyCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	imb := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Kianne(me.ID, 1))
	old := oldB41Exiled(g, me, "Earlier", "Sorcery", "{2}")
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, imb, 0, game.ActivateAbilityParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, old)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, old) {
		t.Error("the study card from an earlier turn did not go to the hand")
	}
}

// --- Shaile, Dean of Radiance // Embrose, Dean of Shadow -----------------------

func oldB41Shaile(owner uuid.UUID, face int) game.Card {
	return oldB41Dean(owner, oldB41ShaileOracle,
		game.Face{Name: "Shaile, Dean of Radiance", TypeLine: "Legendary Creature — Bird Cleric", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 1, Toughness: 1},
		game.Face{Name: "Embrose, Dean of Shadow", TypeLine: "Legendary Creature — Human Warlock", ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: 4, Toughness: 4},
		face)
}

func TestOldB41ShaileCountersOnlyCreaturesThatEnteredThisTurnUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shaile := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Shaile(me.ID, 0))
	old := effectstest.B12Creature(g, me.ID, "Old", "Creature — Bear", 2, 2)
	// An opponent's creature that entered this turn is not yours.
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	fresh := effectstest.CastHoldCreature(t, g, "Fresh", "Creature — Bear", "", 1, 1)
	effectstest.ApaActivate(t, g, me, shaile, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, fresh, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature that entered this turn has %d counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, old, game.CounterPlusOne); got != 0 {
		t.Errorf("an old creature got %d counters", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestOldB41EmbroseCountersThenDamagesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	embrose := effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Shaile(me.ID, 1))
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.ApaActivate(t, g, me, embrose, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	if got := effectstest.Pr6Marked(g, bear); got != 2 {
		t.Errorf("the Bear has %d damage, want 2", got)
	}
	// Embrose is not a legal target of her own ability.
	effectstest.P1g3Untap(g, embrose)
	err := g.ActivateCatalogAbility(me.ID, embrose, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: embrose}}})
	if err == nil {
		t.Error("Embrose targeted herself")
	}
}

func TestOldB41EmbroseDrawsWhenACreatureWithACounterDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, oldB41Shaile(me.ID, 1))
	effectstest.SeedLibrary(me, "A", "B", "C")
	mine := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	hand := me.Hand.Size()
	effectstest.P1Destroy(t, g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a creature with no counter drew a card")
	}
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("an opponent's creature drew a card")
	}
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after a creature you control with a counter died", me.Hand.Size(), hand+1)
	}
}
