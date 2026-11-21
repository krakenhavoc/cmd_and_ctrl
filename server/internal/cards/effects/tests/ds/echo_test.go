package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The first upkeep after the permanent arrived charges the echo; paying
// keeps it, and the upkeep after that charges nothing.
func TestEchoChargesTheFirstUpkeepOnlyAndPayingKeepsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	patrol := effectstest.CastEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", effectstest.OracleGoblinPatrol)

	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	if effectstest.TriggerOnStack(g, patrol) == nil {
		t.Fatal("echo did not trigger at the first upkeep after Goblin Patrol entered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no echo prompt after the trigger resolved")
	}
	if prompt.PayCost != "{R}" {
		t.Fatalf("echo cost = %q, want {R}", prompt.PayCost)
	}
	if !g.ChoicePromptBlocksTable(prompt) {
		t.Error("the echo prompt does not hold the upkeep (#997)")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	before := len(g.Events)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay echo: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, patrol); !ok {
		t.Fatal("a paid echo sacrificed Goblin Patrol")
	}
	paid := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventEchoPaid && ev.CardID == patrol && ev.Actor == me.ID {
			paid = true
		}
	}
	if !paid {
		t.Error("paying the echo emitted no EventEchoPaid")
	}

	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	if effectstest.TriggerOnStack(g, patrol) != nil {
		t.Fatal("echo triggered again at the second upkeep (CR 702.30a: only since your last upkeep)")
	}
}

// Declining sacrifices it, and so does a "yes" with no mana to pay.
func TestEchoSacrificesWhenNotPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		patrol := effectstest.CastEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", effectstest.OracleGoblinPatrol)
		effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.EchoPrompt(g, me.ID)
		if prompt == nil {
			t.Fatal("no echo prompt")
		}
		if err := g.ResolvePayUnless(prompt.ID, me.ID, pay); err != nil {
			t.Fatalf("answer echo: %v", err)
		}
		if _, ok := effectstest.BattlefieldCard(g, patrol); ok {
			t.Errorf("pay=%v without mana: Goblin Patrol survived its unpaid echo", pay)
		}
	}
}

// A permanent that came under your control another way — a control
// change — is charged at your next upkeep (CR 702.30a: "came under your
// control"), and its echo is charged to its new controller.
func TestEchoChargesAfterAControlChange(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	patrol := effectstest.CastEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", effectstest.OracleGoblinPatrol)
	// Seat 0 pays its echo at the next upkeep, so the Patrol is settled
	// with seat 0 by the time seat 1 takes it.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ResolvePayUnless(effectstest.EchoPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("pay echo: %v", err)
	}
	if !g.GainControlForEffect(uuid.New(), patrol, them.ID, game.Duration{Kind: game.Indefinite}, "test theft") {
		t.Fatal("GainControlForEffect refused")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	c, _ := effectstest.BattlefieldCard(g, patrol)
	if c.Controller != them.ID {
		t.Fatalf("the theft did not land: controller %s", c.Controller)
	}
	if effectstest.TriggerOnStack(g, patrol) == nil {
		t.Fatal("echo did not trigger for the player who gained control of it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EchoPrompt(g, them.ID); p == nil {
		t.Fatal("the new controller was not asked to pay the echo")
	}
}

// CR 603.4: a permanent that changed controller after the trigger and
// before it resolved is not charged — "sacrifice it" is about a
// permanent you control (CR 701.21a).
func TestEchoChecksItsConditionAgainOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	patrol := effectstest.CastEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", effectstest.OracleGoblinPatrol)
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	if effectstest.TriggerOnStack(g, patrol) == nil {
		t.Fatal("no echo trigger")
	}
	g.GainControlForEffect(uuid.New(), patrol, them.ID, game.Duration{Kind: game.Indefinite}, "test theft")
	g.RecomputeLayersIfStaleLocked()
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EchoPrompt(g, me.ID); p != nil {
		t.Fatal("the echo asked its controller to pay for a permanent they no longer control")
	}
	if _, ok := effectstest.BattlefieldCard(g, patrol); !ok {
		t.Fatal("an echo whose condition failed on resolution still sacrificed the permanent")
	}
}

// Deepcavern Imp: "Echo—Discard a card." The payer names the card; the
// wrong count or a card not in hand is refused with the prompt left
// open; an empty hand can't pay and "yes" is a decline.
func TestEchoDiscardPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	imp := effectstest.CastEchoCreature(t, g, "Deepcavern Imp", "Creature — Imp Rebel", effectstest.OracleDeepcavernImp)
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no echo prompt")
	}
	action := prompt.PayAction()
	if action == nil || action.Kind != game.PayActionDiscard || action.Count != 1 {
		t.Fatalf("pay action = %+v, want discard 1", action)
	}
	if prompt.PayCost != "Discard a card" {
		t.Errorf("PayCost = %q, want %q", prompt.PayCost, "Discard a card")
	}
	if me.Hand.Size() < 2 {
		effectstest.PushHandCard(g, me)
		effectstest.PushHandCard(g, me)
	}
	two := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, two); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two cards for a one-card discard: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{imp}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a battlefield card as the discard: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, false, two[:1]); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cards on a decline: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a payable discard answered with no card: %v, want ErrInvalidParam", err)
	}
	if effectstest.EchoPrompt(g, me.ID) == nil {
		t.Fatal("a refused answer consumed the prompt")
	}
	handBefore := me.Hand.Size()
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, two[:1]); err != nil {
		t.Fatalf("discard to pay echo: %v", err)
	}
	if me.Hand.Size() != handBefore-1 || me.Hand.Contains(two[0]) {
		t.Fatal("the discard was not made")
	}
	if me.Graveyard == nil || !me.Graveyard.Contains(two[0]) {
		t.Fatal("the discarded card is not in the graveyard")
	}
	if _, ok := effectstest.BattlefieldCard(g, imp); !ok {
		t.Fatal("a paid discard echo sacrificed Deepcavern Imp")
	}
}

func TestEchoDiscardWithAnEmptyHandIsADecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	imp := effectstest.CastEchoCreature(t, g, "Deepcavern Imp", "Creature — Imp Rebel", effectstest.OracleDeepcavernImp)
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	prompt := effectstest.EchoPrompt(g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("yes with nothing to discard: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, imp); ok {
		t.Fatal("Deepcavern Imp survived an echo its controller could not pay")
	}
}

// Skizzik Surger: "Echo—Sacrifice two lands." Two lands the payer
// controls; a non-land or an opponent's land is refused.
func TestEchoSacrificePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	surger := effectstest.CastEchoCreature(t, g, "Skizzik Surger", "Creature — Elemental", effectstest.OracleSkizzikSurger)
	a := effectstest.PushLand(g, me.ID, "Island A")
	b := effectstest.PushLand(g, me.ID, "Island B")
	theirs := effectstest.PushLand(g, them.ID, "Their Island")
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no echo prompt")
	}
	if prompt.PayCost != "Sacrifice two lands" {
		t.Errorf("PayCost = %q", prompt.PayCost)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{a, theirs}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("an opponent's land: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{a, surger}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a creature as a land: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{a, a}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("one land named twice: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("sacrifice two lands: %v", err)
	}
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("a land paid toward the echo is still on the battlefield")
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, surger); !ok {
		t.Fatal("a paid echo sacrificed Skizzik Surger")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
		t.Fatal("the opponent's land left the battlefield")
	}
}

// Polar Kraken: "Cumulative upkeep—Sacrifice a land." Two age counters
// is two lands, all at once (CR 702.24a).
func TestCumulativeUpkeepSacrificeScalesWithAgeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	kraken := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Polar Kraken", OracleID: effectstest.OraclePolarKraken, TypeLine: "Creature — Kraken", Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID})
	var lands []uuid.UUID
	for i := 0; i < 4; i++ {
		lands = append(lands, effectstest.PushLand(g, me.ID, "Island"))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil || prompt.PayAction() == nil || prompt.PayAction().Count != 1 {
		t.Fatalf("first upkeep: prompt %+v, want sacrifice 1", prompt)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, lands[:1]); err != nil {
		t.Fatalf("pay one land: %v", err)
	}
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("second upkeep: no cumulative upkeep prompt")
	}
	if prompt.PayAction().Count != 2 || prompt.PayCost != "Sacrifice two lands" {
		t.Fatalf("second upkeep: prompt %q %+v, want sacrifice two lands", prompt.PayCost, prompt.PayAction())
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, lands[1:2]); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a partial payment: %v, want ErrInvalidParam (CR 702.24a)", err)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, lands[1:3]); err != nil {
		t.Fatalf("pay two lands: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, kraken); !ok {
		t.Fatal("a paid cumulative upkeep sacrificed Polar Kraken")
	}
}

// EchoX reads its amount as the trigger resolves (the Volcano Hellion
// ruling: "equal to your life total as the echo triggered ability
// resolves"), and a negative amount is zero (CR 107.1b).
func TestEchoXReadsTheAmountOnResolution(t *testing.T) {
	const oid = "test-echo-x-life-total"
	// "X is your life total minus three", so a small life total makes the
	// amount negative without the player losing the game.
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oid, Name: "Test Hellion", Triggered: []game.TriggeredAbility{
		effects.EchoX("Test Hellion", "your life total minus three", func(ctx *effects.Context) int {
			return ctx.Game.PlayerByIDForEffect(ctx.Controller()).Life - 3
		}),
	}})
	for _, tc := range []struct {
		life int
		want string
	}{{40, "{37}"}, {1, "{0}"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Hellion", OracleID: oid,
			TypeLine: "Creature — Hellion", Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		// The trigger is on the stack; the life total changes before it
		// resolves, and the cost is read from the new one.
		me.Life = tc.life
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.EchoPrompt(g, me.ID)
		if p == nil {
			t.Fatalf("life %d: no echo prompt", tc.life)
		}
		if p.PayCost != tc.want {
			t.Errorf("life %d: echo cost %q, want %q", tc.life, p.PayCost, tc.want)
		}
	}
}

// Phyrexian Soulgorger may be sacrificed to its own cumulative upkeep:
// it is a creature, and the cost is "Sacrifice a creature".
func TestCumulativeUpkeepSoulgorgerCountsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	gorger := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Phyrexian Soulgorger", OracleID: effectstest.OraclePhyrexianSoulgorg, TypeLine: "Snow Artifact Creature — Phyrexian Construct", Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no cumulative upkeep prompt")
	}
	opts := g.PayActionOptionsForEffect(me.ID, prompt.PayAction())
	if len(opts) != 1 || opts[0] != gorger {
		t.Fatalf("options = %v, want just the Soulgorger", opts)
	}
}
