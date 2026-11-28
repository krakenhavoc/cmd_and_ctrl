package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b654_cards_test.go — Invasion, batch 654 (S82, #5616). Each test
// plays the real catalog entry through the engine.

const (
	oldB654DisruptOracle   = "76347258-71d5-4874-adde-ff7aa932b808"
	oldB654ClimberOracle   = "91320afd-1d42-4cb5-ae40-4eed2fa91dfe"
	oldB654TrekOracle      = "929c231a-12d9-4f1a-9849-64a8b5a65f20"
	oldB654SwayOracle      = "9a7e2298-8855-43a7-8cb1-4b31e4058c3f"
	oldB654TrailOracle     = "a26c4076-c488-4700-8e32-4ceeeac2c29a"
	oldB654ResponseOracle  = "b2faa8b6-e433-4171-9774-9170484530c4"
	oldB654StunOracle      = "d3491972-44c5-4962-a680-42b79357a189"
	oldB654DredgeOracle    = "efc9498a-b740-44de-aef3-01a1fa4ee49a"
	oldB654ShockOracle     = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
	oldB654StoneRainOracle = "6e880df6-fc00-43d2-a9c8-f575f40b78c6"
	oldB654IcyOracle       = "3608f1f7-8dc5-4dd1-ae91-c830e1de9529"
)

func TestOldB654EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB654DisruptOracle: "Disrupt", oldB654ClimberOracle: "Kavu Climber", oldB654TrekOracle: "Scouting Trek",
		oldB654SwayOracle: "Sway of Illusion", oldB654TrailOracle: "Quirion Trailblazer",
		oldB654ResponseOracle: "Teferi's Response", oldB654StunOracle: "Stun", oldB654DredgeOracle: "Dredge",
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

func oldB654Seed(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.SeedLibraryTop(p, "Filler", "Sorcery", "{1}")
	}
}

func oldB654Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	return effectstest.CastFromHandForTest(t, g, p, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true, Targets: targets})
}

func oldB654Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestOldB654DisruptCountersUnlessPaidAndDraws(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		oldB654Seed(me, 3)
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
		shock := oldB654Cast(t, g, opp, "Shock", "Instant", "{R}", oldB654ShockOracle, "R", oldB654Card(bear))
		hand := effectstest.HandCount(g, me)
		oldB654Cast(t, g, me, "Disrupt", "Instant", "{U}", oldB654DisruptOracle, "U", oldB654Card(shock))
		effectstest.FloatForTest(g, opp, "C")
		effectstest.PassPriorityUntilChoice(t, g)
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HandCount(g, me); got != hand+1 {
			t.Errorf("pay=%v: hand %d, want %d (the draw)", pay, got, hand+1)
		}
		countered := effectstest.GraveyardHas(opp, shock) && effectstest.DamageMarkedOn(g, bear) == 0
		if pay && effectstest.DamageMarkedOn(g, bear) != 2 {
			t.Errorf("paid {1} but Shock did not resolve")
		}
		if !pay && !countered {
			t.Errorf("did not pay but Shock was not countered")
		}
	}
}

func TestOldB654DisruptRefusesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.CastFromHandForTest(t, g, opp, "Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.FloatForTest(g, me, "U")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Disrupt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oldB654DisruptOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{oldB654Card(bear)}}); err == nil {
		t.Error("Disrupt targeted a creature spell")
	}
}

func TestOldB654KavuClimberDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB654Seed(me, 3)
	before := effectstest.HandCount(g, me)
	effectstest.CastHoldCreature(t, g, "Kavu Climber", "Creature — Kavu", oldB654ClimberOracle, 3, 3)
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d (the cast nets zero, one drawn)", got, before+1)
	}
}

func TestOldB654ScoutingTrekPutsTheLandsOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 5)
	var forests []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		forests = append(forests, id)
		me.Library.PushBottom(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CG")
	effectstest.CastFromHandForTest(t, g, me, "Scouting Trek", "Sorcery", "{1}{G}", oldB654TrekOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forests...)
	effectstest.PassPriorityAroundTable(t, g)
	top := map[uuid.UUID]bool{}
	for i := 0; i < 2; i++ {
		c, err := me.Library.PopTop()
		if err != nil {
			t.Fatal(err)
		}
		top[c.InstanceID] = true
	}
	for _, id := range forests {
		if !top[id] {
			t.Errorf("a found Forest is not among the top two cards")
		}
	}
}

func TestOldB654ScoutingTrekMayFindNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 3)
	me.Library.PushBottom(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CG")
	effectstest.CastFromHandForTest(t, g, me, "Scouting Trek", "Sorcery", "{1}{G}", oldB654TrekOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the spell did not finish")
	}
}

func TestOldB654SwayOfIllusionRecoloursEveryTargetAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 3)
	a := effectstest.SeedColoredCreature(g, me.ID, "Green Bear", []string{"G"}, 0)
	b := effectstest.SeedColoredCreature(g, g.Seats[1].ID, "Red Bear", []string{"R"}, 0)
	untouched := effectstest.SeedColoredCreature(g, me.ID, "Black Bear", []string{"B"}, 0)
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Sway of Illusion", "Instant", "{1}{U}", oldB654SwayOracle, "CU", oldB654Card(a), oldB654Card(b))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.EffectiveColors(t, g, id); len(got) != 1 || got[0] != "U" {
			t.Errorf("colours %v, want only blue", got)
		}
	}
	if got := effectstest.EffectiveColors(t, g, untouched); len(got) != 1 || got[0] != "B" {
		t.Errorf("an untargeted creature is %v, want black", got)
	}
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if got := effectstest.EffectiveColors(t, g, a); len(got) != 1 || got[0] != "G" {
		t.Errorf("colour %v after cleanup, want green again", got)
	}
}

func TestOldB654SwayOfIllusionWithNoTargetsStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 3)
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Sway of Illusion", "Instant", "{1}{U}", oldB654SwayOracle, "CU")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB654QuirionTrailblazerFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB654Seed(me, 3)
	forest := uuid.New()
	me.Library.PushBottom(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Quirion Trailblazer", "Creature — Elf Scout", oldB654TrailOracle, 1, 2)
	effectstest.TfSettle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.TfSettle(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok {
		t.Fatal("the Forest did not enter the battlefield")
	}
	if !c.Tapped {
		t.Error("the Forest entered untapped")
	}
}

func TestOldB654QuirionTrailblazerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := uuid.New()
	me.Library.PushBottom(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Quirion Trailblazer", "Creature — Elf Scout", oldB654TrailOracle, 1, 2)
	effectstest.TfSettle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, forest); ok {
		t.Error("a declined search still put a land onto the battlefield")
	}
}

func TestOldB654TeferisResponseCountersASpellAimedAtYourLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB654Seed(me, 4)
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	rain := oldB654Cast(t, g, opp, "Stone Rain", "Sorcery", "{2}{R}", oldB654StoneRainOracle, "CCR", oldB654Card(land))
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Teferi's Response", "Instant", "{1}{U}", oldB654ResponseOracle, "CU", oldB654Card(rain))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Error("the Forest was destroyed despite the counter")
	}
	if got := effectstest.HandCount(g, me); got != hand+2 {
		t.Errorf("hand %d, want %d (two cards)", got, hand+2)
	}
}

func TestOldB654TeferisResponseDestroysThePermanentWhoseAbilityItCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB654Seed(me, 4)
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	icy := effectstest.PushCatalogPermanent(g, opp.ID, "Icy Manipulator", "Artifact", oldB654IcyOracle, false)
	effectstest.FloatForTest(g, opp, "C")
	if err := g.ActivateCatalogAbility(opp.ID, icy, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB654Card(land)}}); err != nil {
		t.Fatalf("activate Icy: %v", err)
	}
	var ability uuid.UUID
	for id, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemActivated {
			ability = id
		}
	}
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Teferi's Response", "Instant", "{1}{U}", oldB654ResponseOracle, "CU", oldB654Card(ability))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, icy); ok {
		t.Error("the Icy Manipulator survived its countered ability")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, land); tapped {
		t.Error("the Forest was tapped: the ability was not countered")
	}
	if got := effectstest.HandCount(g, me); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}

func TestOldB654TeferisResponseRefusesAnAbilityAtYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	icy := effectstest.PushCatalogPermanent(g, opp.ID, "Icy Manipulator", "Artifact", oldB654IcyOracle, false)
	effectstest.FloatForTest(g, opp, "C")
	if err := g.ActivateCatalogAbility(opp.ID, icy, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB654Card(bear)}}); err != nil {
		t.Fatalf("activate Icy: %v", err)
	}
	var ability uuid.UUID
	for id, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemActivated {
			ability = id
		}
	}
	effectstest.FloatForTest(g, me, "CU")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Teferi's Response", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: oldB654ResponseOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{oldB654Card(ability)}}); err == nil {
		t.Error("Teferi's Response targeted an ability that targets a creature, not a land")
	}
}

func TestOldB654StunStopsABlockAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB654Seed(me, 3)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Stun", "Instant", "{1}{R}", oldB654StunOracle, "CR", oldB654Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if !effectstest.MidB03CantBlock(t, g, bear) {
		t.Error("the creature can still block")
	}
}

func TestOldB654DredgeSacrificesACreatureOrLandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 4)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	other := effectstest.B12Permanent(g, me.ID, "Ring", "Artifact")
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Dredge", "Instant", "{B}", oldB654DredgeOracle, "B")
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, id := range prompt.ChooseCards {
		if id == other {
			t.Error("a noncreature, nonland permanent was offered")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, land); ok {
		t.Error("the Forest was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the Bear was sacrificed")
	}
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB654DredgeDrawsEvenWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB654Seed(me, 4)
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandCount(g, me)
	oldB654Cast(t, g, me, "Dredge", "Instant", "{B}", oldB654DredgeOracle, "B")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}
