package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The +1 is the card: it attaches and it grows the creature. Both
// halves happen at resolution, not at announce — only the loyalty is
// paid up front (CR 601/602).
func TestAethersparkPlusOneAttachesAndPutsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, spark, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 5 {
		t.Errorf("loyalty after +1 on a 4-loyalty Aetherspark: got %d, want 5", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, spark); host.Kind == game.TargetCard {
		t.Error("the attach happened at announce; it belongs at resolution")
	}

	effectstest.PassPriorityAroundTable(t, g)

	host := effectstest.AttachmentHostOf(t, g, spark)
	if host.Kind != game.TargetCard || host.ID != bear {
		t.Errorf("attached to %+v, want the bear %s", host, bear)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the equipped creature: got %d, want 1", got)
	}
}

// "Up to one target": declining is a legal activation. The Aetherspark
// still ticks up and attaches to nothing, which is the whole reason
// the clause carries Min 0.
func TestAethersparkPlusOneWithNoTargetStillTicksUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, spark, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate +1 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoyaltyCount(g, spark); got != 5 {
		t.Errorf("loyalty after a targetless +1: got %d, want 5", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, spark); host.Kind == game.TargetCard {
		t.Errorf("attached to %+v with no target chosen", host)
	}
}

// The +1 only ever attaches to a creature its controller controls.
func TestAethersparkPlusOneCannotAttachToAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Theirs")

	err := g.ActivateCatalogAbility(me.ID, spark, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	if err == nil {
		t.Fatal("the +1 attached to a creature its controller does not control")
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 4 {
		t.Errorf("a refused activation changed loyalty: got %d, want 4", got)
	}
}

func TestAethersparkMinusFiveDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spark := effectstest.PushAetherspark(g, me.ID, 6)
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, spark, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -5: %v", err)
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 1 {
		t.Errorf("loyalty after -5 on a 6-loyalty Aetherspark: got %d, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand size %d -> %d, want +2 from the draw", handBefore, got)
	}
}

// CR 606.6 through the printed costs: a −5 needs five counters.
func TestAethersparkCannotMinusFiveBelowCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spark := effectstest.PushAetherspark(g, me.ID, 4)

	if err := g.ActivateCatalogAbility(me.ID, spark, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Errorf("-5 at 4 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
}

// "Add ten mana of any one color" is ONE colour pick worth ten mana
// (CR 605.1a keeps a loyalty ability off the mana-ability path, so it
// resolves off the stack like any other ability).
func TestAethersparkMinusTenAddsTenOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spark := effectstest.PushAetherspark(g, me.ID, 11)

	if err := g.ActivateCatalogAbility(me.ID, spark, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["R"] != 10 {
		t.Fatalf("pick = %+v, want ten of any one color", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := me.ManaPool.Total(); got != 10 {
		t.Errorf("pool holds %d mana, want 10", got)
	}
}

// The granted trigger: the equipped creature deals combat damage on
// its controller's turn, and that many loyalty counters land on The
// Aetherspark.
func TestAethersparkGainsLoyaltyEqualToEquippedCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, spark, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, spark); got != 5 {
		t.Fatalf("loyalty after the +1: got %d, want 5", got)
	}

	// A real combat damage step, not a lone synthetic event: the trigger
	// settles on the creature's total for the step (#5760), at the
	// boundary before a player gets priority. The 2/2 is a 3/3 with the
	// +1's counter.
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := effectstest.LoyaltyCount(g, spark); got != 5 {
		t.Error("the trigger skipped the stack — loyalty moved before it resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoyaltyCount(g, spark); got != 8 {
		t.Errorf("loyalty after 3 combat damage: got %d, want 8", got)
	}
}

// Unattached, The Aetherspark has no such ability at all: the clause
// is "as long as The Aetherspark is attached to a creature".
func TestAethersparkUnattachedGainsNoLoyaltyFromCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	life := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if life-opp.Life != 2 {
		t.Fatalf("the defender lost %d, want the bear's 2", life-opp.Life)
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 4 {
		t.Errorf("an unattached Aetherspark gained loyalty: got %d, want 4", got)
	}
}

// "During your turn" is a real restriction: the equipped creature
// blocking on somebody else's turn deals combat damage and The
// Aetherspark gains nothing.
func TestAethersparkGainsNoLoyaltyOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	blocker := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Blocker")
	spark := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "The Aetherspark",
		TypeLine:   "Legendary Artifact Planeswalker — Equipment",
		OracleID:   effectstest.TheAethersparkOracle,
		Owner:      opp.ID,
		Controller: opp.ID,
		Counters:   map[string]int{game.CounterLoyalty: 4},
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: blocker},
	})
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Test", 0, 3)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if c, ok := effectstest.BattlefieldCard(g, attacker); !ok || c.DamageMarked != 2 {
		t.Fatal("the equipped blocker did not deal its 2 combat damage")
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 4 {
		t.Errorf("loyalty grew on another player's turn: got %d, want 4", got)
	}
}
