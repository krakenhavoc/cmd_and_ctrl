package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// echo_test.go — ADR 0108 §5 (#1888), CR 702.30a: "At the beginning of
// your upkeep, if this permanent came under your control since the
// beginning of your last upkeep, sacrifice it unless you pay [cost]."

const (
	oracleGoblinPatrol      = "edf25cbe-fa28-44ca-b4be-1d2312e4f6ac"
	oracleDeepcavernImp     = "1f295f2f-969a-4b82-98d1-7fe307ec83a7"
	oracleSkizzikSurger     = "ded349de-0599-4744-ad3e-aec95caf9f99"
	oraclePolarKraken       = "d0ff30c8-ddb7-439d-b14d-c83fe3c8da87"
	oraclePhyrexianSoulgorg = "b6e36e77-0cea-4b7a-b968-774b1e74a992"
)

// echoPrompt is the open pay-unless prompt `chooser` owes, or nil.
func echoPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// castEchoCreature casts a catalogued creature for the active seat and
// resolves it, returning the permanent.
func castEchoCreature(t *testing.T, g *game.Game, name, typeLine, oracleID string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, typeLine, oracleID, nil)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatalf("%s did not reach the battlefield", name)
	}
	return id
}

// leaveUpkeepThenAdvanceTo walks off the step the cursor is on and on to
// `seat`'s next upkeep.
func leaveUpkeepThenAdvanceTo(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, seat)
}

// The first upkeep after the permanent arrived charges the echo; paying
// keeps it, and the upkeep after that charges nothing.
func TestEchoChargesTheFirstUpkeepOnlyAndPayingKeepsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	patrol := castEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", oracleGoblinPatrol)

	leaveUpkeepThenAdvanceTo(t, g, 0)
	if triggerOnStack(g, patrol) == nil {
		t.Fatal("echo did not trigger at the first upkeep after Goblin Patrol entered")
	}
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
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
	if _, ok := battlefieldCard(g, patrol); !ok {
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

	leaveUpkeepThenAdvanceTo(t, g, 0)
	if triggerOnStack(g, patrol) != nil {
		t.Fatal("echo triggered again at the second upkeep (CR 702.30a: only since your last upkeep)")
	}
}

// Declining sacrifices it, and so does a "yes" with no mana to pay.
func TestEchoSacrificesWhenNotPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		patrol := castEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", oracleGoblinPatrol)
		leaveUpkeepThenAdvanceTo(t, g, 0)
		passPriorityAroundTable(t, g)
		prompt := echoPrompt(g, me.ID)
		if prompt == nil {
			t.Fatal("no echo prompt")
		}
		if err := g.ResolvePayUnless(prompt.ID, me.ID, pay); err != nil {
			t.Fatalf("answer echo: %v", err)
		}
		if _, ok := battlefieldCard(g, patrol); ok {
			t.Errorf("pay=%v without mana: Goblin Patrol survived its unpaid echo", pay)
		}
	}
}

// A permanent that came under your control another way — a control
// change — is charged at your next upkeep (CR 702.30a: "came under your
// control"), and its echo is charged to its new controller.
func TestEchoChargesAfterAControlChange(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	patrol := castEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", oracleGoblinPatrol)
	// Seat 0 pays its echo at the next upkeep, so the Patrol is settled
	// with seat 0 by the time seat 1 takes it.
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ResolvePayUnless(echoPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("pay echo: %v", err)
	}
	if !g.GainControlForEffect(uuid.New(), patrol, them.ID, game.Duration{Kind: game.Indefinite}, "test theft") {
		t.Fatal("GainControlForEffect refused")
	}
	advanceToUpkeepOf(t, g, 1)
	c, _ := battlefieldCard(g, patrol)
	if c.Controller != them.ID {
		t.Fatalf("the theft did not land: controller %s", c.Controller)
	}
	if triggerOnStack(g, patrol) == nil {
		t.Fatal("echo did not trigger for the player who gained control of it")
	}
	passPriorityAroundTable(t, g)
	if p := echoPrompt(g, them.ID); p == nil {
		t.Fatal("the new controller was not asked to pay the echo")
	}
}

// CR 603.4: a permanent that changed controller after the trigger and
// before it resolved is not charged — "sacrifice it" is about a
// permanent you control (CR 701.21a).
func TestEchoChecksItsConditionAgainOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	patrol := castEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", oracleGoblinPatrol)
	leaveUpkeepThenAdvanceTo(t, g, 0)
	if triggerOnStack(g, patrol) == nil {
		t.Fatal("no echo trigger")
	}
	g.GainControlForEffect(uuid.New(), patrol, them.ID, game.Duration{Kind: game.Indefinite}, "test theft")
	g.RecomputeLayersIfStaleLocked()
	passPriorityAroundTable(t, g)
	if p := echoPrompt(g, me.ID); p != nil {
		t.Fatal("the echo asked its controller to pay for a permanent they no longer control")
	}
	if _, ok := battlefieldCard(g, patrol); !ok {
		t.Fatal("an echo whose condition failed on resolution still sacrificed the permanent")
	}
}

// Deepcavern Imp: "Echo—Discard a card." The payer names the card; the
// wrong count or a card not in hand is refused with the prompt left
// open; an empty hand can't pay and "yes" is a decline.
func TestEchoDiscardPayment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	imp := castEchoCreature(t, g, "Deepcavern Imp", "Creature — Imp Rebel", oracleDeepcavernImp)
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
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
		pushHandCard(g, me)
		pushHandCard(g, me)
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
	if echoPrompt(g, me.ID) == nil {
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
	if _, ok := battlefieldCard(g, imp); !ok {
		t.Fatal("a paid discard echo sacrificed Deepcavern Imp")
	}
}

func TestEchoDiscardWithAnEmptyHandIsADecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	imp := castEchoCreature(t, g, "Deepcavern Imp", "Creature — Imp Rebel", oracleDeepcavernImp)
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	prompt := echoPrompt(g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("yes with nothing to discard: %v", err)
	}
	if _, ok := battlefieldCard(g, imp); ok {
		t.Fatal("Deepcavern Imp survived an echo its controller could not pay")
	}
}

// Skizzik Surger: "Echo—Sacrifice two lands." Two lands the payer
// controls; a non-land or an opponent's land is refused.
func TestEchoSacrificePayment(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	surger := castEchoCreature(t, g, "Skizzik Surger", "Creature — Elemental", oracleSkizzikSurger)
	a := pushLand(g, me.ID, "Island A")
	b := pushLand(g, me.ID, "Island B")
	theirs := pushLand(g, them.ID, "Their Island")
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
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
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("a land paid toward the echo is still on the battlefield")
		}
	}
	if _, ok := battlefieldCard(g, surger); !ok {
		t.Fatal("a paid echo sacrificed Skizzik Surger")
	}
	if _, ok := battlefieldCard(g, theirs); !ok {
		t.Fatal("the opponent's land left the battlefield")
	}
}

// Polar Kraken: "Cumulative upkeep—Sacrifice a land." Two age counters
// is two lands, all at once (CR 702.24a).
func TestCumulativeUpkeepSacrificeScalesWithAgeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	kraken := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Polar Kraken", OracleID: oraclePolarKraken, TypeLine: "Creature — Kraken", Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID})
	var lands []uuid.UUID
	for i := 0; i < 4; i++ {
		lands = append(lands, pushLand(g, me.ID, "Island"))
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
	if prompt == nil || prompt.PayAction() == nil || prompt.PayAction().Count != 1 {
		t.Fatalf("first upkeep: prompt %+v, want sacrifice 1", prompt)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, lands[:1]); err != nil {
		t.Fatalf("pay one land: %v", err)
	}
	leaveUpkeepThenAdvanceTo(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt = echoPrompt(g, me.ID)
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
	if _, ok := battlefieldCard(g, kraken); !ok {
		t.Fatal("a paid cumulative upkeep sacrificed Polar Kraken")
	}
}

// Phyrexian Soulgorger may be sacrificed to its own cumulative upkeep:
// it is a creature, and the cost is "Sacrifice a creature".
func TestCumulativeUpkeepSoulgorgerCountsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	gorger := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Phyrexian Soulgorger", OracleID: oraclePhyrexianSoulgorg, TypeLine: "Snow Artifact Creature — Phyrexian Construct", Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no cumulative upkeep prompt")
	}
	opts := g.PayActionOptionsForEffect(me.ID, prompt.PayAction())
	if len(opts) != 1 || opts[0] != gorger {
		t.Fatalf("options = %v, want just the Soulgorger", opts)
	}
}
