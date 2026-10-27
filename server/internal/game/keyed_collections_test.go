package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// keyed_collections_test.go pins #4290's second half: a choose_cards run
// whose legs are keyed picks (ChooseCardsRunThenKeyedForEffect), a
// permanent pick run whose legs carry a registered set rule, and ADR
// 0175's ordered collector as plain data (ordered_picks_keyed.go). Each
// restore test answers the first seat, takes a STRICT restore through
// JSON with the next still open, and answers on the restored table.

var (
	// A rule over what a seat picks, registered with its leg: no card named
	// "Bad".
	testNoBadLeg = RegisterChooseCardsRunLegRule(testEffectKeyPrefix+"no-bad-leg", func(picked []Card) bool {
		for _, c := range picked {
			if c.Name == "Bad" {
				return false
			}
		}
		return true
	})
	// The same rule on a keyed permanent-pick run.
	testNoBadPickRun = RegisterPromptRunThenRule(testEffectKeyPrefix+"no-bad-pick-run", func(picked []Card) bool {
		for _, c := range picked {
			if c.Name == "Bad" {
				return false
			}
		}
		return true
	}, recordKeyedRun)
	// An ordered collection's leg with a set rule: the picks must be bears.
	testBearsOnlyLeg = RegisterOrderedPickLegRule(testEffectKeyPrefix+"bears-only-leg", func(picked []Card) bool {
		for _, c := range picked {
			if c.Name != "Bears" {
				return false
			}
		}
		return true
	})
)

// chooseCardsPickFor is the open choose_cards prompt `seat` owes.
func chooseCardsPickFor(g *Game, seat uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseCards && c.Chooser == seat {
			return c
		}
	}
	return nil
}

// namedHandFor gives `p` a hand of the named cards (a name per card) and returns
// their IDs.
func namedHandFor(t *testing.T, g *Game, p *Player, names ...string) []uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	var ids []uuid.UUID
	for _, n := range names {
		ids = append(ids, handCardForDiscard(t, g, p, n))
	}
	return ids
}

// keyedChooseRunTable is three seats each asked for one card of their own
// hand by a keyed choose_cards run, the third leg carrying a set rule.
func keyedChooseRunTable(t *testing.T) (*Game, [][]uuid.UUID, uuid.UUID) {
	t.Helper()
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	source := departureTestSource(g, g.Seats[3].ID, "Ill-Gotten Gains")
	hands := make([][]uuid.UUID, 3)
	var legs []ChooseCardsPrompt
	for i, p := range g.Seats[:3] {
		hands[i] = namedHandFor(t, g, p, "Good", "Bad")
		leg := ChooseCardsPrompt{
			Chooser: p.ID, FromPlayer: p.ID, Source: source,
			Question: "choose a card", Cards: hands[i], Min: 1, Max: 1, Zone: ZoneHand,
		}
		if i == 2 {
			leg.ThenKey = testNoBadLeg
		}
		legs = append(legs, leg)
	}
	asked := &AskedItem{ID: uuid.New(), Kind: StackItemSpell, Controller: g.Seats[3].ID, Owner: g.Seats[3].ID,
		SourceCardID: source, Label: "Ill-Gotten Gains", XValue: 3}
	g.WithWriteLock(func() {
		n, err := g.ChooseCardsRunThenKeyedForEffect(legs, testKeyedRun, source,
			RunData{Carry: []uuid.UUID{g.Seats[3].ID}, Nums: []int{7, 9}, Asked: asked})
		if err != nil || n != 3 {
			t.Fatalf("ChooseCardsRunThenKeyedForEffect = %d, %v, want 3 legs", n, err)
		}
	})
	return g, hands, source
}

// TestAKeyedChooseCardsRunIsARestorePointWithALegAnswered: three seats are
// asked, the first answers, the table restores strictly with the other two
// open, and the run pays out once after the last, with the data it carried.
func TestAKeyedChooseCardsRunIsARestorePointWithALegAnswered(t *testing.T) {
	g, hands, source := keyedChooseRunTable(t)
	seats := []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID}
	if err := g.ResolveChooseCards(chooseCardsPickFor(g, seats[0]).ID, seats[0], hands[0][:1]); err != nil {
		t.Fatalf("first leg: %v", err)
	}
	if keyedRunCall.ran != 0 {
		t.Fatal("the run paid out with two legs open")
	}
	snap := g.CaptureSnapshot()
	if len(snap.PromptRuns) != 1 || snap.PromptRuns[0].Outstanding != 2 || len(snap.PromptRuns[0].Landed) != 1 {
		t.Fatalf("the snapshot's run is %+v, want two legs outstanding and the first seat's pick", snap.PromptRuns)
	}
	if got := snap.PromptRuns[0]; len(got.Nums) != 2 || got.Asked == nil || got.Asked.Label != "Ill-Gotten Gains" {
		t.Fatalf("the run lost its data on disk: %+v", got)
	}

	g = strictRestore(t, g)
	if err := g.ResolveChooseCards(chooseCardsPickFor(g, seats[1]).ID, seats[1], hands[1][:1]); err != nil {
		t.Fatalf("second leg on the restored table: %v", err)
	}
	if keyedRunCall.ran != 0 {
		t.Fatal("the restored run paid out with a leg still open")
	}
	// The third leg's set rule survived the restore: "Bad" is refused.
	last := chooseCardsPickFor(g, seats[2])
	if err := g.ResolveChooseCards(last.ID, seats[2], hands[2][1:]); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("a set the registered rule refuses: err = %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveChooseCards(last.ID, seats[2], hands[2][:1]); err != nil {
		t.Fatalf("third leg: %v", err)
	}
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored run ran %d times, want 1", keyedRunCall.ran)
	}
	got := keyedRunCall.got
	if got.Source != source || len(got.Carry) != 1 || len(got.Nums) != 2 || got.Nums[1] != 9 {
		t.Errorf("the run was handed %+v, want its source, carry and numbers", got)
	}
	if item := got.Item(); item == nil || item.Label != "Ill-Gotten Gains" || item.XValue != 3 {
		t.Errorf("the run's asking item is %+v, want the one it carried", item)
	}
	picks := got.Picks()
	for i, s := range seats {
		if by := picks.By(s); len(by) != 1 || by[0] != hands[i][0] {
			t.Errorf("seat %d's pick is %v, want %v", i, by, hands[i][0])
		}
	}
}

// TestAKeyedChooseCardsRunWithAnUnknownKeyIsRefused: the run's key, and a
// leg's, are checked at restore.
func TestAKeyedChooseCardsRunWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, _ := keyedChooseRunTable(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PromptRuns[0].Key = testEffectKeyPrefix + "never-registered"
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown run key: err = %v, want ErrUnknownEffectKey", err)
	}
	snap = throughJSON(t, g.CaptureSnapshot())
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].PickThen == testNoBadLeg.Key() {
			snap.PendingChoices[i].PickThen = testEffectKeyPrefix + "never-registered"
		}
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown leg key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestASeatLeavingSettlesItsKeyedChooseCardsLeg: departure settles a keyed
// choose_cards leg with nothing chosen, on a restored table, and the run
// still pays out after the last answer.
func TestASeatLeavingSettlesItsKeyedChooseCardsLeg(t *testing.T) {
	g, hands, _ := keyedChooseRunTable(t)
	seats := []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID}
	g = strictRestore(t, g)
	if err := g.Concede(seats[1]); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if err := g.ResolveChooseCards(chooseCardsPickFor(g, seats[0]).ID, seats[0], hands[0][:1]); err != nil {
		t.Fatal(err)
	}
	if err := g.ResolveChooseCards(chooseCardsPickFor(g, seats[2]).ID, seats[2], hands[2][:1]); err != nil {
		t.Fatal(err)
	}
	if keyedRunCall.ran != 1 || keyedRunCall.got.Picks().Count() != 2 {
		t.Errorf("the run ran %d times with %v, want once with the two seats that answered", keyedRunCall.ran, keyedRunCall.got.Landed)
	}
}

// TestAKeyedPermanentPickRunKeepsItsRegisteredSetRule: the rule registered
// with the run (RegisterPromptRunThenRule) is applied to a leg before and
// after a strict restore, and a leg answered on the restored table settles
// the restored run.
func TestAKeyedPermanentPickRunKeepsItsRegisteredSetRule(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Tragic Arrogance")
	mine := pushFodderForTest(g, me.ID, 2)
	theirs := pushFodderForTest(g, opp.ID, 1)
	var bad uuid.UUID
	g.WithWriteLock(func() {
		c := NewCard("Bad", opp.ID)
		c.TypeLine = "Artifact"
		g.Battlefield.PushTop(c)
		bad = c.InstanceID
	})
	candidates := func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var out []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == of && (c.Name == "Fodder" || c.Name == "Bad") {
				out = append(out, c.InstanceID)
			}
		}
		return out, 1, 1
	}
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: me.ID, Source: source, Question: "choose one",
			Of: []uuid.UUID{me.ID, opp.ID}, Candidates: candidates,
		}, testNoBadPickRun, RunData{}); err != nil {
			t.Fatalf("PermanentsPickedThenKeyedForEffect: %v", err)
		}
	})
	own := pendingChoiceOfKind(g, PendingChoiceOwnPermanents)
	if err := g.ResolveOwnPermanents(own.ID, me.ID, mine[:1]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	g = strictRestore(t, g)
	their := pendingChoiceOfKind(g, PendingChoiceTheirPermanents)
	if their == nil {
		t.Fatal("the restored table lost the their_permanents leg")
	}
	if err := g.ResolveTheirPermanents(their.ID, me.ID, []uuid.UUID{bad}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("a set the registered rule refuses: err = %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveTheirPermanents(their.ID, me.ID, theirs[:1]); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored run ran %d times, want 1", keyedRunCall.ran)
	}
}

// keyedEdictTable is the ordered collector on four seats: "each other
// player sacrifices a creature of their choice", every one holding two
// bears, with the caster on seat 3 and a keyed run waiting.
func keyedEdictTable(t *testing.T) (*Game, [][]uuid.UUID, uuid.UUID) {
	t.Helper()
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	source := departureTestSource(g, g.Seats[3].ID, "Syphon Flesh")
	bears := make([][]uuid.UUID, 4)
	for i, p := range g.Seats {
		bears[i] = []uuid.UUID{pushBear(g, p.ID), pushBear(g, p.ID)}
	}
	g.WithWriteLock(func() {
		if err := g.EachPlayerSacrificesThenKeyedForEffect(source, g.Seats[3].ID, nil, "sacrifice a permanent",
			testKeyedRun, RunData{Carry: []uuid.UUID{g.Seats[3].ID}, Nums: []int{4}}); err != nil {
			t.Fatalf("EachPlayerSacrificesThenKeyedForEffect: %v", err)
		}
	})
	return g, bears, source
}

// orderedPromptFor is the open keyed ordered-collection prompt `seat` owes.
func orderedPromptFor(g *Game, seat uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.ordered != nil && c.Chooser == seat {
			return c
		}
	}
	return nil
}

// TestAKeyedEdictCollectionIsARestorePointWithASeatAnswered is the headline
// for ADR 0175: seats choose one at a time in APNAP order, seeing the
// earlier picks; the table restores strictly with the first seat answered,
// nothing has been sacrificed, and the picks leave together after the
// last seat on the restored table.
func TestAKeyedEdictCollectionIsARestorePointWithASeatAnswered(t *testing.T) {
	g, bears, source := keyedEdictTable(t)
	order := g.apnapSeatsExceptLocked(g.Seats[3].ID)
	if len(order) != 3 {
		t.Fatalf("expected three seats asked, got %v", order)
	}
	index := func(id uuid.UUID) int {
		for i, p := range g.Seats {
			if p.ID == id {
				return i
			}
		}
		return -1
	}
	if only := len(g.PendingChoices); only != 1 {
		t.Fatalf("one seat is asked at a time: %d prompts open", only)
	}
	first := orderedPromptFor(g, order[0])
	if first == nil || first.PickThen != orderedPickLegKey || first.Kind != PendingChoiceChooseCards {
		t.Fatalf("the first seat's prompt is %+v, want the engine's ordered leg", first)
	}
	if len(first.EarlierPermanentChoicesForEffect()) != 0 {
		t.Error("the first seat saw earlier picks")
	}
	firstPick := bears[index(order[0])][0]
	if err := g.ResolveChooseCards(first.ID, order[0], []uuid.UUID{firstPick}); err != nil {
		t.Fatalf("first seat: %v", err)
	}
	if findBattlefieldCard(g, firstPick) == nil {
		t.Fatal("a pick left the battlefield before the last seat had chosen")
	}
	if keyedRunCall.ran != 0 {
		t.Fatal("the run ran with two seats still to choose")
	}

	g = strictRestore(t, g)
	second := orderedPromptFor(g, order[1])
	if second == nil {
		t.Fatal("the restored table lost the second seat's prompt")
	}
	earlier := second.EarlierPermanentChoicesForEffect()
	if len(earlier) != 1 || earlier[0].Seat != order[0] || len(earlier[0].Cards) != 1 || earlier[0].Cards[0] != firstPick {
		t.Errorf("the second seat sees %+v, want the first seat's pick (CR 101.4b)", earlier)
	}
	secondPick := bears[index(order[1])][1]
	if err := g.ResolveChooseCards(second.ID, order[1], []uuid.UUID{secondPick}); err != nil {
		t.Fatalf("second seat: %v", err)
	}
	third := orderedPromptFor(g, order[2])
	if third == nil {
		t.Fatal("the third seat was not asked")
	}
	thirdPick := bears[index(order[2])][0]
	if err := g.ResolveChooseCards(third.ID, order[2], []uuid.UUID{thirdPick}); err != nil {
		t.Fatalf("third seat: %v", err)
	}
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored collection ran its continuation %d times, want 1", keyedRunCall.ran)
	}
	for _, id := range []uuid.UUID{firstPick, secondPick, thirdPick} {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("%s was chosen and should have been sacrificed", id)
		}
	}
	got := keyedRunCall.got
	if got.Source != source || len(got.Carry) != 1 || len(got.Nums) != 1 || got.Nums[0] != 4 {
		t.Errorf("the continuation was handed %+v, want its source, carry and number", got)
	}
	if got.Sacrifices().Count() != 3 || len(got.Landed) != 3 {
		t.Errorf("the answer is %v, want one entry per seat asked and three sacrifices", got.Landed)
	}
	for i, id := range order {
		if by := got.Sacrifices().By(id); len(by) != 1 || by[0] != []uuid.UUID{firstPick, secondPick, thirdPick}[i] {
			t.Errorf("seat %d's entry is %v", i, by)
		}
	}
}

// TestAKeyedEdictCollectionWithAnUnknownKeyIsRefused: the collection names
// its run continuation, and the prompt the engine's leg key.
func TestAKeyedEdictCollectionWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, _ := keyedEdictTable(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	found := false
	for i := range snap.PendingChoices {
		if o := snap.PendingChoices[i].Ordered; o != nil {
			o.Key = testEffectKeyPrefix + "never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no ordered collection")
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown collection key: err = %v, want ErrUnknownEffectKey", err)
	}
	snap = throughJSON(t, g.CaptureSnapshot())
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown leg key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestAKeyedEdictCollectionBindsPicksToTheObjectsTheyWere: a permanent
// chosen by the first seat that has left and come back before the last seat
// answers is a new object (CR 400.7) and is not sacrificed, on a restored
// table too.
func TestAKeyedEdictCollectionBindsPicksToTheObjectsTheyWere(t *testing.T) {
	g, bears, _ := keyedEdictTable(t)
	order := g.apnapSeatsExceptLocked(g.Seats[3].ID)
	index := func(id uuid.UUID) int {
		for i, p := range g.Seats {
			if p.ID == id {
				return i
			}
		}
		return -1
	}
	first := orderedPromptFor(g, order[0])
	firstPick := bears[index(order[0])][0]
	if err := g.ResolveChooseCards(first.ID, order[0], []uuid.UUID{firstPick}); err != nil {
		t.Fatal(err)
	}
	g = strictRestore(t, g)
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, firstPick); c != nil {
			c.ObjectEpoch++ // it left and came back
		}
	})
	second := orderedPromptFor(g, order[1])
	if err := g.ResolveChooseCards(second.ID, order[1], []uuid.UUID{bears[index(order[1])][0]}); err != nil {
		t.Fatal(err)
	}
	third := orderedPromptFor(g, order[2])
	if err := g.ResolveChooseCards(third.ID, order[2], []uuid.UUID{bears[index(order[2])][0]}); err != nil {
		t.Fatal(err)
	}
	if findBattlefieldCard(g, firstPick) == nil {
		t.Error("a permanent that left and came back was sacrificed as the object it was chosen as")
	}
	if got := keyedRunCall.got.Sacrifices(); got.Sacrificed(order[0]) || got.Count() != 2 {
		t.Errorf("the answer is %v, want only the two picks that are still the same objects", got)
	}
}

// TestASeatLeavingSettlesItsKeyedEdictLeg: a seat that concedes with its
// turn to choose open chooses nothing (CR 800.4a) and the collection goes
// on to the next seat, on a restored table.
func TestASeatLeavingSettlesItsKeyedEdictLeg(t *testing.T) {
	g, bears, _ := keyedEdictTable(t)
	order := g.apnapSeatsExceptLocked(g.Seats[3].ID)
	index := func(id uuid.UUID) int {
		for i, p := range g.Seats {
			if p.ID == id {
				return i
			}
		}
		return -1
	}
	g = strictRestore(t, g)
	if err := g.Concede(order[0]); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	second := orderedPromptFor(g, order[1])
	if second == nil {
		t.Fatal("the collection did not go on to the next seat")
	}
	if err := g.ResolveChooseCards(second.ID, order[1], []uuid.UUID{bears[index(order[1])][0]}); err != nil {
		t.Fatal(err)
	}
	third := orderedPromptFor(g, order[2])
	if err := g.ResolveChooseCards(third.ID, order[2], []uuid.UUID{bears[index(order[2])][0]}); err != nil {
		t.Fatal(err)
	}
	if keyedRunCall.ran != 1 || keyedRunCall.got.Sacrifices().Count() != 2 {
		t.Errorf("the run ran %d times with %v, want once with the two seats that chose", keyedRunCall.ran, keyedRunCall.got.Landed)
	}
}

// TestAFireAndForgetEdictCollectionIsARestorePoint: "each player sacrifices
// a creature" with nothing waiting (Grave Pact, Liliana, Dreadhorde
// General's −4) collects through the same plain data, so the table
// restores strictly between two seats and the picks still leave together.
func TestAFireAndForgetEdictCollectionIsARestorePoint(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	source := departureTestSource(g, g.Seats[3].ID, "Grave Pact")
	bears := make([][]uuid.UUID, 4)
	for i, p := range g.Seats {
		bears[i] = []uuid.UUID{pushBear(g, p.ID), pushBear(g, p.ID)}
	}
	var asked int
	g.WithWriteLock(func() { asked = g.EachPlayerSacrificesForEffect(source, g.Seats[3].ID, nil, "sacrifice a permanent") })
	if asked != 3 {
		t.Fatalf("asked %d seats, want 3", asked)
	}
	order := g.apnapSeatsExceptLocked(g.Seats[3].ID)
	index := func(id uuid.UUID) int {
		for i, p := range g.Seats {
			if p.ID == id {
				return i
			}
		}
		return -1
	}
	first := orderedPromptFor(g, order[0])
	if first == nil {
		t.Fatalf("no collection prompt for the first seat: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(first.ID, order[0], []uuid.UUID{bears[index(order[0])][0]}); err != nil {
		t.Fatal(err)
	}
	g = strictRestore(t, g)
	for _, seat := range order[1:] {
		c := orderedPromptFor(g, seat)
		if c == nil {
			t.Fatalf("seat %s was not asked on the restored table", seat)
		}
		if err := g.ResolveChooseCards(c.ID, seat, []uuid.UUID{bears[index(seat)][0]}); err != nil {
			t.Fatal(err)
		}
	}
	for _, seat := range order {
		if findBattlefieldCard(g, bears[index(seat)][0]) != nil {
			t.Errorf("seat %s's pick was not sacrificed", seat)
		}
		if findBattlefieldCard(g, bears[index(seat)][1]) == nil {
			t.Errorf("seat %s's other bear was sacrificed", seat)
		}
	}
	if findBattlefieldCard(g, bears[3][0]) == nil || findBattlefieldCard(g, bears[3][1]) == nil {
		t.Error("the caster, excepted, lost a creature")
	}
}

// TestAKeyedCollectionLegRuleSurvivesARestore: an ordered collection's leg
// with its own registered rule (Liliana, Dreadhorde General's −9) keeps it
// across a strict restore.
func TestAKeyedCollectionLegRuleSurvivesARestore(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	me := g.Seats[1]
	source := departureTestSource(g, g.Seats[0].ID, "Liliana")
	bear, fodder := pushBear(g, me.ID), pushFodderForTest(g, me.ID, 1)[0]
	g.WithWriteLock(func() {
		err := g.ChoosePermanentsInOrderKeyedForEffect([]ChooseCardsPrompt{{
			Chooser: me.ID, FromPlayer: me.ID, Source: source, Question: "keep a bear",
			Cards: []uuid.UUID{bear, fodder}, Min: 1, Max: 1, Zone: ZoneBattlefield, ThenKey: testBearsOnlyLeg,
		}}, testKeyedRun, source, RunData{})
		if err != nil {
			t.Fatalf("ChoosePermanentsInOrderKeyedForEffect: %v", err)
		}
	})
	g = strictRestore(t, g)
	c := orderedPromptFor(g, me.ID)
	if c == nil || c.PickThen != testBearsOnlyLeg.Key() {
		t.Fatalf("the restored prompt is %+v, want the leg with its rule", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fodder}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("a set the leg's rule refuses: err = %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	if keyedRunCall.ran != 1 || keyedRunCall.got.Picks().Count() != 1 {
		t.Errorf("the run ran %d times with %v, want once with the bear", keyedRunCall.ran, keyedRunCall.got.Landed)
	}
}

// TestAnOrderedCollectionRefusesALegWithAnUnregisteredKey keeps the keyed
// door closed to anything a restore could not carry. (A leg cannot hold a
// closure at all since #4361 S7b deleted ChooseCardsPrompt.Then and
// .Validate.)
func TestAnOrderedCollectionRefusesALegWithAnUnregisteredKey(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		leg := ChooseCardsPrompt{Chooser: me.ID, Cards: []uuid.UUID{bear}, Min: 0, Max: 1, ThenKey: ChooseCardsRunLeg}
		if err := g.ChoosePermanentsInOrderKeyedForEffect([]ChooseCardsPrompt{leg}, testKeyedRun, uuid.Nil, RunData{}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a leg naming a run's leg key: err = %v, want ErrInvalidParam", err)
		}
		if err := g.ChoosePermanentsInOrderKeyedForEffect(nil, PromptRunThen{}, uuid.Nil, RunData{}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("no continuation: err = %v, want ErrInvalidParam", err)
		}
	})
}
