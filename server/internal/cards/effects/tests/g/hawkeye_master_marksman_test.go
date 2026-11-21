package testsg

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hawkeye_master_marksman_test.go — #4579: "you may pay {1} up to three
// times. When you do, choose up to that many —", through the engine.

// Bullet indexes, in printed order.
const (
	hawkeyeNet = iota
	hawkeyeExplosive
	hawkeyeBoomerang
)

// hawkeyeTaps puts Hawkeye on the active seat's battlefield beside an
// opposing creature, fills the pool with `mana` colourless mana, taps
// Hawkeye by attacking with it, and resolves the tap trigger so that its
// "pay {1} up to three times" question is open (when there is mana to
// ask it with).
func hawkeyeTaps(t *testing.T, g *game.Game, mana int) (me, opp *game.Player, hawkeye, victim uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hawkeye = effectstest.PushDiesCreatureForTest(g, me.ID, "Hawkeye, Master Marksman", cardsh.HawkeyeMasterMarksmanOracle, "Legendary Creature — Human Archer Hero", 2, 2)
	victim = effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	me.Hand.PushTop(game.NewCard("Pitch", me.ID))
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hawkeye, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.FillPool(me, mana)
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp, hawkeye, victim
}

// payHawkeye answers the "how many times" prompt with n and pays.
func payHawkeye(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	effectstest.AnswerNumber(t, g, me.ID, n)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, true); c.PayCost != "{"+strconv.Itoa(n)+"}" {
		t.Fatalf("payment = %s, want {%d}", c.PayCost, n)
	}
}

func TestHawkeyeMasterMarksmanIsFull(t *testing.T) {
	spec, ok := effects.Lookup(cardsh.HawkeyeMasterMarksmanOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Triggered) != 1 {
		t.Fatalf("one tap trigger, Full: %+v", spec)
	}
	if len(spec.Caveats) != 0 {
		t.Errorf("Full takes no caveats: %v", spec.Caveats)
	}
}

// The count is capped at three whatever the pool holds, and "paying
// nothing" is the same as declining: no payment, no reflexive trigger.
func TestHawkeyePayingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _, _ := hawkeyeTaps(t, g, 5)
	pa := effectstest.AnswerNumber(t, g, me.ID, 0)
	if pa.Min != 0 || pa.Max != 3 {
		t.Fatalf("X prompt = %+v, want 0..3 although five mana is on offer", pa)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a payment was asked for 0 payments")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NoModePick(t, g, me.ID, "nothing was paid")
	if n := me.ManaPool.Total(); n != 5 {
		t.Errorf("pool %d, want 5 untouched", n)
	}
}

// With no mana at all the question is not asked (CR 118.3).
func TestHawkeyeWithNoManaAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _, _ := hawkeyeTaps(t, g, 0)
	if effectstest.PayAmountPrompt(g, me.ID) != nil || effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to pay with no mana to pay it with")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NoModePick(t, g, me.ID, "nothing could be paid")
}

// The mana can't cover a third payment, so three is refused (the ceiling
// is what the table can pay) and two is the most there is.
func TestHawkeyeCeilingIsWhatTheManaCovers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _, _ := hawkeyeTaps(t, g, 2)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil || c.PayAmount.Max != 2 {
		t.Fatalf("X prompt = %+v, want 0..2", c)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err == nil {
		t.Fatal("three payments accepted with two mana")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("two payments: %v", err)
	}
}

// Declining the payment after choosing a number makes no trigger.
func TestHawkeyeDecliningThePaymentDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _, _ := hawkeyeTaps(t, g, 3)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NoModePick(t, g, me.ID, "the payment was declined")
	if n := me.ManaPool.Total(); n != 3 {
		t.Errorf("pool %d, want 3 untouched", n)
	}
}

// Paid once: one bullet, two refused. The explosive deals 2 damage to the
// target player.
func TestHawkeyePaidOnceChoosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _, _ := hawkeyeTaps(t, g, 3)
	payHawkeye(t, g, me, 1)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool %d, want 2 left after one payment", n)
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt after paying")
	}
	if c.ModeMin != 0 || c.ModeMax != 1 {
		t.Errorf("prompt bounds %d..%d, want 0..1", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{hawkeyeNet, hawkeyeExplosive}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two bullets after one payment: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{hawkeyeExplosive, hawkeyeExplosive}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a bullet twice: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{hawkeyeExplosive}); err != nil {
		t.Fatalf("explosive: %v", err)
	}
	life := opp.Life
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("explosive dealt %d, want 2", got)
	}
}

// Paid three times: all three bullets, resolved in printed order whatever
// order they were chosen in, each with its own target. The pool keeps
// what was not spent.
func TestHawkeyePaidThriceRunsAllThreeBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _, victim := hawkeyeTaps(t, g, 5)
	payHawkeye(t, g, me, 3)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool %d, want 2 left after three payments", n)
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt after paying")
	}
	if c.ModeMax != 3 || len(c.ModeOptionIndex) != 3 {
		t.Fatalf("three payments: max %d, offered %v, want 3 of 3", c.ModeMax, c.ModeOptionIndex)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{hawkeyeBoomerang, hawkeyeExplosive, hawkeyeNet}); err != nil {
		t.Fatalf("all three: %v", err)
	}
	// One target per targeted bullet, in the order chosen: the explosive's
	// player, then the net's creature.
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	life, hand := opp.Life, effectstest.HandCount(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("explosive dealt %d, want 2", got)
	}
	if !effectstest.SbCantBlock(t, g, victim) {
		t.Error("net: the creature can still block")
	}
	if got := effectstest.HandCount(g, me); got != hand {
		t.Errorf("discard a card, then draw a card: hand %d → %d", hand, got)
	}
}

// Choosing fewer than were paid for is legal, and so is choosing none,
// which removes the trigger (CR 700.2b).
func TestHawkeyeMayChooseFewerThanPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _, _ := hawkeyeTaps(t, g, 3)
	payHawkeye(t, g, me, 3)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if err := g.ResolveModePick(c.ID, me.ID, []int{hawkeyeBoomerang}); err != nil {
		t.Fatalf("one of three: %v", err)
	}
	life, hand := opp.Life, effectstest.HandCount(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Error("the explosive was not chosen but damage was dealt")
	}
	if got := effectstest.HandCount(g, me); got != hand {
		t.Errorf("boomerang: hand %d → %d", hand, got)
	}

	g = effectstest.NewCatalogGame(t)
	me, _, _, _ = hawkeyeTaps(t, g, 3)
	payHawkeye(t, g, me, 2)
	c = effectstest.ModePickChoiceFor(g, me.ID)
	if err := g.ResolveModePick(c.ID, me.ID, []int{}); err != nil {
		t.Fatalf("choosing none: %v", err)
	}
	effectstest.NoModePick(t, g, me.ID, "asked again after choosing none")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked of a trigger with no bullet")
	}
}

// Any tap triggers it, not only the attack: an effect that taps Hawkeye
// on the controller's own turn asks the same question.
func TestHawkeyeTriggersOnAnyTapNotOnlyAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hawkeye := effectstest.PushDiesCreatureForTest(g, me.ID, "Hawkeye, Master Marksman", cardsh.HawkeyeMasterMarksmanOracle, "Legendary Creature — Human Archer Hero", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(hawkeye) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 1)
	if pa.Max != 2 {
		t.Errorf("X prompt = %+v, want 0..2", pa)
	}
}

// The payment count, the payment and the reflexive trigger on the stack
// are restore points, and the restored trigger still carries its
// bullets and its count.
func TestHawkeyePromptsAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _, _ := hawkeyeTaps(t, g, 5)
	c := effectstest.PayXPrompt(t, g, me.ID)
	if c.PickThen != cardsh.HawkeyeTrickArrowsPaid.Key() || c.MayPayX.Symbol != "{1}" || c.PayAmount.Max != 3 {
		t.Fatalf("the X prompt = %+v, want Hawkeye's, 0..3 of {1}", c)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c = effectstest.PayXPrompt(t, g, me.ID)
	if c.PayAmount.Min != 0 || c.PayAmount.Max != 3 {
		t.Fatalf("restored X prompt = %+v, want 0..3", c.PayAmount)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer: %v", err)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	pay := effectstest.PayUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{2}" || pay.PickNumber != 2 {
		t.Fatalf("restored payment prompt = %+v, want {2} carrying 2", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}

	// The mode and target prompts of the reflexive trigger are the
	// announcement of a trigger with no catalog row (trigger_frame_keyed.go,
	// owner decision D5), as for every reflexive trigger: not restore points,
	// so there is nothing to round-trip until it is on the stack.
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mp := effectstest.ModePickChoiceFor(g, me.ID)
	if mp == nil || mp.ModeMax != 2 || len(mp.ModeOptionIndex) != 3 {
		t.Fatalf("restored mode prompt = %+v, want up to 2 of 3", mp)
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{hawkeyeExplosive}); err != nil {
		t.Fatalf("explosive: %v", err)
	}
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}

	// The reflexive trigger is on the stack: restore it there.
	g = effectstest.RestoreRoundTrip(t, g, true)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the restored trigger dealt %d, want 2", got)
	}
}
