package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_batch_order_test.go — #1529, CR 603.3b / 603.3d. A targeted
// triggered ability used to open its target prompt BEFORE it joined the
// trigger queue, and the untargeted triggers that fired with it drained
// onto the stack meanwhile. It then drained alone, on top, so it always
// resolved first and its controller was never asked. The drain now holds
// the whole queue while a trigger of the batch is still announcing, and
// the targeted item joins the ordering prompt with the rest (ADR 0018's
// #1529 amendment).
//
// The repro is the issue's own: a prowess creature plus Caldera Pyremaw,
// and then the case where the order changes the result — a Pyremaw that
// has prowess itself deals one more damage pumped-then-burned than
// burned-then-pumped.

const pyremawBurnLabel = "Caldera Pyremaw — a +1/+1 counter, then damage equal to its power to target opponent"

// answerTriggerOrderResolvingLast is the mirror: `last` resolves after
// every other item.
func answerTriggerOrderResolvingLast(t *testing.T, g *game.Game, last uuid.UUID) {
	t.Helper()
	ch := effectstest.TriggerOrderPrompt(g)
	if ch == nil {
		t.Fatal("no trigger_order prompt to answer")
	}
	var order []uuid.UUID
	for _, id := range ch.TriggerOrderIDs {
		if id != last {
			order = append(order, id)
		}
	}
	order = append(order, last)
	if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, order); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
}

// topStackItem is the triggered item that resolves next.
func topStackItem(g *game.Game) *game.StackItem {
	var top *game.StackItem
	for _, it := range g.StackMeta {
		if it == nil || it.Kind != game.StackItemTriggered {
			continue
		}
		if top == nil || it.Seq > top.Seq {
			top = it
		}
	}
	return top
}

func promptHolds(ch *game.PendingChoice, id uuid.UUID) bool {
	for _, x := range ch.TriggerOrderIDs {
		if x == id {
			return true
		}
	}
	return false
}

// settleBatch passes priority until nothing is on or headed for the
// stack, failing on any prompt.
func settleBatch(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				t.Fatalf("a prompt is open while the batch settles: %+v", g.PendingChoices[0])
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// TestPyremawPlusProwessHoldsTheBatchForTheTarget — the issue's repro.
// While the Pyremaw's target prompt is open, the prowess trigger that
// fired with it waits in the queue instead of reaching the stack. Once
// the target is picked, both are offered in ONE ordering prompt, and the
// controller can put the Pyremaw's trigger under the prowess one.
func TestPyremawPlusProwessHoldsTheBatchForTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := effectstest.PushProwessCreature(g, me.ID, "Monk", 1, 1)
	pyremaw := effectstest.B43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", effectstest.B43CalderaPyremawOracle, 3, 3)
	before := opp.Life

	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)

	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the Pyremaw trigger opened no target prompt")
	}
	if n := effectstest.ProwessItemsFrom(g, monk); n != 1 {
		t.Fatalf("prowess items in flight = %d, want 1", n)
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered {
			t.Fatalf("%q reached the stack while its batch-mate was still choosing a target", it.Label)
		}
	}
	if effectstest.TriggerOrderPrompt(g) != nil {
		t.Fatal("the ordering prompt opened before the targeted trigger joined the batch")
	}

	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)

	ch := effectstest.TriggerOrderPrompt(g)
	if ch == nil {
		t.Fatal("no CR 603.3b ordering prompt once the target was chosen")
	}
	burn := effectstest.PendingItemLabelled(g, pyremawBurnLabel, pyremaw)
	if burn == nil {
		t.Fatal("the Pyremaw trigger is not in the queue with its batch")
	}
	if len(ch.TriggerOrderIDs) != 2 || !promptHolds(ch, burn.ID) {
		t.Fatalf("prompt orders %v, want the prowess item and the Pyremaw item %s", ch.TriggerOrderIDs, burn.ID)
	}
	if len(burn.Targets) != 1 || burn.Targets[0].ID != opp.ID {
		t.Fatalf("the Pyremaw item carries targets %+v, want the chosen opponent", burn.Targets)
	}

	// Burn resolves LAST: below the prowess trigger.
	answerTriggerOrderResolvingLast(t, g, burn.ID)
	if top := topStackItem(g); top == nil || top.SourceCardID != monk {
		t.Fatalf("top of the stack after ordering is %+v, want the prowess trigger", top)
	}
	settleBatch(t, g)

	if p := effectstest.EffectivePower(t, g, monk); p != 2 {
		t.Errorf("Monk power %d, want 2", p)
	}
	if opp.Life != before-4 {
		t.Errorf("opponent life %d → %d, want 4 damage (3/3 plus its counter)", before, opp.Life)
	}
}

// TestPyremawWithProwessOrderDecidesTheDamage — the reason the order is
// a real choice. A Pyremaw that has prowess itself deals 5 when its pump
// resolves first and 4 when its burn does; before #1529 the engine
// always picked 4.
func TestPyremawWithProwessOrderDecidesTheDamage(t *testing.T) {
	for _, tc := range []struct {
		name      string
		pumpFirst bool
		want      int
	}{
		{"pump then burn", true, 5},
		{"burn then pump", false, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pyremaw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Caldera Pyremaw", TypeLine: "Creature — Dragon",
				OracleID: effectstest.B43CalderaPyremawOracle, Power: 3, Toughness: 3,
				Keywords: []string{game.KeywordProwess},
				Owner:    me.ID, Controller: me.ID,
			})
			before := opp.Life

			effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
			effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
			burn := effectstest.PendingItemLabelled(g, pyremawBurnLabel, pyremaw)
			if burn == nil || effectstest.TriggerOrderPrompt(g) == nil {
				t.Fatal("the burn trigger and its prowess trigger were not offered for ordering")
			}
			if tc.pumpFirst {
				answerTriggerOrderResolvingLast(t, g, burn.ID)
			} else {
				effectstest.AnswerTriggerOrderResolvingFirst(t, g, burn.ID)
			}
			settleBatch(t, g)

			if got := before - opp.Life; got != tc.want {
				t.Errorf("damage %d, want %d", got, tc.want)
			}
			// effectivePower is the layered P/T without counters: 3 plus
			// the prowess pump. The counter is counted on its own.
			if p := effectstest.EffectivePower(t, g, pyremaw); p != 4 {
				t.Errorf("Pyremaw layered power %d after both resolved, want 4 (3 + prowess)", p)
			}
			if n := effectstest.CounterOn(g, pyremaw, game.CounterPlusOne); n != 1 {
				t.Errorf("Pyremaw +1/+1 counters %d, want 1", n)
			}
		})
	}
}

// TestHeldTargetPromptStillValidatesTheTarget — holding the batch does
// not loosen the CR 603.3d pick: an illegal answer ("target opponent",
// answered with the controller) is refused, the prompt stays open and
// the batch stays held.
func TestHeldTargetPromptStillValidatesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := effectstest.PushProwessCreature(g, me.ID, "Monk", 1, 1)
	effectstest.B43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", effectstest.B43CalderaPyremawOracle, 3, 3)

	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range p.PickTargetPlayers {
		if id == me.ID {
			t.Fatal("the controller was offered as a target opponent")
		}
	}
	err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("picking yourself for \"target opponent\": err %v, want ErrIllegalTarget", err)
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the refused answer closed the prompt")
	}
	if pendingProwess := effectstest.ProwessItemsFrom(g, monk); pendingProwess != 1 || len(g.PendingTriggers) != 1 {
		t.Fatalf("after a refused pick: %d pending, want the held prowess item alone", len(g.PendingTriggers))
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	if effectstest.TriggerOrderPrompt(g) == nil {
		t.Fatal("a legal pick did not bring the batch to its ordering prompt")
	}
}

// TestHeldBatchRoundTripsUndo — undo is Clone + RestoreFrom. Restoring
// to the open target prompt gives back the same held batch and a prompt
// that can be answered again (the answer is not consumed by the undone
// attempt); restoring to the open ordering prompt gives back the same
// batch to order the other way.
func TestHeldBatchRoundTripsUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyremaw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Caldera Pyremaw", TypeLine: "Creature — Dragon",
		OracleID: effectstest.B43CalderaPyremawOracle, Power: 3, Toughness: 3,
		Keywords: []string{game.KeywordProwess},
		Owner:    me.ID, Controller: me.ID,
	})
	before := opp.Life

	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	atPick := g.Clone()
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	atOrder := g.Clone()
	burn := effectstest.PendingItemLabelled(g, pyremawBurnLabel, pyremaw)
	if burn == nil {
		t.Fatal("no burn item queued")
	}
	effectstest.AnswerTriggerOrderResolvingFirst(t, g, burn.ID)
	settleBatch(t, g)
	if got := before - opp.Life; got != 4 {
		t.Fatalf("first run: damage %d, want 4", got)
	}

	// Undo back to the ordering prompt, and order it the other way.
	// RestoreFrom swaps the seats for the snapshot's, so the player is
	// read again after every restore.
	g.RestoreFrom(atOrder)
	opp = g.Seats[1]
	if opp.Life != before {
		t.Fatalf("restore to the ordering prompt: opponent life %d, want %d", opp.Life, before)
	}
	burn = effectstest.PendingItemLabelled(g, pyremawBurnLabel, pyremaw)
	if burn == nil || effectstest.TriggerOrderPrompt(g) == nil || !promptHolds(effectstest.TriggerOrderPrompt(g), burn.ID) {
		t.Fatal("the restored game lost the ordering prompt or the burn item")
	}
	answerTriggerOrderResolvingLast(t, g, burn.ID)
	settleBatch(t, g)
	if got := before - opp.Life; got != 5 {
		t.Errorf("after undo to the ordering prompt: damage %d, want 5", got)
	}

	// Undo back to the open target prompt: the pick can be made again.
	g.RestoreFrom(atPick)
	opp = g.Seats[1]
	if top := topStackItem(g); top != nil || len(g.PendingTriggers) != 1 || effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatalf("restore to the target prompt: trigger on the stack %+v, %d pending; want none, the held prowess item and the open prompt",
			top, len(g.PendingTriggers))
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	burn = effectstest.PendingItemLabelled(g, pyremawBurnLabel, pyremaw)
	if burn == nil || effectstest.TriggerOrderPrompt(g) == nil {
		t.Fatal("re-answering the restored target prompt did not queue the burn with its batch")
	}
	if len(burn.Targets) != 1 || burn.Targets[0].ID != opp.ID {
		t.Fatalf("re-answered burn targets %+v, want exactly the opponent", burn.Targets)
	}
	answerTriggerOrderResolvingLast(t, g, burn.ID)
	settleBatch(t, g)
	if got := before - opp.Life; got != 5 {
		t.Errorf("after undo to the target prompt: damage %d, want 5", got)
	}
	if n := effectstest.CounterOn(g, pyremaw, game.CounterPlusOne); n != 1 {
		t.Errorf("Pyremaw counters %d, want exactly 1", n)
	}
}
