package game

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// trigger_order_memory_test.go — "use this order again" (ADR 0018's
// eighth amendment, #3016).

// queueRowForTest queues a stamped catalog trigger of row `row` (its
// AbilityRef key) from source, labelled label.
func queueRowForTest(g *Game, owner *Player, source uuid.UUID, row, label string, log *[]string) *StackItem {
	item := queueTriggerForTest(g, owner, source, label, log)
	item.Body = "catalog/triggered"
	item.Params.Ability = &AbilityRef{Key: row, Slot: AbilitySlotTriggered, Ref: "own:0", Name: row + "-trigger"}
	return item
}

// alwaysAskForTest puts owner on "always ask", so every batch of two or
// more is asked whatever the footprint check would say (E1).
func alwaysAskForTest(t *testing.T, g *Game, owner *Player) {
	t.Helper()
	if err := g.SetTriggerOrderPreference(owner.ID, TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
}

func rememberTrue() *bool  { v := true; return &v }
func rememberFalse() *bool { v := false; return &v }

// teachForTest queues a Plunderer and a Noble, drains, and answers
// Plunderer first with remember. It returns the prompt's key.
func teachForTest(t *testing.T, g *Game, me *Player, log *[]string) string {
	t.Helper()
	p := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", log)
	n := queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil || prompt.AutoAnswerKey == "" {
		t.Fatalf("want a keyed prompt, got %+v", prompt)
	}
	if err := g.ResolveTriggerOrderRemember(prompt.ID, me.ID, []uuid.UUID{p.ID, n.ID}, rememberTrue()); err != nil {
		t.Fatal(err)
	}
	if _, ok := me.TriggerOrderMemory[prompt.AutoAnswerKey]; !ok {
		t.Fatalf("remember: true stored nothing: %v", me.TriggerOrderMemory)
	}
	return prompt.AutoAnswerKey
}

// A1: the key is the set of distinct rows. Counts and source instances
// don't matter.
func TestTriggerOrderBatchKeyIgnoresCountsAndSources(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	mk := func(row string) *StackItem {
		it := queueRowForTest(g, me, uuid.New(), row, row, &log)
		return it
	}
	one := []*StackItem{mk("noble"), mk("solemn")}
	three := []*StackItem{mk("noble"), mk("noble"), mk("solemn"), mk("noble")}
	reversed := []*StackItem{mk("solemn"), mk("noble")}
	other := []*StackItem{mk("noble"), mk("plunderer")}
	g.WithWriteLock(func() {
		k1, _ := g.triggerOrderBatchKeyLocked(one)
		k3, _ := g.triggerOrderBatchKeyLocked(three)
		kr, _ := g.triggerOrderBatchKeyLocked(reversed)
		ko, _ := g.triggerOrderBatchKeyLocked(other)
		if k1 == "" || !strings.HasPrefix(k1, TriggerOrderAutoAnswerPrefix) {
			t.Fatalf("key = %q", k1)
		}
		if k1 != k3 || k1 != kr {
			t.Errorf("counts or order changed the key: %q / %q / %q", k1, k3, kr)
		}
		if k1 == ko {
			t.Error("a different set of rows has the same key")
		}
	})
}

// A closure row, a delayed trigger and a face-down or hidden source
// leave the whole batch unkeyed (A, G).
func TestTriggerOrderBatchKeyLeavesUnkeyedBatches(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	row := queueRowForTest(g, me, uuid.New(), "noble", "Noble", &log)
	closure := queueTriggerForTest(g, me, uuid.New(), "closure", &log)
	delayed := queueTriggerForTest(g, me, uuid.New(), "delayed", &log)
	delayed.Body = "draw/one-card"
	faceDownID := pushBattlefieldForTest(g, me.ID, "Morph", "Creature", "oracle-morph")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == faceDownID {
				g.Battlefield.Cards[i].FaceDown = true
			}
		}
	})
	faceDown := queueRowForTest(g, me, faceDownID, "morph", "Morph", &log)
	inHand := uuid.New()
	me.Hand.PushTop(Card{InstanceID: inHand, Name: "Hidden", OracleID: "oracle-hidden", Owner: me.ID, Controller: me.ID})
	hand := queueRowForTest(g, me, inHand, "hidden", "Hidden", &log)
	prowessSrc := pushBattlefieldForTest(g, me.ID, "Monk", "Creature — Monk", "oracle-monk")
	var monk *Card
	g.WithWriteLock(func() { monk = g.findCardByIDLocked(prowessSrc) })
	prowess := prowessTrigger.Build(Event{}, monk, Characteristic{}, g)
	prowess.ID = uuid.New()

	g.WithWriteLock(func() {
		if k, _ := g.triggerOrderBatchKeyLocked([]*StackItem{row, prowess}); k == "" {
			t.Error("a catalog row and a keyword trigger were left unkeyed")
		}
		for name, it := range map[string]*StackItem{"closure": closure, "delayed": delayed} {
			if k, _ := g.triggerOrderBatchKeyLocked([]*StackItem{row, it}); k != "" {
				t.Errorf("a batch with a %s trigger was keyed %q", name, k)
			}
		}
		for name, it := range map[string]*StackItem{"face-down": faceDown, "in-hand": hand} {
			k, hidden := g.triggerOrderBatchKeyLocked([]*StackItem{row, it})
			if k != "" || !hidden {
				t.Errorf("a batch with a %s source was keyed %q (hidden=%v)", name, k, hidden)
			}
		}
	})
}

// The memory applies on a repeat batch, by stable sort, as an ADR 0127
// automatic answer, with copies of one row kept in the order they
// triggered (B).
func TestRememberedTriggerOrderAnswersARepeatBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me) // E1: a remembered order still answers
	var log []string
	key := teachForTest(t, g, me, &log)

	// Three drains and a Plunderer: the same set of rows.
	n1 := queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	p := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	n2 := queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil || prompt.AutoAnswerKey != key {
		t.Fatalf("repeat batch prompt = %+v, want key %q", prompt, key)
	}
	choiceID, chooser, ok := g.NextAutoAnswer()
	if !ok || choiceID != prompt.ID || chooser != me.ID {
		t.Fatalf("NextAutoAnswer = %v %v %v, want the repeat prompt", choiceID, chooser, ok)
	}
	g.TurnTally.LoopRun = map[string]int{"sac loop": 7}
	seq, err := g.AutoAnswer(choiceID)
	if err != nil {
		t.Fatal(err)
	}
	if g.TurnTally.LoopRun["sac loop"] != 7 {
		t.Errorf("a remembered order restarted the CR 732 loop run: %v", g.TurnTally.LoopRun)
	}
	if findTriggerOrderPrompt(g, me.ID) != nil {
		t.Fatal("the prompt is still open")
	}
	// Resolution order Plunderer, Noble, Noble: the first to resolve is
	// the highest on the stack, and the drains keep their order.
	if !(p.Seq > n1.Seq && n1.Seq > n2.Seq) {
		t.Errorf("stack seqs Plunderer %d, Noble %d, Noble %d; want Plunderer, then the Nobles in trigger order", p.Seq, n1.Seq, n2.Seq)
	}
	var ev *Event
	for i := range g.Events {
		if g.Events[i].Seq == seq {
			ev = &g.Events[i]
		}
	}
	if ev == nil || ev.Kind != EventAutoAnswer || ev.Call != AutoAnswerCallOrder || ev.Actor != me.ID {
		t.Fatalf("auto_answer event = %+v", ev)
	}
	if ev.Label != "Plunderer — Treasure, then Noble — drain 1" || ev.AnswerKey != key {
		t.Errorf("event label %q key %q", ev.Label, ev.AnswerKey)
	}
}

// A batch with a different set of rows asks again.
func TestRememberedTriggerOrderAsksADifferentBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	teachForTest(t, g, me, &log)
	queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	queueRowForTest(g, me, uuid.New(), "solemn", "Solemn — draw", &log)
	drainForTest(g)
	if findTriggerOrderPrompt(g, me.ID) == nil {
		t.Fatal("no prompt for the new batch")
	}
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Error("a batch with a different set of rows was answered automatically")
	}
}

// B: an answer that interleaves copies of one row is not a ranking and
// is refused for remember, with nothing changed.
func TestRememberRefusesAnInterleavedOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	p1 := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer", &log)
	n := queueRowForTest(g, me, uuid.New(), "noble", "Noble", &log)
	p2 := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer", &log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)
	err := g.ResolveTriggerOrderRemember(prompt.ID, me.ID, []uuid.UUID{p1.ID, n.ID, p2.ID}, rememberTrue())
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("interleaved remember: err = %v, want ErrInvalidParam", err)
	}
	if findTriggerOrderPrompt(g, me.ID) == nil || len(me.TriggerOrderMemory) != 0 {
		t.Fatal("a refused answer changed the game")
	}
	// The same order without remember is an ordinary answer.
	if err := g.ResolveTriggerOrderRemember(prompt.ID, me.ID, []uuid.UUID{p1.ID, n.ID, p2.ID}, nil); err != nil {
		t.Fatal(err)
	}
}

// remember: false forgets the batch's order; ForgetTriggerOrder ("Ask me
// next time") forgets one by key or all with none.
func TestForgettingARememberedTriggerOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	teachForTest(t, g, me, &log)

	// Asked by hand (say, undone): answering with the toggle off forgets.
	p := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	n := queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)
	prompt.AskedByHand = AskedByHandUndone
	if err := g.ResolveTriggerOrderRemember(prompt.ID, me.ID, []uuid.UUID{n.ID, p.ID}, rememberFalse()); err != nil {
		t.Fatal(err)
	}
	if len(me.TriggerOrderMemory) != 0 {
		t.Fatalf("remember: false kept %v", me.TriggerOrderMemory)
	}

	key := teachForTest(t, g, me, &log)
	if err := g.ForgetTriggerOrder(me.ID, key); err != nil {
		t.Fatal(err)
	}
	if len(me.TriggerOrderMemory) != 0 {
		t.Fatalf("Ask me next time kept %v", me.TriggerOrderMemory)
	}
	queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	drainForTest(g)
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Error("a forgotten order still answered")
	}

	g.WithWriteLock(func() {
		g.storeTriggerOrderLocked(me, "trigger_order|a", []string{"a"})
		g.storeTriggerOrderLocked(me, "trigger_order|b", []string{"b"})
	})
	if err := g.ForgetTriggerOrder(me.ID, ""); err != nil || len(me.TriggerOrderMemory) != 0 {
		t.Errorf("forget all: err %v, left %v", err, me.TriggerOrderMemory)
	}
	if err := g.ForgetTriggerOrder(me.ID, strings.Repeat("k", MaxForgetTriggerOrderKeyLen+1)); err == nil {
		t.Error("an oversized key was accepted")
	}
}

// H: the memory is written by the commit that taught it, so undoing
// that commit takes it back, and undoing an automatic answer keeps it.
func TestUndoTakesTheMemoryBack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	pre := g.Clone()
	teachForTest(t, g, me, &log)
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if len(g.Seats[g.Turn.ActiveSeat].TriggerOrderMemory) != 0 {
		t.Fatal("undoing the teaching commit kept the memory (RestoreFrom must not carry it)")
	}

	me = g.Seats[g.Turn.ActiveSeat]
	key := teachForTest(t, g, me, &log)
	queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	drainForTest(g)
	choiceID, _, ok := g.NextAutoAnswer()
	if !ok {
		t.Fatal("the repeat batch was not offered for an automatic answer")
	}
	pre = g.Clone()
	if _, err := g.AutoAnswer(choiceID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	g.MarkAskedByHand(choiceID, AskedByHandUndone)
	me = g.Seats[g.Turn.ActiveSeat]
	if _, ok := me.TriggerOrderMemory[key]; !ok {
		t.Fatal("undoing the automatic answer lost the memory")
	}
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Error("the undone prompt was answered again")
	}
	var restored *PendingChoice
	g.WithWriteLock(func() { _, restored = g.findChoiceLocked(choiceID) })
	if restored == nil || restored.AskedByHand != AskedByHandUndone {
		t.Fatalf("undone prompt = %+v, want it reopened and asked by hand", restored)
	}
	g.WithWriteLock(func() {
		if !g.TriggerOrderRememberedLocked(restored) {
			t.Error("the reopened prompt does not say its order is remembered")
		}
	})
}

// Owner decision 7: at most 32 per seat; the 33rd replaces the one
// applied least recently.
func TestTriggerOrderMemoryCap(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < MaxTriggerOrderMemory; i++ {
			g.storeTriggerOrderLocked(me, "trigger_order|"+strconv.Itoa(i), []string{"r"})
		}
		// Re-use #0: now #1 is the least recently applied.
		g.storeTriggerOrderLocked(me, "trigger_order|0", []string{"r"})
		g.storeTriggerOrderLocked(me, "trigger_order|new", []string{"r"})
	})
	if len(me.TriggerOrderMemory) != MaxTriggerOrderMemory {
		t.Fatalf("memory holds %d, want %d", len(me.TriggerOrderMemory), MaxTriggerOrderMemory)
	}
	if _, ok := me.TriggerOrderMemory["trigger_order|1"]; ok {
		t.Error("the least recently applied order was kept")
	}
	for _, k := range []string{"trigger_order|0", "trigger_order|new", "trigger_order|2"} {
		if _, ok := me.TriggerOrderMemory[k]; !ok {
			t.Errorf("%s was dropped", k)
		}
	}
}

// The ADR 0127 hand checks apply: a loop notice and a controlled
// chooser are asked by hand, and a batch whose rows changed since it
// was queued is too.
func TestRememberedTriggerOrderHandChecks(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	teachForTest(t, g, me, &log)
	queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	n := queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)

	g.LoopNotice = &LoopNotice{Label: "a loop", Count: 25}
	if _, _, ok := g.NextAutoAnswer(); ok || prompt.AskedByHand != AskedByHandLoop {
		t.Fatalf("under a loop notice: answered=%v mark=%q", ok, prompt.AskedByHand)
	}
	g.LoopNotice = nil
	prompt.AskedByHand = ""

	n.Params.Ability = &AbilityRef{Key: "solemn", Slot: AbilitySlotTriggered, Ref: "own:0", Name: "x"}
	if _, _, ok := g.NextAutoAnswer(); ok || prompt.AskedByHand != AskedByHandChanged {
		t.Fatalf("a changed batch: answered=%v mark=%q", ok, prompt.AskedByHand)
	}
}

// A bot seat never remembers, and is never answered.
func TestRememberedTriggerOrderIgnoresABot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	me.IsBot = true
	var log []string
	p := queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer", &log)
	n := queueRowForTest(g, me, uuid.New(), "noble", "Noble", &log)
	drainForTest(g)
	prompt := findTriggerOrderPrompt(g, me.ID)
	if err := g.ResolveTriggerOrderRemember(prompt.ID, me.ID, []uuid.UUID{p.ID, n.ID}, rememberTrue()); err != nil {
		t.Fatal(err)
	}
	if len(me.TriggerOrderMemory) != 0 {
		t.Errorf("a bot seat remembered %v", me.TriggerOrderMemory)
	}
}

// "Never" asks nothing and learns nothing; a remembered order stays,
// unused (E).
func TestNeverModeLeavesTheMemoryUnused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	key := teachForTest(t, g, me, &log)
	if err := g.SetTriggerOrderPreference(me.ID, TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	queueRowForTest(g, me, uuid.New(), "plunderer", "Plunderer — Treasure", &log)
	queueRowForTest(g, me, uuid.New(), "noble", "Noble — drain 1", &log)
	if !drainForTest(g) || findTriggerOrderPrompt(g, me.ID) != nil {
		t.Fatal("never mode asked")
	}
	if _, ok := me.TriggerOrderMemory[key]; !ok {
		t.Error("never mode dropped the memory")
	}
}

// C1: cloned, snapshotted additively, restored.
func TestTriggerOrderMemoryRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alwaysAskForTest(t, g, me)
	var log []string
	key := teachForTest(t, g, me, &log)

	c := g.Clone()
	cm := c.Seats[g.Turn.ActiveSeat].TriggerOrderMemory
	if len(cm[key].Rows) != 2 {
		t.Fatalf("Clone lost the memory: %v", cm)
	}
	cm[key].Rows[0] = "changed"
	if me.TriggerOrderMemory[key].Rows[0] == "changed" {
		t.Error("Clone shares the ranking")
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"triggerOrderMemory"`) {
		t.Error("the snapshot has no triggerOrderMemory")
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	got := r.Seats[g.Turn.ActiveSeat].TriggerOrderMemory[key]
	if len(got.Rows) != 2 || got.Rows[0] != me.TriggerOrderMemory[key].Rows[0] {
		t.Errorf("restore = %+v, want %+v", got, me.TriggerOrderMemory[key])
	}
}
