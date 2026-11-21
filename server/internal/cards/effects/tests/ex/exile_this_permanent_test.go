package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_this_permanent_test.go — #1404's proof cards: "Exile this
// <permanent>:" paid from the battlefield. The engine half is
// game/exile_self_battlefield_test.go.

const (
	hangedExecutionerOracle = "6ff4ff67-ad08-447f-a112-1a071c1474a4"
	nyxWeaverOracle         = "6f8bf968-0571-4582-8c04-9791e44d5df0"
	feldonsCaneOracle       = "9b884dfd-59f4-45c0-bf1e-6ad9f5b58895"
)

// Perpetual Timepiece's second ability: the artifact is in exile the
// moment the activation returns — before the ability resolves — and on
// resolution the named graveyard cards go into the library while the
// unnamed one stays.
func TestPerpetualTimepieceExilesItselfAndShufflesTheTargetsIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	piece := effectstest.B12Push(g, me.ID, "Perpetual Timepiece", "Artifact", effectstest.B18PerpetualTimepieceOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushGraveyardCard(g, me, "Named A")
	b := effectstest.PushGraveyardCard(g, me, "Named B")
	keep := effectstest.PushGraveyardCard(g, me, "Left Behind")
	effectstest.B06AddMana(me, "C", "C")
	lib := me.Library.Size()

	err := g.ActivateCatalogAbility(me.ID, piece, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a},
		{Kind: game.TargetCard, ID: b},
	}})
	if err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(piece) || !g.Exile.Contains(piece) {
		t.Fatal("the Timepiece is not in exile once its cost is paid")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Library.Contains(a) || !me.Library.Contains(b) {
		t.Error("the two targets were not shuffled into the library")
	}
	if !me.Graveyard.Contains(keep) {
		t.Error("an untargeted card left the graveyard")
	}
	if got := me.Library.Size(); got != lib+2 {
		t.Errorf("library %d → %d, want +2", lib, got)
	}
	if me.Graveyard.Contains(piece) || me.Library.Contains(piece) {
		t.Error("the Timepiece is not in exile after resolution")
	}
}

// "Any number" includes none: the activation with no targets is legal
// and still exiles the Timepiece.
func TestPerpetualTimepieceWithNoTargetsStillActivates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	piece := effectstest.B12Push(g, me.ID, "Perpetual Timepiece", "Artifact", effectstest.B18PerpetualTimepieceOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, piece, 1, game.ActivateAbilityParams{})
	if !g.Exile.Contains(piece) {
		t.Error("the Timepiece was not exiled")
	}
}

// The #1404 contract, on a creature where it matters: exiling Hanged
// Executioner as a cost fires a leaves-the-battlefield watcher and
// neither a dies watcher nor a sacrifice watcher. The LTB trigger is
// on the stack ABOVE the ability before anyone passes (CR 603.3b —
// the cost was paid before the ability was put on the stack).
func TestHangedExecutionerLeavesTheBattlefieldWithoutDying(t *testing.T) {
	const (
		ltbOracle  = "test-1404-ltb-watcher"
		diesOracle = "test-1404-dies-watcher"
		sacOracle  = "test-1404-sacrifice-watcher"
	)
	yoursLeft := func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && ev.CardID != source.InstanceID && c.Owner == source.Controller && c.IsCreature()
	}
	effectstest.RegisterForTest(t, effects.Spec{OracleID: ltbOracle, Name: "LTB Watcher", Triggered: []game.TriggeredAbility{
		effects.On(game.EventLTB, yoursLeft, "LTB Watcher — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
	}})
	effectstest.RegisterForTest(t, effects.Spec{OracleID: diesOracle, Name: "Dies Watcher", Triggered: []game.TriggeredAbility{
		effects.On(game.EventLTB, effects.ACreatureYouControlDied, "Dies Watcher — gain 10 life", effects.Do(effects.GainLife{Amount: 10})),
	}})
	effectstest.RegisterForTest(t, effects.Spec{OracleID: sacOracle, Name: "Sacrifice Watcher", Triggered: []game.TriggeredAbility{
		effects.On(game.EventSacrifice, effects.ByYou, "Sacrifice Watcher — gain 100 life", effects.Do(effects.GainLife{Amount: 100})),
	}})

	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ltb := effectstest.B12Push(g, me.ID, "LTB Watcher", "Enchantment", ltbOracle, 0, 0)
	effectstest.B12Push(g, me.ID, "Dies Watcher", "Enchantment", diesOracle, 0, 0)
	effectstest.B12Push(g, me.ID, "Sacrifice Watcher", "Enchantment", sacOracle, 0, 0)
	exec := effectstest.B12Push(g, me.ID, "Hanged Executioner", "Creature — Spirit", hangedExecutionerOracle, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, exec, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
	}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if !g.Exile.Contains(exec) || me.Graveyard.Contains(exec) {
		t.Fatal("the Executioner is not in exile once its cost is paid")
	}
	trig := effectstest.TriggerOnStack(g, ltb)
	if trig == nil {
		t.Fatal("the leaves-the-battlefield watcher did not trigger off the cost")
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.Seq > trig.Seq {
			t.Error("the ability is above its cost's LTB trigger — the cost is paid before the ability is on the stack")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the target creature was not exiled")
	}
	if got := me.Life - life; got != 1 {
		t.Errorf("life %+d, want +1: the LTB watcher once, and no dies (+10) or sacrifice (+100) watcher", got)
	}
}

// The ETB half and the keyword: two flyers for three.
func TestHangedExecutionerMakesAFlyingSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	exec := effectstest.CastAndResolveCreature(t, g, "Hanged Executioner", "Creature — Spirit", hangedExecutionerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Spirit"); n != 1 {
		t.Fatalf("%d Spirit tokens, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, exec, "flying") {
		t.Error("the Executioner has flying")
	}
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.FindBattlefieldByName(g, "Spirit"), "flying") {
		t.Error("the Spirit token has flying")
	}
}

// Nyx Weaver: the upkeep mill, then the exile that buys back a card.
// The Weaver goes to exile, never the graveyard, so it can never be
// the card it returns.
func TestNyxWeaverMillsTwoThenExilesItselfToRegrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	weaver := effectstest.B12Push(g, me.ID, "Nyx Weaver", "Enchantment Creature — Spider", nyxWeaverOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, weaver, "reach") {
		t.Error("Nyx Weaver has reach")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	grave := me.Graveyard.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != grave+2 {
		t.Fatalf("graveyard %d → %d at upkeep, want two milled", grave, got)
	}

	effectstest.AdvanceToMain(t, g)
	want := me.Graveyard.Cards[0].InstanceID
	effectstest.B06AddMana(me, "B", "G", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, weaver, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: want},
	}})
	if !me.Hand.Contains(want) || me.Hand.Size() != hand+1 {
		t.Error("the targeted card did not return to hand")
	}
	if !g.Exile.Contains(weaver) || me.Graveyard.Contains(weaver) {
		t.Error("the Weaver is not in exile")
	}
}

// An illegal activation leaves the board untouched: a Nyx Weaver
// pointed at a card in ANOTHER player's graveyard is refused at
// announce (CR 601.2c), before the exile is paid.
func TestNyxWeaverRefusedActivationStaysOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weaver := effectstest.B12Push(g, me.ID, "Nyx Weaver", "Enchantment Creature — Spider", nyxWeaverOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushGraveyardCard(g, opp, "Their Card")
	effectstest.B06AddMana(me, "B", "G", "C")

	err := g.ActivateCatalogAbility(me.ID, weaver, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs},
	}})
	if err == nil {
		t.Fatal("a target in an opponent's graveyard was accepted")
	}
	if !g.Battlefield.Contains(weaver) || g.Exile.Contains(weaver) {
		t.Error("a refused activation exiled the Weaver")
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %v, want the mana untouched", me.ManaPool)
	}
}

// Feldon's Cane: {T} and the exile are one cost, the Cane is gone
// before the graveyard is reset — so it is not shuffled in with it.
func TestFeldonsCaneShufflesTheGraveyardInWithoutItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cane := effectstest.B12Push(g, me.ID, "Feldon's Cane", "Artifact", feldonsCaneOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	var milled []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		milled = append(milled, effectstest.PushGraveyardCard(g, me, n))
	}
	lib := me.Library.Size()

	effectstest.B16Activate(t, g, me.ID, cane, 0, game.ActivateAbilityParams{})

	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard holds %d cards, want 0", me.Graveyard.Size())
	}
	for _, id := range milled {
		if !me.Library.Contains(id) {
			t.Error("a graveyard card was not shuffled into the library")
		}
	}
	if got := me.Library.Size(); got != lib+3 {
		t.Errorf("library %d → %d, want +3 (the Cane is not one of them)", lib, got)
	}
	if !g.Exile.Contains(cane) {
		t.Error("the Cane is not in exile")
	}
}
