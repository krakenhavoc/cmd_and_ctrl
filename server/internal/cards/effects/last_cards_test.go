package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// last_cards_test.go — S79: the last missing cards of Avatar: The Last
// Airbender Eternal (#5075), Avatar: The Last Airbender (#4459) and
// Secrets of Strixhaven Commander (#4834).

const (
	lcTrainingGroundsOracle = "8de3eabe-9a8f-4865-beef-a4888c081d39"
	lcShadrixOracle         = "2143f413-7baa-4fd6-a7ca-32f52bfa553f"
	lcEmptyCityRuseOracle   = "b0391ea4-e37d-48c9-91df-79aefe376bcc"
	lcMirrorEntityOracle    = "17e905ca-c0bd-473d-95a7-e180ba5fea43"
	lcWalkingBallistaOracle = "4b515bb0-f275-4400-8032-3173b799ab40"
	lcPlainCreatureTypeLine = "Creature — Bear"
	lcEnchantmentTypeLine   = "Enchantment"
	lcShadrixTypeLine       = "Legendary Creature — Elder Dragon"
	lcTrainingGroundsName   = "Training Grounds"
	lcShadrixName           = "Shadrix Silverquill"
	lcEmptyCityRuseName     = "Empty City Ruse"
	lcEmptyCityRuseTypeLine = "Sorcery"
	lcWalkingBallistaName   = "Walking Ballista"
	lcWalkingBallistaType   = "Artifact Creature — Construct"
	lcMirrorEntityName      = "Mirror Entity"
	lcMirrorEntityTypeLine  = "Creature — Shapeshifter"
)

func TestLastCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lcTrainingGroundsOracle: lcTrainingGroundsName,
		lcShadrixOracle:         lcShadrixName,
		lcEmptyCityRuseOracle:   lcEmptyCityRuseName,
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

// --- Training Grounds ---------------------------------------------------------

// lcPrice prices a made-up ability of `cost` on the permanent `id` for
// `activator`, at X = x, through the same pricer the activation pays.
func lcPrice(t *testing.T, g *game.Game, activator, id uuid.UUID, cost string, x int) string {
	t.Helper()
	var (
		q   game.ParsedCost
		err error
	)
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		ab := game.ActivatedAbilityShape{Label: cost + ": test", Cost: game.AbilityCost{Mana: cost}}
		q, err = g.AbilityManaCostAtXForEffect(activator, c, game.ZoneBattlefield, ab, nil, x)
	})
	if err != nil {
		t.Fatalf("price %s: %v", cost, err)
	}
	return q.SettleX(x).String()
}

// The rulings' examples, one per row: generic only, never below one
// mana, never raised to one.
func TestLastCardsTrainingGroundsTakesGenericOffButLeavesOneMana(t *testing.T) {
	g, me, _ := spendTable(t)
	bear := b12Creature(g, me.ID, "Bear", lcPlainCreatureTypeLine, 2, 2)
	cases := []struct{ cost, want string }{
		{"{2}{G}", "{G}"},
		{"{1}{G}", "{G}"},
		{"{3}{G}", "{1}{G}"},
		{"{2}", "{1}"},
		{"{4}", "{2}"},
		{"{3}", "{1}"},
		{"{1}", "{1}"},
		{"{R}{R}", "{R}{R}"},
	}
	for _, tc := range cases {
		if got := lcPrice(t, g, me.ID, bear, tc.cost, 0); got != tc.cost {
			t.Errorf("without Training Grounds %s costs %s, want it unchanged", tc.cost, got)
		}
	}
	pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
	for _, tc := range cases {
		if got := lcPrice(t, g, me.ID, bear, tc.cost, 0); got != tc.want {
			t.Errorf("%s under Training Grounds costs %s, want %s", tc.cost, got, tc.want)
		}
	}
}

// Only creatures, and only the controller's: an artifact's ability and
// an opponent's creature's are priced as printed. Two Training Grounds
// take {4} off a {5} but still leave one mana.
func TestLastCardsTrainingGroundsOnlyYourCreaturesAndStacksToTheFloor(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Bear", lcPlainCreatureTypeLine, 2, 2)
	mine := b12Creature(g, me.ID, "Bear", lcPlainCreatureTypeLine, 2, 2)
	if got := lcPrice(t, g, me.ID, rock, "{3}", 0); got != "{3}" {
		t.Errorf("an artifact's {3} costs %s, want {3}", got)
	}
	if got := lcPrice(t, g, opp.ID, theirs, "{3}", 0); got != "{3}" {
		t.Errorf("an opponent's creature's {3} costs %s, want {3}", got)
	}
	if got := lcPrice(t, g, me.ID, mine, "{5}", 0); got != "{3}" {
		t.Errorf("one Training Grounds: {5} costs %s, want {3}", got)
	}
	pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
	if got := lcPrice(t, g, me.ID, mine, "{5}", 0); got != "{1}" {
		t.Errorf("two Training Grounds: {5} costs %s, want {1}", got)
	}
	if got := lcPrice(t, g, me.ID, mine, "{4}{U}", 0); got != "{U}" {
		t.Errorf("two Training Grounds: {4}{U} costs %s, want {U}", got)
	}
}

// {X} counts at its announced value: {X}{B}{B} at X = 5 is {3}{B}{B}
// (the Drana ruling), and an {X} alone keeps one mana, so X = 3 is {1}
// and X = 1 is still {1}.
func TestLastCardsTrainingGroundsReadsTheAnnouncedX(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
	bear := b12Creature(g, me.ID, "Bear", lcPlainCreatureTypeLine, 2, 2)
	for _, tc := range []struct {
		cost string
		x    int
		want string
	}{
		{"{X}{B}{B}", 5, "{3}{B}{B}"},
		{"{X}", 3, "{1}"},
		{"{X}", 1, "{1}"},
		{"{X}", 0, ""},
		{"{X}{1}", 1, "{1}"},
	} {
		if got := lcPrice(t, g, me.ID, bear, tc.cost, tc.x); got != tc.want {
			t.Errorf("%s at X = %d under Training Grounds costs %s, want %s", tc.cost, tc.x, got, tc.want)
		}
	}
}

// Engine-driven: Walking Ballista's {4} ability is paid from a pool of
// exactly two under Training Grounds, and refused without it.
func TestLastCardsTrainingGroundsDiscountsARealActivation(t *testing.T) {
	for _, with := range []bool{false, true} {
		g, me, _ := spendTable(t)
		if with {
			pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
		}
		ballista := pushCatalogPermanent(g, me.ID, lcWalkingBallistaName, lcWalkingBallistaType, lcWalkingBallistaOracle, false)
		floatForTest(g, me, "CC")
		err := g.ActivateCatalogAbility(me.ID, ballista, 0, game.ActivateAbilityParams{Strict: true})
		if with && err != nil {
			t.Fatalf("Walking Ballista's {4} under Training Grounds from {C}{C}: %v", err)
		}
		if !with && err == nil {
			t.Fatal("Walking Ballista's {4} was paid with {C}{C} and no Training Grounds")
		}
		if with {
			if me.ManaPool.Total() != 0 {
				t.Errorf("%d mana left, want 0", me.ManaPool.Total())
			}
			passPriorityAroundTable(t, g)
			if n := countersOn(g, ballista, game.CounterPlusOne); n != 1 {
				t.Errorf("Walking Ballista has %d +1/+1 counters, want 1", n)
			}
		}
	}
}

// Engine-driven with X: Mirror Entity's {X} at X = 3 is paid with one
// mana, and at X = 1 still costs one.
func TestLastCardsTrainingGroundsMirrorEntityPaysXLessTwo(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, lcTrainingGroundsName, lcEnchantmentTypeLine, lcTrainingGroundsOracle, false)
	entity := pushCatalogPermanent(g, me.ID, lcMirrorEntityName, lcMirrorEntityTypeLine, lcMirrorEntityOracle, false)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, entity, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("Mirror Entity at X = 3 from {C}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left after X = 3, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, entity); p != 3 || tg != 3 {
		t.Errorf("Mirror Entity is %d/%d, want 3/3", p, tg)
	}
	if err := g.ActivateCatalogAbility(me.ID, entity, 0, game.ActivateAbilityParams{Strict: true, XValue: 1}); err == nil {
		t.Error("Mirror Entity at X = 1 was activated for nothing")
	}
}

// Register refuses the floor anywhere it means nothing.
func TestLastCardsKeepsOneManaRefusedOnTheWrongShapes(t *testing.T) {
	for name, m := range map[string]game.CostModifier{
		"an increase":       KeepsOneMana(CostsMore(1, "x")),
		"a unit reduction":  KeepsOneMana(ActivationCostsLessMana("{1}{U}", "x")),
		"a cost floor":      KeepsOneMana(CostsAtLeast(3, "x")),
		"a plain reduction": ActivationCostsLess(2, "x"),
	} {
		got := m.KeepsOneManaProblem()
		if name == "a plain reduction" {
			if got != "" {
				t.Errorf("%s: %q, want no problem", name, got)
			}
			continue
		}
		if got == "" {
			t.Errorf("%s with a one-mana floor was accepted", name)
		}
	}
}

// --- Shadrix Silverquill -------------------------------------------------------

// lcShadrixTrigger walks to my beginning of combat with Shadrix out and
// answers its "you may choose two" with `yes`.
func lcShadrixTrigger(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	pushCatalogPermanent(g, me.ID, lcShadrixName, lcShadrixTypeLine, lcShadrixOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); c != nil {
		t.Fatal("the modes were asked before the \"you may\"")
	}
	p := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if p == nil {
		t.Fatalf("Shadrix asked no \"you may\" at the beginning of combat: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(p.ID, me.ID, yes); err != nil {
		t.Fatalf("answer the \"you may\": %v", err)
	}
}

func lcPickPlayer(g *game.Game, chooser, player uuid.UUID) error {
	p := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, chooser)
	if p == nil {
		return errNoPrompt
	}
	return g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player})
}

type lcError string

func (e lcError) Error() string { return string(e) }

const errNoPrompt = lcError("no target prompt")

func lcCreatures(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// Exactly two modes, two different players: the Inkling for me, the
// draw and the life loss for an opponent.
func TestLastCardsShadrixChoosesTwoModesForTwoDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lcShadrixTrigger(t, g, me, true)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode pick after \"yes\": %+v", g.PendingChoices)
	}
	if pick.ModeMin != 2 || pick.ModeMax != 2 {
		t.Fatalf("choose two: min %d, max %d", pick.ModeMin, pick.ModeMax)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err == nil {
		t.Fatal("one mode was accepted")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("modes 0 and 1: %v", err)
	}
	if err := lcPickPlayer(g, me.ID, me.ID); err != nil {
		t.Fatalf("the Inkling for me: %v", err)
	}
	if err := lcPickPlayer(g, me.ID, me.ID); err == nil {
		t.Fatal("the second mode targeted the player the first one did")
	}
	if err := lcPickPlayer(g, me.ID, opp.ID); err != nil {
		t.Fatalf("the draw for an opponent: %v", err)
	}
	hand, life := opp.Hand.Size(), opp.Life
	myLife := me.Life
	passPriorityAroundTable(t, g)
	if n := len(lcCreatures(g, me.ID, "Inkling")); n != 1 {
		t.Errorf("I have %d Inklings, want 1", n)
	}
	if n := len(lcCreatures(g, opp.ID, "Inkling")); n != 0 {
		t.Errorf("the opponent has %d Inklings, want 0", n)
	}
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("the opponent's hand %d, want %d", got, hand+1)
	}
	if opp.Life != life-1 || me.Life != myLife {
		t.Errorf("life: opponent %d (want %d), me %d (want %d)", opp.Life, life-1, me.Life, myLife)
	}
}

// The counters bullet puts one on each creature the target player
// controls, and on nobody else's.
func TestLastCardsShadrixCountersOnEachCreatureOfTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Bear A", lcPlainCreatureTypeLine, 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", lcPlainCreatureTypeLine, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", lcPlainCreatureTypeLine, 2, 2)
	lcShadrixTrigger(t, g, me, true)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode pick")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("modes 1 and 2: %v", err)
	}
	if err := lcPickPlayer(g, me.ID, me.ID); err != nil {
		t.Fatalf("the draw for me: %v", err)
	}
	if err := lcPickPlayer(g, me.ID, opp.ID); err != nil {
		t.Fatalf("the counters for an opponent: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("the opponent's creature has %d counters, want 1", n)
		}
	}
	if n := countersOn(g, mine, game.CounterPlusOne); n != 0 {
		t.Errorf("my creature has %d counters, want 0", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("my hand %d, want %d", got, hand+1)
	}
}

// "No" is choosing no mode: nothing is asked and nothing goes on the
// stack (CR 700.2b).
func TestLastCardsShadrixDecliningChoosesNoMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lcShadrixTrigger(t, g, me, false)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); c != nil {
		t.Fatal("modes were asked after \"no\"")
	}
	if g.Stack.Size() != 0 {
		t.Errorf("the stack holds %d items after \"no\", want 0", g.Stack.Size())
	}
}

// Only on its controller's turn.
func TestLastCardsShadrixOnlyOnYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, lcShadrixName, lcShadrixTypeLine, lcShadrixOracle, false)
	endTurn(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID); c != nil {
		t.Error("Shadrix asked at an opponent's beginning of combat")
	}
	if n := triggersOnStackFrom(g, id); n != 0 {
		t.Errorf("%d triggers at an opponent's beginning of combat", n)
	}
}

// --- Empty City Ruse ----------------------------------------------------------

// lcStepsOfTurn walks from the start of `seat`'s next turn through to
// its end and reports whether any combat step began in it and whether
// its postcombat main phase did.
func lcStepsOfTurn(t *testing.T, g *game.Game, seat int) (combat, postcombatMain bool) {
	t.Helper()
	for i := 0; i < 400 && (g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to seat %d: %v", seat, err)
		}
	}
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		switch game.PhaseOf(g.Turn.Step) {
		case game.PhaseCombat:
			combat = true
		case game.PhasePostcombatMain:
			postcombatMain = true
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep in seat %d's turn: %v", seat, err)
		}
	}
	return combat, postcombatMain
}

// The opponent's next turn has no combat phase and keeps its second
// main phase; the turn after it and the opponent's turn after that keep
// theirs.
func TestLastCardsEmptyCityRuseSkipsTheOpponentsNextCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	castCatalogSpell(t, g, lcEmptyCityRuseName, lcEmptyCityRuseTypeLine, lcEmptyCityRuseOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if combat, main2 := lcStepsOfTurn(t, g, 1); combat || !main2 {
		t.Fatalf("the opponent's next turn: combat %v, postcombat main %v; want no combat and a second main phase", combat, main2)
	}
	if combat, _ := lcStepsOfTurn(t, g, 2); !combat {
		t.Error("the next player's turn lost its combat")
	}
	if combat, _ := lcStepsOfTurn(t, g, 1); !combat {
		t.Error("the opponent's turn after that had no combat")
	}
}

// "Target opponent": the caster can't be the target.
func TestLastCardsEmptyCityRuseTargetsAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: lcEmptyCityRuseName, TypeLine: lcEmptyCityRuseTypeLine,
			OracleID: lcEmptyCityRuseOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("Empty City Ruse targeted its caster")
	}
}
