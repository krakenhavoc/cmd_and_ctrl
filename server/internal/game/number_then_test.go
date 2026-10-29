package game

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// number_then_test.go — #4294: a number prompt whose consequence is a
// registered key (RegisterNumberThen) carries no closure, so a table
// waiting on it is a restore point, and the branch is handed the number
// and the data the prompt carried.

// numberProbeSeen is what the probe branch was handed, one entry per run.
var numberProbeSeen []NumberAnswer

var numberProbe = RegisterNumberThen(testEffectKeyPrefix+"number-probe", func(_ *Game, r NumberAnswer) error {
	numberProbeSeen = append(numberProbeSeen, r)
	return nil
})

// numberProbeBranch is the probe with the data every test hands it.
func numberProbeBranch(item *StackItem) NumberBranch {
	return NumberBranch{
		Then:  numberProbe,
		Carry: []uuid.UUID{{1}, {2}},
		Words: []string{"charge", "+1/+1"},
		Nums:  []int{3, 0, 7},
		Asked: NewAskedItem(item),
	}
}

func numberProbeItem(controller, source, target uuid.UUID) *StackItem {
	return &StackItem{
		ID: uuid.New(), Kind: StackItemTriggered, Controller: controller, Owner: controller,
		SourceCardID: source, Label: "probe", XValue: 2,
		Targets: []TargetRef{{Kind: TargetCard, ID: target}},
	}
}

// A chosen number carries the branch's key and data and no frame: the
// table is a restore point, the restored prompt keeps all of it, and its
// answer runs the branch with the number, the data and the asking item.
func TestKeyedChooseNumberSurvivesARestore(t *testing.T) {
	for _, resource := range []string{PayResourceNone, PayResourceLife} {
		numberProbeSeen = nil
		g := newActiveGame(t)
		me := g.Seats[0]
		lifeBefore := me.Life
		source, target := uuid.New(), uuid.New()
		item := numberProbeItem(me.ID, source, target)
		queueChooseNumber(t, g, ChooseNumber{
			Chooser: me.ID, Source: source, Resource: resource, Min: 0, Max: 5, Goal: 2,
			Unit: PayAmountCounters, Question: "how many?", Branch: numberProbeBranch(item),
		})
		c := payAmountChoice(g)
		if c == nil || c.PickThen != numberProbe.Key() {
			t.Fatalf("%s: prompt = %+v, want the probe's key and no frame", resource, c)
		}

		g = restoreStrictForTest(t, g)
		me = g.Seats[0]
		c = payAmountChoice(g)
		if c == nil || c.PickThen != numberProbe.Key() || c.PayAmount == nil || c.PayAmount.Goal != 2 || c.PayAmount.Unit != PayAmountCounters {
			t.Fatalf("%s: restored prompt = %+v", resource, c)
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
			t.Fatalf("%s: answer: %v", resource, err)
		}
		if len(numberProbeSeen) != 1 {
			t.Fatalf("%s: the branch ran %d times, want once", resource, len(numberProbeSeen))
		}
		r := numberProbeSeen[0]
		if r.Number != 4 || r.Chooser != me.ID || r.Source != source {
			t.Errorf("%s: handed number %d chooser %v source %v", resource, r.Number, r.Chooser, r.Source)
		}
		if !reflect.DeepEqual(r.Carry, []uuid.UUID{{1}, {2}}) || !reflect.DeepEqual(r.Words, []string{"charge", "+1/+1"}) || !reflect.DeepEqual(r.Nums, []int{3, 0, 7}) {
			t.Errorf("%s: handed data %v %v %v", resource, r.Carry, r.Words, r.Nums)
		}
		if r.Item == nil || r.Item.ID != item.ID || r.Item.XValue != 2 || len(r.Item.Targets) != 1 || r.Item.Targets[0].ID != target || r.Item.Controller != me.ID {
			t.Errorf("%s: handed item %+v, want the asking item rebuilt from data", resource, r.Item)
		}
		if resource == PayResourceLife && me.Life != lifeBefore-4 {
			t.Errorf("life %d after paying 4, want %d", me.Life, lifeBefore-4)
		}
	}
}

// An energy amount is paid before the branch runs, restored or not.
func TestKeyedEnergyAmountSurvivesARestore(t *testing.T) {
	numberProbeSeen = nil
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 5)
	g.WithWriteLock(func() {
		if err := g.QueuePayEnergyAmountForEffect(PayEnergyAmount{
			Chooser: me.ID, Source: uuid.New(), Unit: PayAmountDamage, Goal: 3, Question: "pay?",
			Branch: numberProbeBranch(nil),
		}); err != nil {
			t.Fatal(err)
		}
	})
	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	c := payAmountChoice(g)
	if c == nil || c.PayAmount.Max != 5 || c.PayAmount.Goal != 3 {
		t.Fatalf("restored prompt = %+v, want 0..5 with goal 3", c)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 6); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("answer past the energy: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if PlayerEnergy(me) != 2 {
		t.Errorf("energy %d, want 2", PlayerEnergy(me))
	}
	if len(numberProbeSeen) != 1 || numberProbeSeen[0].Number != 3 || numberProbeSeen[0].Item != nil {
		t.Errorf("the branch saw %+v, want 3 and no asking item", numberProbeSeen)
	}
}

// Nothing to ask: no energy to pay runs the branch at once with 0, and a
// chooser who leaves while a prompt is open gives the fallback: 0 for a
// payment, the floor for a number that is not paid (CR 800.4f).
func TestKeyedNumberRunsOnceWhenNothingIsAskedAndOnDeparture(t *testing.T) {
	numberProbeSeen = nil
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.QueuePayEnergyAmountForEffect(PayEnergyAmount{
			Chooser: me.ID, Source: uuid.New(), Branch: numberProbeBranch(nil),
		}); err != nil {
			t.Fatal(err)
		}
	})
	if payAmountChoice(g) != nil || len(numberProbeSeen) != 1 || numberProbeSeen[0].Number != 0 {
		t.Fatalf("no energy: prompt %v, runs %+v; want no prompt and one run with 0", payAmountChoice(g), numberProbeSeen)
	}

	for _, tc := range []struct {
		resource string
		want     int
	}{{PayResourceNone, 2}, {PayResourceLife, 0}} {
		numberProbeSeen = nil
		g := newActiveGameWithSeats(t, 3)
		me := g.Seats[0]
		// The card that asked is another seat's, so it outlives the chooser
		// (CR 800.4a) and the rest of the card still runs.
		source := departureTestSource(g, g.Seats[1].ID, "Probe")
		queueChooseNumber(t, g, ChooseNumber{
			Chooser: me.ID, Source: source, Resource: tc.resource, Min: 2, Max: 5, Branch: numberProbeBranch(nil),
		})
		g = restoreStrictForTest(t, g)
		if err := g.Concede(g.Seats[0].ID); err != nil {
			t.Fatal(err)
		}
		if payAmountChoice(g) != nil {
			t.Errorf("%s: a prompt outlived its chooser", tc.resource)
		}
		if len(numberProbeSeen) != 1 || numberProbeSeen[0].Number != tc.want {
			t.Errorf("%s: departure ran %+v, want once with %d", tc.resource, numberProbeSeen, tc.want)
		}
	}
}

// A branch no binary registered is an error rather than a silent no-op.
// (The closure form a branch could be set beside is gone: #4361.)
func TestKeyedNumberRefusesAnUnknownKey(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		err := g.runNumberBranch(testEffectKeyPrefix+"never-registered", NumberAnswer{}, carriedCause{})
		if !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("an unknown branch: err = %v, want ErrUnknownEffectKey", err)
		}
	})
}

// A restore point naming a branch this binary lacks is refused whole,
// rather than restored into a prompt with nothing behind it.
func TestKeyedNumberPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Max: 3, Branch: numberProbeBranch(nil),
	})
	snap := g.CaptureSnapshot()
	if len(snap.PendingChoices) != 1 || snap.PendingChoices[0].PickThen != numberProbe.Key() {
		t.Fatalf("snapshot prompts = %+v", snap.PendingChoices)
	}
	snap.PendingChoices[0].PickThen = testEffectKeyPrefix + "never-registered"
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown number branch: err = %v, want ErrUnknownEffectKey", err)
	}
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("strict restore with an unknown number branch: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- the engine's own number branches ---------------------------------

// A cast announcement is data: every field an assist re-run reads
// survives JSON, including the three that are unexported on the cast.
func TestAssistAnnounceSurvivesJSON(t *testing.T) {
	zero := 1
	p := CastSpellParams{
		FromZone: "hand", Targets: []TargetRef{{Kind: TargetPlayer, ID: uuid.New(), Slot: 1}},
		Modes: []int{1}, XValue: 3, Distribution: map[uuid.UUID]int{{7}: 2},
		DiscardIDs: []uuid.UUID{{3}}, OptionalCosts: []int{0, 0}, CostBranch: &zero,
		Strict: true, AutoTap: true, AssistPlayer: uuid.New(),
		PaperPayment:     &PaperPayment{Free: true, Mana: []PaperMana{{Color: "R", Count: 2}}},
		commanderAnswers: map[uuid.UUID]bool{{9}: true},
		opponentPicks:    []opponentPick{{mode: 1, slot: 2, chooser: uuid.UUID{5}, answered: true, refs: []TargetRef{{Kind: TargetCard, ID: uuid.UUID{6}}}}},
		splicedRefs:      []TargetRef{{Kind: TargetCard, ID: uuid.UUID{8}, Splice: 1}},
	}
	raw, err := json.Marshal(newAssistAnnounce(p))
	if err != nil {
		t.Fatal(err)
	}
	var back AssistAnnounce
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	got := back.params()
	if !reflect.DeepEqual(got, p) {
		t.Errorf("announcement after JSON:\n got %+v\nwant %+v", got, p)
	}
	// The maps and lists a re-run writes are the caller's own: editing them
	// edits nothing the prompt carries.
	got.commanderAnswers[uuid.UUID{9}] = false
	got.opponentPicks[0].answered = false
	if again := back.params(); !again.commanderAnswers[uuid.UUID{9}] || !again.opponentPicks[0].answered {
		t.Error("params() handed out the prompt's own data")
	}
}

// The helper's prompt is a restore point: restored while it is open, the
// answer re-runs the parked cast, the helper pays their share and the
// caster the rest.
func TestAssistPromptSurvivesARestore(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Swamp", 2)
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	before := assistPrompt(t, g)
	if before.PickThen != assistHelpKey {
		t.Fatalf("assist prompt names %q, want the engine's branch", before.PickThen)
	}

	g = restoreStrictForTest(t, g)
	me, helper = g.Seats[0], g.Seats[1]
	c := assistPrompt(t, g)
	if a := c.Assist; a == nil || a.Caster != me.ID || a.Spell != id || a.Generic != 3 || a.ManaCost != "{3}{B}" || a.Announce == nil {
		t.Fatalf("restored assist = %+v", a)
	}
	if !reflect.DeepEqual(c.Assist.Spend.Types, before.Assist.Spend.Types) || c.Assist.Spend.Purpose != SpendPurposeCast {
		t.Errorf("restored spend context = %+v, want the cast's", c.Assist.Spend)
	}
	if !assistInHand(me, id) {
		t.Fatal("the restored spell is not still in the caster's hand")
	}
	if err := g.ResolvePayAmount(c.ID, helper.ID, 3); err != nil {
		t.Fatalf("answer: %v", err)
	}
	item := g.StackMeta[id]
	if item == nil || assistInHand(me, id) {
		t.Fatal("the spell is not on the stack")
	}
	if got := assistTapped(g, theirs); got != 3 {
		t.Errorf("helper tapped %d Plains, want 3", got)
	}
	if got := assistTapped(g, mine); got != 1 {
		t.Errorf("caster tapped %d Swamps, want 1 (the {B})", got)
	}
	if item.Paid.AssistedBy != helper.ID || item.Paid.AssistPaid != 3 {
		t.Errorf("paid = by %v, %d; want the helper, 3", item.Paid.AssistedBy, item.Paid.AssistPaid)
	}
}

// The announcement's X, and a helper who leaves while the restored prompt
// is open: the cast goes on unassisted, with the X it was announced with.
func TestAssistPromptRestoredKeepsXAndDropsWithItsHelper(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, me, "Swamp", 4)
	assistLands(g, helper, "Plains", 2)
	id := assistSpell(me, "{X}{B}")
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AutoTap: true, AssistPlayer: helper.ID, XValue: 2}); err != nil {
		t.Fatal(err)
	}
	assistPrompt(t, g)

	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	if got := assistPrompt(t, g).Assist.Announce.Params.XValue; got != 2 {
		t.Fatalf("restored announcement X = %d, want 2", got)
	}
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	item := g.StackMeta[id]
	if item == nil || item.XValue != 2 || item.Paid.AssistPaid != 0 {
		t.Fatalf("stack item after the helper left = %+v, want the spell on the stack with X = 2 and no help", item)
	}
	if assistInHand(me, id) {
		t.Error("the spell is still in hand")
	}
}

// A tokens-attacking split is data: the template, the things to attack and
// the shares so far survive a restore, and the keyed tail runs once with
// the tokens that landed.
var (
	attackSplitTailSeen [][]uuid.UUID
	attackSplitTail     = RegisterTail(testEffectKeyPrefix+"attack-split-tail", func(_ *Game, o TailOutcome) error {
		attackSplitTailSeen = append(attackSplitTailSeen, o.IDs)
		return nil
	})
)

func TestAttackSplitSurvivesARestoreAtEachQuestion(t *testing.T) {
	attackSplitTailSeen = nil
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0]
	var targets []uuid.UUID
	for _, s := range g.Seats[1:] {
		targets = append(targets, s.ID)
	}
	g.WithWriteLock(func() {
		err := g.StartAttackSplitForEffect(AttackSplit{
			Controller: me.ID, Source: uuid.New(), Count: 5, Tapped: true, Label: "Split Probe",
			Template: Card{Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1, Colors: []string{"R"}},
			Targets:  targets, Tail: KeyedTail{Ref: attackSplitTail, Player: me.ID},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	c := payAmountChoice(g)
	if c == nil || c.PickThen != attackSplitKey || c.AttackSplit == nil || c.PayAmount.Max != 5 || c.PayAmount.Goal != 5 {
		t.Fatalf("first question = %+v, want 0..5 with goal 5 and the split's state", c)
	}

	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	c = payAmountChoice(g)
	if s := c.AttackSplit; s == nil || s.Template.Name != "Elemental" || s.Template.Power != 1 || !s.Tapped || len(s.Targets) != 3 || len(s.Shares) != 0 || s.Tail.IsZero() {
		t.Fatalf("restored split = %+v", s)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}

	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	c = payAmountChoice(g)
	if s := c.AttackSplit; s == nil || !reflect.DeepEqual(s.Shares, []int{2}) || c.PayAmount.Max != 3 || c.PayAmount.Goal != 0 {
		t.Fatalf("restored second question = %+v / %+v, want one share of 2 fixed and 3 left", c.AttackSplit, c.PayAmount)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	if payAmountChoice(g) != nil {
		t.Fatal("a third question: the last opponent takes the rest")
	}
	at := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			if !c.Tapped {
				t.Error("an Elemental entered untapped")
			}
			at[c.AttackingTarget]++
		}
	}
	if at[targets[0]] != 2 || at[targets[1]] != 1 || at[targets[2]] != 2 {
		t.Errorf("attacking %v, want 2, 1 and 2 at the three opponents in asking order", at)
	}
	if len(attackSplitTailSeen) != 1 || len(attackSplitTailSeen[0]) != 5 {
		t.Errorf("the tail ran %v, want once with the five tokens", attackSplitTailSeen)
	}
}

// With one thing to attack, or none, nothing is asked: the tokens are
// made at once, attacking it, and the tail runs.
func TestAttackSplitAsksNothingWithOneOrNoTarget(t *testing.T) {
	for _, targets := range [][]uuid.UUID{nil, {uuid.New()}} {
		attackSplitTailSeen = nil
		g := newActiveGameWithSeats(t, 3)
		me := g.Seats[0]
		if len(targets) == 1 {
			targets = []uuid.UUID{g.Seats[1].ID}
		}
		g.WithWriteLock(func() {
			err := g.StartAttackSplitForEffect(AttackSplit{
				Controller: me.ID, Source: uuid.New(), Count: 2, Label: "Split Probe",
				Template: Card{Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1},
				Targets:  targets, Tail: KeyedTail{Ref: attackSplitTail, Player: me.ID},
			})
			if err != nil {
				t.Fatal(err)
			}
		})
		if payAmountChoice(g) != nil {
			t.Fatalf("%d targets: asked a question", len(targets))
		}
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Elemental" && c.Controller == me.ID {
				n++
				if len(targets) == 1 && c.AttackingTarget != targets[0] {
					t.Errorf("the token attacks %v, want %v", c.AttackingTarget, targets[0])
				}
			}
		}
		if n != 2 || len(attackSplitTailSeen) != 1 {
			t.Errorf("%d targets: %d tokens, tail runs %v; want 2 and one run", len(targets), n, attackSplitTailSeen)
		}
	}
}
