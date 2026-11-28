package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b811_test.go — Mirage, batch 811 (#5633): Auras.

const (
	oldB811Ritual    = "0af91b63-a874-4184-8f8f-3871429fd1ad"
	oldB811Thirst    = "61b3ad9d-485f-4479-b18f-506847af43f2"
	oldB811Cloak     = "68b06917-7f83-47d4-a466-4e6698cae56d"
	oldB811Curse     = "6e422d5e-d9f5-492d-b82a-31f7274d4153"
	oldB811Wellspr   = "8650bd97-ecf6-420f-9b72-a749970afcd4"
	oldB811Ferocity  = "8efbe68c-622c-4986-a375-7e0bf780eb2e"
	oldB811Harness   = "bba30360-c54a-4906-8b2e-84bb1209c43d"
	oldB811Agony     = "dcd41ed1-4a12-416b-8ba0-51a9c0927ddb"
	oldB811Destiny   = "ee04afb0-184f-4a13-8d44-0fbe121b1eff"
	oldB811AuraType  = "Enchantment — Aura"
	oldB811BearType  = "Creature — Bear"
	oldB811WallType  = "Creature — Wall"
	oldB811StrictErr = "the cast was refused"
)

func TestOldB811EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB811Ritual: "Ritual of Steel", oldB811Thirst: "Thirst", oldB811Cloak: "Cloak of Invisibility",
		oldB811Curse: "Teferi's Curse", oldB811Wellspr: "Wellspring", oldB811Ferocity: "Consuming Ferocity",
		oldB811Harness: "Mind Harness", oldB811Agony: "Binding Agony", oldB811Destiny: "Favorable Destiny",
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

func oldB811Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB811Aura casts the Aura from the active player's hand onto target and
// resolves it. The cast is strict, so an illegal target is refused.
func oldB811Aura(t *testing.T, g *game.Game, name, oracle, mana string, target uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	me, _ := oldB811Seats(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: oldB811AuraType, ManaCost: mana, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target), Strict: true}); err != nil {
		return id, err
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id, nil
}

func oldB811Prompt(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == who {
			return c
		}
	}
	return nil
}

func TestOldB811RitualOfSteelPumpsToughnessAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB811Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB811BearType, 2, 2)
	hand := len(me.Hand.Cards)
	if _, err := oldB811Aura(t, g, "Ritual of Steel", oldB811Ritual, "", bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 2/4", p, tough)
	}
	if len(me.Hand.Cards) != hand {
		t.Fatalf("a card was drawn on entering (hand %d, was %d)", len(me.Hand.Cards), hand)
	}
	// The next turn's upkeep is another player's: the draw is still mine.
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOfSeat(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d after the next upkeep", len(me.Hand.Cards), hand+1)
	}
	// My own upkeep, before my draw step: the delayed draw is spent.
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d: the draw happened more than once", len(me.Hand.Cards))
	}
}

func TestOldB811ThirstTapsLocksAndDemandsU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB811BearType, 2, 2)
	thirst, err := oldB811Aura(t, g, "Thirst", oldB811Thirst, "", theirs)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !effectstest.TappedNow(g, theirs) {
		t.Fatal("the enchanted creature was not tapped")
	}
	// Their untap step: it stays tapped.
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat) // wraps all the way round to my upkeep
	if !effectstest.TappedNow(g, theirs) {
		t.Error("the enchanted creature untapped")
	}
	prompt := oldB811Prompt(g, me.ID)
	if prompt == nil {
		effectstest.PassPriorityAroundTable(t, g)
		prompt = oldB811Prompt(g, me.ID)
	}
	if prompt == nil {
		t.Fatal("no pay-{U}-or-sacrifice offer at my upkeep")
	}
	effectstest.FloatMana(t, g, me, "{U}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(thirst) {
		t.Fatal("Thirst was sacrificed though {U} was paid")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thirst) {
		t.Error("Thirst stayed though {U} was not paid")
	}
}

func TestOldB811CloakOfInvisibilityGivesPhasingAndOnlyWallsBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB811BearType, 2, 2)
	if _, err := oldB811Aura(t, g, "Cloak of Invisibility", oldB811Cloak, "", bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordPhasing) {
		t.Error("the enchanted creature does not have phasing")
	}
	other := effectstest.B12Creature(g, opp.ID, "Blocker", oldB811BearType, 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", oldB811WallType, 0, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(other, bear); err == nil {
		t.Error("a non-Wall blocked the enchanted creature")
	}
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Errorf("a Wall could not block the enchanted creature: %v", err)
	}
}

func TestOldB811TeferisCurseGivesAnArtifactOrCreaturePhasingAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB811Seats(g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if _, err := oldB811Aura(t, g, "Teferi's Curse", oldB811Curse, "", land); err == nil {
		t.Error("the Curse was cast on a land")
	}
	curse, err := oldB811Aura(t, g, "Teferi's Curse", oldB811Curse, "", rock)
	if err != nil {
		t.Fatalf("cast on an artifact: %v", err)
	}
	if !g.Battlefield.Contains(curse) {
		t.Fatal("the Curse did not enter")
	}
	if !effectstest.HasEffectiveKeyword(t, g, rock, game.KeywordPhasing) {
		t.Error("the enchanted artifact does not have phasing")
	}
}

func TestOldB811WellspringTakesTheLandForTheTurnAndAgainEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.TapCard(g, island)
	if _, err := oldB811Aura(t, g, "Wellspring", oldB811Wellspr, "", island); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := effectstest.ControllerOf(t, g, island); got != me.ID {
		t.Fatalf("the land is controlled by %s, want me for the turn", got)
	}
	if !effectstest.TappedNow(g, island) {
		t.Error("entering must not untap the land")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if got := effectstest.ControllerOf(t, g, island); got != opp.ID {
		t.Errorf("control stayed with %s after the turn ended", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, island); got != me.ID {
		t.Errorf("the land is controlled by %s at my upkeep, want me", got)
	}
	if effectstest.TappedNow(g, island) {
		t.Error("the land was not untapped at my upkeep")
	}
}

func oldB811SeatIndex(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func TestOldB811ConsumingFerocityGrowsThenBurnsAndDestroysAtThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", oldB811WallType, 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB811BearType, 2, 6)
	if _, err := oldB811Aura(t, g, "Consuming Ferocity", oldB811Ferocity, "", wall); err == nil {
		t.Error("the Aura was cast on a Wall")
	}
	if _, err := oldB811Aura(t, g, "Consuming Ferocity", oldB811Ferocity, "", bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if p := effects.B31CurrentPowerOf(g, bear); p != 3 {
		t.Errorf("power %d, want 3 (+1/+0)", p)
	}
	life := opp.Life
	for i := 1; i <= 2; i++ {
		effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
		effectstest.PassPriorityAroundTable(t, g)
		if p := effects.B31CurrentPowerOf(g, bear); p != 3+i {
			t.Errorf("after %d upkeeps power %d, want %d", i, p, 3+i)
		}
		if !g.Battlefield.Contains(bear) || opp.Life != life {
			t.Fatalf("after %d upkeeps: bear on battlefield %v, life %d→%d", i, g.Battlefield.Contains(bear), life, opp.Life)
		}
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("life %d→%d, want -6 (power 2 +1 +3 counters)", life, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived its third counter")
	}
}

func TestOldB811MindHarnessStealsARedOrGreenCreatureForACumulativeUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	blue := effectstest.B16Creature(g, opp.ID, "Blue Bear", oldB811BearType, 2, 2, "U")
	red := effectstest.B16Creature(g, opp.ID, "Red Bear", oldB811BearType, 2, 2, "R")
	if _, err := oldB811Aura(t, g, "Mind Harness", oldB811Harness, "", blue); err == nil {
		t.Error("the Aura was cast on a blue creature")
	}
	harness, err := oldB811Aura(t, g, "Mind Harness", oldB811Harness, "", red)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := effectstest.ControllerOf(t, g, red); got != me.ID {
		t.Fatalf("the creature is controlled by %s, want me", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if oldB811Prompt(g, me.ID) == nil {
		t.Fatal("no cumulative upkeep offer")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(harness) || effectstest.AgeCountersOn(t, g, harness) != 1 {
		t.Fatalf("after one payment: on battlefield %v, age counters %d", g.Battlefield.Contains(harness), effectstest.AgeCountersOn(t, g, harness))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB811SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(harness) {
		t.Error("Mind Harness stayed though its upkeep was not paid")
	}
	if got := effectstest.ControllerOf(t, g, red); got != opp.ID {
		t.Errorf("control stayed with %s after the Aura left", got)
	}
}

func TestOldB811BindingAgonyReflectsDamageToTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB811Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB811BearType, 2, 6)
	if _, err := oldB811Aura(t, g, "Binding Agony", oldB811Agony, "", bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life, mine := opp.Life, me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 || me.Life != mine {
		t.Errorf("life me %d→%d, opponent %d→%d, want only the creature's controller to lose 3", mine, me.Life, life, opp.Life)
	}
	// Lethal damage still costs the controller.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 4) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Errorf("life %d→%d, want -7 in all", life, opp.Life)
	}
}

func TestOldB811FavorableDestinyPumpsWhiteCreaturesAndShroudsWithCompany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB811Seats(g)
	white := effectstest.B16Creature(g, me.ID, "White Bear", oldB811BearType, 2, 2, "W")
	green := effectstest.B16Creature(g, me.ID, "Green Bear", oldB811BearType, 2, 2, "G")
	if _, err := oldB811Aura(t, g, "Favorable Destiny", oldB811Destiny, "", white); err != nil {
		t.Fatalf("cast on white: %v", err)
	}
	if _, err := oldB811Aura(t, g, "Favorable Destiny", oldB811Destiny, "", green); err != nil {
		t.Fatalf("cast on green: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, white); p != 3 || tough != 4 {
		t.Errorf("white Bear is %d/%d, want 3/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, green); p != 2 || tough != 2 {
		t.Errorf("green Bear is %d/%d, want 2/2 (not white)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, white, game.KeywordShroud) {
		t.Error("the white Bear has no shroud though its controller controls another creature")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), green, 5) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, white, game.KeywordShroud) {
		t.Error("the white Bear kept shroud with no other creature")
	}
}
