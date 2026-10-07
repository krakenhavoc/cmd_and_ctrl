package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// auto_answer_test.go — ADR 0127 (#1961): coverage, keys, standing
// answers, and the state they live in. The card-level cases (Rhystic
// Study, Consecrated Sphinx, Esper Sentinel) are in
// cards/effects/auto_answer_test.go, where the catalog is.

const testRowKey = "test-study|triggered|Test Study — draw unless caster pays {1}|"

// resolvingTestRow runs fn as though a catalog triggered ability named
// testRowKey were resolving, with g.mu held — the context a card's
// Effect queues its prompts in.
func resolvingTestRow(g *Game, fn func()) {
	item := &StackItem{ID: uuid.New(), Params: EffectParams{Ability: &AbilityRef{
		Key: "test-study", Slot: AbilitySlotTriggered, Ref: "own:0",
		Name: "Test Study — draw unless caster pays {1}",
	}}}
	g.WithWriteLock(func() {
		g.beginResolvingLocked(item)
		g.resolutionDepth++
		fn()
		g.resolutionDepth--
		g.resolving = nil
	})
}

func choiceByID(g *Game, id uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return c
		}
	}
	return nil
}

func lastChoice(g *Game) *PendingChoice {
	if len(g.PendingChoices) == 0 {
		return nil
	}
	return g.PendingChoices[len(g.PendingChoices)-1]
}

// queueTestTax queues Rhystic Study's shape: payer is asked "{1}", and
// a decline asks drawer "draw a card?". It returns the tax prompt.
func queueTestTax(t *testing.T, g *Game, payer, drawer uuid.UUID, drew *bool) *PendingChoice {
	t.Helper()
	resolvingTestRow(g, func() {
		err := g.QueuePayUnlessForEffect(payer, uuid.Nil, "{1}", "Test Study — pay {1}?", func(g *Game) error {
			g.QueueConfirmForEffect(ConfirmPrompt{
				Chooser:  drawer,
				Question: "Test Study — draw a card?",
				OnAccept: func(*Game) error { *drew = true; return nil },
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	c := lastChoice(g)
	if c == nil || c.Kind != PendingChoicePayUnless {
		t.Fatalf("no pay_unless queued: %+v", c)
	}
	return c
}

func TestAutoAnswerKeysFollowTheRowAndTheBranch(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if want := testRowKey + "pay_unless#1"; tax.AutoAnswerKey != want {
		t.Fatalf("tax key = %q, want %q", tax.AutoAnswerKey, want)
	}
	if tax.AutoAnswerPrompt != "Test Study — pay {1}?" {
		t.Errorf("tax prompt = %q", tax.AutoAnswerPrompt)
	}
	if err := g.ResolvePayUnless(tax.ID, payer.ID, false); err != nil {
		t.Fatal(err)
	}
	draw := lastChoice(g)
	if draw == nil || draw.Kind != PendingChoiceConfirm {
		t.Fatalf("no draw prompt after the decline: %+v", draw)
	}
	if want := tax.AutoAnswerKey + ">confirm#1"; draw.AutoAnswerKey != want {
		t.Errorf("draw key = %q, want %q (derived from the tax's)", draw.AutoAnswerKey, want)
	}

	// A second resolution of the same row starts its ordinals again, so
	// the same question gets the same key every time it is asked.
	again := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if again.AutoAnswerKey != tax.AutoAnswerKey {
		t.Errorf("second ask's key = %q, want %q", again.AutoAnswerKey, tax.AutoAnswerKey)
	}
	// Two prompts of one kind in one resolution are #1 and #2.
	var second *PendingChoice
	resolvingTestRow(g, func() {
		_ = g.QueuePayUnlessForEffect(payer.ID, uuid.Nil, "{1}", "a", nil)
		_ = g.QueuePayUnlessForEffect(payer.ID, uuid.Nil, "{2}", "b", nil)
		second = lastChoice(g)
	})
	if !strings.HasSuffix(second.AutoAnswerKey, "|pay_unless#2") {
		t.Errorf("second prompt of one resolution keyed %q, want ordinal #2", second.AutoAnswerKey)
	}
}

func TestAutoAnswerNeverCoversABoardDependentPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	resolvingTestRow(g, func() {
		cases := map[string]func(){
			"guards a spell": func() {
				_ = g.queuePayUnlessLocked(me, uuid.Nil, "{1}", "ward", nil, TurnStep{}, uuid.New(), nil)
			},
			"owed in this step": func() {
				_ = g.QueueUpkeepPayUnlessForEffect(UpkeepPayUnlessPrompt{Chooser: me, Cost: "{1}", Question: "upkeep"})
			},
			"a non-mana payment": func() {
				_ = g.queuePayUnlessLocked(me, uuid.Nil, "discard a card", "echo", nil, TurnStep{}, uuid.Nil, nil,
					payUnlessExtras{action: &PayAction{Kind: PayActionDiscard, Count: 1}})
			},
			"a waterbend payment": func() {
				_ = g.queuePayUnlessLocked(me, uuid.Nil, "{4}", "waterbend", nil, TurnStep{}, uuid.Nil,
					&TapPermanentsCost{Key: "waterbend", Spec: &TargetSpec{}})
			},
			"an X cost": func() {
				_ = g.QueuePayUnlessForEffect(me, uuid.Nil, "{X}", "x", nil)
			},
			"a Phyrexian cost": func() {
				_ = g.QueuePayUnlessForEffect(me, uuid.Nil, "{W/P}", "phyrexian", nil)
			},
			"a confirm with named branches": func() {
				g.QueueConfirmForEffect(ConfirmPrompt{Chooser: me, AcceptLabel: "Exile it", DeclineLabel: "Keep it"})
			},
			"a confirm that costs life": func() {
				g.QueueConfirmForEffect(ConfirmPrompt{Chooser: me, LifeCost: 4})
			},
		}
		for name, queue := range cases {
			before := len(g.PendingChoices)
			queue()
			if len(g.PendingChoices) != before+1 {
				t.Errorf("%s: nothing queued", name)
				continue
			}
			if c := lastChoice(g); c.AutoAnswerKey != "" {
				t.Errorf("%s: keyed %q, want no key", name, c.AutoAnswerKey)
			}
		}
	})
	// Outside any resolution or branch there is no row to name.
	g.WithWriteLock(func() {
		_ = g.QueuePayUnlessForEffect(me, uuid.Nil, "{1}", "loose", nil)
	})
	if c := lastChoice(g); c.AutoAnswerKey != "" {
		t.Errorf("a prompt outside a resolution keyed %q", c.AutoAnswerKey)
	}
	// A face-down source names nothing.
	hidden := Card{InstanceID: uuid.New(), Name: "Secret", FaceDown: true, Owner: me, Controller: me}
	g.Battlefield.PushTop(hidden)
	resolvingTestRow(g, func() {
		_ = g.QueuePayUnlessForEffect(me, hidden.InstanceID, "{1}", "hidden", nil)
	})
	if c := lastChoice(g); c.AutoAnswerKey != "" {
		t.Errorf("a face-down source's prompt keyed %q", c.AutoAnswerKey)
	}
}

func TestAutoAnswerTriggerPromptCoverage(t *testing.T) {
	row := catalogRowID{key: "test-sphinx", index: 0}
	plain := TriggeredAbility{Key: "Test Sphinx — draw two cards", row: row, OptionalPrompt: &TriggerOptionalPrompt{}}
	src := Card{Name: "Test Sphinx"}
	if got, want := triggerPromptKey(src, plain), "test-sphinx|triggered|Test Sphinx — draw two cards"; got != want {
		t.Errorf("plain optional trigger keyed %q, want %q", got, want)
	}
	for name, mutate := range map[string]func(*TriggeredAbility, *Card){
		"targeted":   func(a *TriggeredAbility, _ *Card) { a.Targets = &TargetSpec{} },
		"modal":      func(a *TriggeredAbility, _ *Card) { a.Modes = &ModeSpec{} },
		"a trade":    func(a *TriggeredAbility, _ *Card) { a.OptionalPrompt = &TriggerOptionalPrompt{Trade: true} },
		"miracle":    func(a *TriggeredAbility, _ *Card) { a.Keyword = AltCostKeyMiracle },
		"not a row":  func(a *TriggeredAbility, _ *Card) { a.row = catalogRowID{} },
		"face down":  func(_ *TriggeredAbility, c *Card) { c.FaceDown = true },
		"no row key": func(a *TriggeredAbility, _ *Card) { a.Key = "" },
	} {
		a, c := plain, src
		mutate(&a, &c)
		if got := triggerPromptKey(c, a); got != "" {
			t.Errorf("%s trigger keyed %q, want none", name, got)
		}
	}
}

func TestAutoAnswerNeverDeclinesAndRunsTheBranch(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	id, chooser, ok := g.NextAutoAnswer()
	if !ok || id != tax.ID || chooser != payer.ID {
		t.Fatalf("NextAutoAnswer = %v %v %v, want the tax for the payer", id, chooser, ok)
	}
	if err := g.AutoAnswer(id); err != nil {
		t.Fatal(err)
	}
	if choiceByID(g, tax.ID) != nil {
		t.Fatal("the tax is still open")
	}
	draw := lastChoice(g)
	if draw == nil || draw.Kind != PendingChoiceConfirm || draw.Chooser != drawer.ID {
		t.Fatalf("Never did not run the decline branch: %+v", draw)
	}
	ev := g.Events[len(g.Events)-1]
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventAutoAnswer {
			ev = g.Events[i]
			break
		}
	}
	if ev.Kind != EventAutoAnswer || ev.Actor != payer.ID || ev.Call != AutoAnswerCallDontPay {
		t.Errorf("event = %+v, want an auto_answer by the payer, dont_pay", ev)
	}
	// The drawer has no rule: their prompt is theirs to answer.
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Error("a prompt with no rule was offered for an automatic answer")
	}
}

func TestAutoAnswerAlwaysPaysFromRealManaOrAsks(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	payer.ManaPool = nil
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Fatal("Always pay answered a tax nothing can pay")
	}
	if tax.AskedByHand != AskedByHandNoMana {
		t.Fatalf("AskedByHand = %q, want %q", tax.AskedByHand, AskedByHandNoMana)
	}
	// Mana arriving later does not answer it under the player.
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Fatal("a prompt asked by hand was answered once mana arrived")
	}

	// A fresh ask with the mana there pays it.
	again := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	id, _, ok := g.NextAutoAnswer()
	if !ok || id != again.ID {
		t.Fatalf("Always pay with {1} floating was not answered: %v %v", id, ok)
	}
	if err := g.AutoAnswer(id); err != nil {
		t.Fatal(err)
	}
	if len(payer.ManaPool) != 0 {
		t.Errorf("pool after paying = %d tokens, want 0", len(payer.ManaPool))
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceConfirm {
			t.Error("a paid tax ran the decline branch")
		}
	}
}

func TestAutoAnswerAlwaysWithAnEmptyLibraryIsAskedByHand(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	payer.Library.Cards = nil
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Fatal("Always was answered with an empty library")
	}
	if tax.AskedByHand != AskedByHandEmptyLibrary {
		t.Errorf("AskedByHand = %q, want %q", tax.AskedByHand, AskedByHandEmptyLibrary)
	}
}

func TestAutoAnswerStopsUnderALoopNotice(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	g.LoopNotice = &LoopNotice{Label: "a loop", Count: 25}
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Fatal("a prompt was answered automatically under a loop notice")
	}
	if tax.AskedByHand != AskedByHandLoop {
		t.Errorf("AskedByHand = %q, want %q", tax.AskedByHand, AskedByHandLoop)
	}
}

func TestAutoAnswerIsNotAPlayerDecision(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	g.TurnTally.LoopRun = map[string]int{"some loop": 7}
	if err := g.AutoAnswer(tax.ID); err != nil {
		t.Fatal(err)
	}
	if g.TurnTally.LoopRun["some loop"] != 7 {
		t.Errorf("an automatic answer restarted the CR 732 loop run: %v", g.TurnTally.LoopRun)
	}
	// The same answer by hand is a decision.
	draw := lastChoice(g)
	if err := g.ResolveConfirm(draw.ID, drawer.ID, false); err != nil {
		t.Fatal(err)
	}
	if len(g.TurnTally.LoopRun) != 0 {
		t.Errorf("an answer by hand did not restart the loop run: %v", g.TurnTally.LoopRun)
	}
}

func TestAutoAnswerNeverAnswersForABot(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	payer.IsBot = true
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Fatal("a bot seat's prompt was answered automatically")
	}
	if tax.AskedByHand != "" {
		t.Errorf("a bot seat's prompt was marked %q", tax.AskedByHand)
	}
	if err := g.AutoAnswer(tax.ID); err == nil {
		t.Error("AutoAnswer accepted a bot seat's prompt")
	}
	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{"k": AutoAnswerNever}); err == nil {
		t.Error("a bot seat accepted standing answers")
	}
}

func TestSetAutoAnswersRefusals(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	many := map[string]AutoAnswer{}
	for i := 0; i <= MaxAutoAnswerRules; i++ {
		many[uuid.NewString()] = AutoAnswerNever
	}
	for name, rules := range map[string]map[string]AutoAnswer{
		"101 rules":    many,
		"a long key":   {strings.Repeat("k", MaxAutoAnswerKeyLen+1): AutoAnswerNever},
		"an empty key": {"": AutoAnswerNever},
		"a bad answer": {"k": "sometimes"},
	} {
		if err := g.SetAutoAnswers(me, rules); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if err := g.SetAutoAnswers(g.ID, map[string]AutoAnswer{"k": AutoAnswerNever}); err == nil {
		t.Error("accepted rules for a non-seat")
	}
	if err := NewGame().SetAutoAnswers(me, nil); err == nil {
		t.Error("accepted rules before the game started")
	}
	if err := g.SetAutoAnswers(me, map[string]AutoAnswer{"k": AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	if err := g.SetAutoAnswers(me, nil); err != nil || len(g.Seats[0].AutoAnswers) != 0 {
		t.Errorf("an empty list did not clear the rules: %v %v", err, g.Seats[0].AutoAnswers)
	}
}

func TestAutoAnswerStateRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var drew bool
	tax := queueTestTax(t, g, payer.ID, drawer.ID, &drew)
	tax.AskedByHand = AskedByHandUndone
	rules := map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerAlways, "other": AutoAnswerNever}
	if err := g.SetAutoAnswers(payer.ID, rules); err != nil {
		t.Fatal(err)
	}

	c := g.Clone()
	if len(c.Seats[0].AutoAnswers) != 2 || len(c.Seats[1].AutoAnswers) != 0 {
		t.Errorf("Clone lost or smeared the rules: %v / %v", c.Seats[0].AutoAnswers, c.Seats[1].AutoAnswers)
	}
	c.Seats[0].AutoAnswers["other"] = AutoAnswerAlways
	if g.Seats[0].AutoAnswers["other"] != AutoAnswerNever {
		t.Error("Clone shares the rules map")
	}
	cc := choiceByID(c, tax.ID)
	if cc.AutoAnswerKey != tax.AutoAnswerKey || cc.AskedByHand != AskedByHandUndone {
		t.Errorf("Clone lost the prompt's key or mark: %+v", cc)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Seats[0].AutoAnswers) != 2 {
		t.Errorf("snapshot lost the rules: %v", snap.Seats[0].AutoAnswers)
	}
	if s := snap.PendingChoices[len(snap.PendingChoices)-1]; s.AutoAnswerKey != tax.AutoAnswerKey || s.AskedByHand != AskedByHandUndone {
		t.Errorf("snapshot lost the prompt's key or mark: %+v", s)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Seats[0].AutoAnswers; len(got) != 2 || got[tax.AutoAnswerKey] != AutoAnswerAlways {
		t.Errorf("restore = %v, want the two rules", got)
	}
}

// A rule this binary does not understand is never acted on: the prompt
// is asked, as with no rule at all.
func TestAutoAnswerIgnoresAnUnknownAnswer(t *testing.T) {
	g := newActiveGame(t)
	var drew bool
	tax := queueTestTax(t, g, g.Seats[0].ID, g.Seats[1].ID, &drew)
	g.Seats[0].AutoAnswers = map[string]AutoAnswer{tax.AutoAnswerKey: "sometimes"}
	if _, _, ok := g.NextAutoAnswer(); ok {
		t.Error("an unknown answer was acted on")
	}
}

// Undo is Clone + RestoreFrom. The rules are a setting, set with no
// undo entry, so rewinding an earlier action must not take them back.
func TestAutoAnswersSurviveUndo(t *testing.T) {
	g := newActiveGame(t)
	pre := g.Clone()
	if err := g.SetAutoAnswers(g.Seats[0].ID, map[string]AutoAnswer{"k": AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if g.Seats[0].AutoAnswers["k"] != AutoAnswerNever {
		t.Error("undo reverted the standing answers")
	}
}

func TestMarkAskedByHand(t *testing.T) {
	g := newActiveGame(t)
	var drew bool
	tax := queueTestTax(t, g, g.Seats[0].ID, g.Seats[1].ID, &drew)
	g.MarkAskedByHand(tax.ID, AskedByHandUndone)
	if tax.AskedByHand != AskedByHandUndone {
		t.Errorf("AskedByHand = %q", tax.AskedByHand)
	}
	g.MarkAskedByHand(uuid.New(), AskedByHandUndone) // not open: ignored
}
