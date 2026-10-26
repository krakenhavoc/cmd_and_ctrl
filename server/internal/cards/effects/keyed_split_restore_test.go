package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_split_restore_test.go — #4294: the last two closures behind a
// number or a pick are data. Myriad's split between an opponent and their
// planeswalkers is a number prompt carrying the settled creation, and
// Calamity, Galloping Inferno's pick and split carry the asking item and
// the round. Each test takes a STRICT restore while a question is open and
// answers it on the restored table.

// myriadWalker is a planeswalker `owner` controls, so that their copy has
// two things to attack and the split asks (CR 508.4).
func myriadWalker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker",
		Owner: owner, Controller: owner, StartingLoyalty: 6, Counters: map[string]int{game.CounterLoyalty: 6}})
}

// splitPrompt is the open attack-split question `me` owes, failing the
// test unless it holds the key `key` and the engine's split data.
func splitPrompt(t *testing.T, g *game.Game, me uuid.UUID, key string) *game.PendingChoice {
	t.Helper()
	c := numberPrompt(t, g, me)
	if c.PickThen != key || c.AttackSplit == nil || c.AttackSplit.Tail.IsZero() {
		t.Fatalf("the split prompt = key %q split %+v, want %q with a tail", c.PickThen, c.AttackSplit, key)
	}
	return c
}

// sendOneTo answers a one-copy question so that the copy attacks `want`:
// all of it when the question asks about `want`, none of it otherwise.
func sendOneTo(t *testing.T, g *game.Game, me uuid.UUID, c *game.PendingChoice, want uuid.UUID) {
	t.Helper()
	n := 0
	if c.AttackSplit.Targets[0] == want {
		n = 1
	}
	if err := g.ResolvePayAmount(c.ID, me, n); err != nil {
		t.Fatalf("answer %d: %v", n, err)
	}
}

// Myriad with two opponents who each have a planeswalker asks twice, once
// per opponent. A restore at the first question, and another at the second
// with the first opponent's copy already placed, make the same copies and
// schedule the same exile as a table that was never restored.
func TestMyriadSplitIsARestorePointAtEachQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me, defender := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3)
	walkers := map[uuid.UUID]uuid.UUID{}
	for _, s := range g.Seats[2:] {
		walkers[s.ID] = myriadWalker(g, s.ID)
	}
	declareAttack(t, g, defender.ID, knight)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)

	c := splitPrompt(t, g, me.ID, "token/settled-split")
	first := c.AttackSplit.Targets
	if len(first) != 2 || c.PayAmount.Max != 1 || c.PayAmount.Goal != 1 {
		t.Fatalf("first question = targets %v, %+v; want one copy and two things to attack", first, c.PayAmount)
	}
	if got := len(s58Named(g, me.ID, "Knight")); got != 1 {
		t.Fatalf("%d Knights before the first answer, want only the attacker: nothing enters before every share is placed", got)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = splitPrompt(t, g, me.ID, "token/settled-split")
	if len(c.AttackSplit.Targets) != 2 || c.AttackSplit.Targets[0] != first[0] || c.AttackSplit.Label == "" {
		t.Fatalf("restored first question = %+v, want the same two things to attack, in the same order", c.AttackSplit)
	}
	// The first opponent's copy goes to their planeswalker.
	var firstWalker uuid.UUID
	for _, w := range walkers {
		for _, id := range first {
			if id == w {
				firstWalker = w
			}
		}
	}
	if firstWalker == uuid.Nil {
		t.Fatalf("the first question's targets %v name no planeswalker", first)
	}
	sendOneTo(t, g, me.ID, c, firstWalker)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = splitPrompt(t, g, me.ID, "token/settled-split")
	if c.AttackSplit.Targets[0] == first[0] || c.AttackSplit.Targets[1] == first[1] {
		t.Fatalf("restored second question asks about %v again, want the other opponent's things", c.AttackSplit.Targets)
	}
	if got := len(s58Named(g, me.ID, "Knight")); got != 1 {
		t.Fatalf("%d Knights between the questions, want only the attacker", got)
	}
	// The second opponent's copy goes at the player.
	var secondPlayer uuid.UUID
	for _, s := range g.Seats[2:] {
		for _, id := range c.AttackSplit.Targets {
			if id == s.ID {
				secondPlayer = s.ID
			}
		}
	}
	if secondPlayer == uuid.Nil {
		t.Fatalf("the second question's targets %v name no player", c.AttackSplit.Targets)
	}
	sendOneTo(t, g, me.ID, c, secondPlayer)
	if payAmountPrompt(g, me.ID) != nil {
		t.Fatal("a third question: every share is placed")
	}

	attacking := map[uuid.UUID]int{}
	var copies []uuid.UUID
	for _, k := range s58Named(g, me.ID, "Knight") {
		if k.InstanceID != knight {
			attacking[k.AttackingTarget]++
			copies = append(copies, k.InstanceID)
			if !k.Tapped || !IsToken(k) {
				t.Errorf("copy %s should be a tapped token", k.InstanceID)
			}
		}
	}
	if len(copies) != 2 || attacking[firstWalker] != 1 || attacking[secondPlayer] != 1 {
		t.Fatalf("copies attack %v, want one at the first opponent's planeswalker and one at the second opponent", attacking)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers, want the one exile of both copies", len(g.DelayedTriggers))
	}
	if got, want := g.DelayedTriggers[0].Label, "Legion Loyalty — exile the myriad tokens"; got != want {
		t.Errorf("the exile is labelled %q, want %q: the tail carries the granting card's name", got, want)
	}
	// "Exile the tokens at end of combat": the keyed tail, told the copies
	// that landed, scheduled it on the restored table.
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	for _, id := range copies {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("copy %s survived end of combat", id)
		}
	}
	if _, ok := battlefieldCardByID(g, knight); !ok {
		t.Error("the attacker itself was exiled")
	}
}

// A token doubler settles before the split: two copies for one opponent,
// asked about once the window has settled, and a restore at that question
// keeps the doubled count.
func TestMyriadDoubledSplitIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, defender, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3)
	pushTokenReplacementCard(g, parallelLivesOracle, "Parallel Lives", "Enchantment", me.ID)
	walker := myriadWalker(g, other.ID)
	declareAttack(t, g, defender.ID, knight)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, false)

	c := splitPrompt(t, g, me.ID, "token/settled-split")
	if c.PayAmount.Max != 2 {
		t.Fatalf("first question = %+v, want the two doubled copies", c.PayAmount)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = splitPrompt(t, g, me.ID, "token/settled-split")
	if c.PayAmount.Max != 2 {
		t.Fatalf("restored question = %+v, want the two doubled copies", c.PayAmount)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	if payAmountPrompt(g, me.ID) != nil {
		t.Fatal("a second question: the last thing takes the rest")
	}
	attacks := map[uuid.UUID]int{}
	for _, k := range s58Named(g, me.ID, "Knight") {
		if k.InstanceID != knight {
			attacks[k.AttackingTarget]++
		}
	}
	first := c.AttackSplit.Targets[0]
	rest := c.AttackSplit.Targets[1]
	if attacks[first] != 1 || attacks[rest] != 1 || len(attacks) != 2 || (first != walker && rest != walker) {
		t.Errorf("copies attack %v, want one each at %s and %s", attacks, first, rest)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("%d delayed triggers, want the one exile of both copies", len(g.DelayedTriggers))
	}
}

// A binary that lacks the number branch the myriad split names, or the
// tail it carries, refuses the file rather than ask a question it cannot
// answer.
func TestMyriadSplitWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, defender := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3)
	myriadWalker(g, g.Seats[2].ID)
	declareAttack(t, g, defender.ID, knight)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, false)
	splitPrompt(t, g, me.ID, "token/settled-split")

	refuseUnknownKey(t, g, "token/settled-split")

	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if s := snap.PendingChoices[i].AttackSplit; s != nil && s.Tail != nil {
			if s.Tail.Key != "attacking-tokens/myriad-exile-the-copies" || s.Tail.Word == "" || s.Tail.Asked == nil {
				t.Fatalf("the split's tail = %+v, want the myriad exile with the granting card and the asking item", s.Tail)
			}
			s.Tail.Key = "attacking-tokens/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no split with a tail")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown tail: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- Calamity, Galloping Inferno ------------------------------------------

// keyedPick is the open card pick `chooser` owes, failing the test unless
// it names `key`.
func keyedPick(t *testing.T, g *game.Game, chooser uuid.UUID, key string) *game.PendingChoice {
	t.Helper()
	c := latestChooseCards(g, chooser)
	if c == nil || c.PickThen != key {
		t.Fatalf("the open pick = %+v, want one keyed %q", c, key)
	}
	return c
}

// Calamity asks twice: which saddler to copy, then (with two things to
// attack) how many of the copy attack each. A restore at the first pick,
// at the first split and at the second round's pick makes two sacrificed
// copies, as a table that was never restored does. The pick carries the
// trigger's item and the round; the split's tail carries both on.
func TestCalamityPickAndSplitAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)
	pickKey := calamityPick.Key()

	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil || choice.PickThen != pickKey {
		t.Fatalf("the first pick = %+v, want one keyed %q", choice, pickKey)
	}
	if choice.Asked == nil || len(choice.PickNums) != 1 || choice.PickNums[0] != 2 {
		t.Fatalf("the first pick carries asked=%+v nums=%v, want the trigger's item and two rounds", choice.Asked, choice.PickNums)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	choice = keyedPick(t, g, me.ID, pickKey)
	if choice.Asked == nil || choice.Asked.SourceCardID != calamity || len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("restored first pick = %+v, want the saddler and Calamity's trigger as the asking item", choice)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("first choice: %v", err)
	}

	// The copy has two or more things to attack: the split is a restore
	// point, and its tail carries the asking item.
	c := splitPrompt(t, g, me.ID, "token/attack-split")
	if c.AttackSplit.Tail.Asked == nil || c.AttackSplit.Tail.N != 2 {
		t.Fatalf("the split's tail = %+v, want the asking item and two rounds", c.AttackSplit.Tail)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = splitPrompt(t, g, me.ID, "token/attack-split")
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); err != nil {
		t.Fatalf("first split: %v", err)
	}
	answerAttackSplit(t, g, me, 0)
	if got := len(calamityTokens(g, me.ID, "Calamity Probe Saddler")); got != 1 {
		t.Fatalf("%d copies after the first round, want 1", got)
	}

	// The second round's pick was asked from the first copy's tail.
	choice = keyedPick(t, g, me.ID, pickKey)
	if len(choice.PickNums) != 1 || choice.PickNums[0] != 1 {
		t.Fatalf("the second pick carries nums=%v, want the one round left", choice.PickNums)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	choice = keyedPick(t, g, me.ID, pickKey)
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("second choice: %v", err)
	}
	answerAttackSplit(t, g, me, 0)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("a third round was offered")
	}
	tokens := calamityTokens(g, me.ID, "Calamity Probe Saddler")
	if len(tokens) != 2 {
		t.Fatalf("%d copies, want 2", len(tokens))
	}
	var ids []uuid.UUID
	for _, tok := range tokens {
		if !tok.Tapped || tok.AttackingTarget == uuid.Nil {
			t.Errorf("copy %s: tapped=%v attacking=%s, want tapped and attacking", tok.InstanceID, tok.Tapped, tok.AttackingTarget)
		}
		ids = append(ids, tok.InstanceID)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("copy %s survived the beginning of the next end step", id)
		}
	}
}

// The restored pick still refuses a saddler that left, as the closure
// form's zone re-check did: only the candidates still on the battlefield
// can be chosen, and a withdrawn pick ends the ability.
func TestCalamityRestoredPickRechecksTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)
	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil || choice.ChooseLiveZone != game.ZoneBattlefield {
		t.Fatalf("the pick = %+v, want a keyed pick that re-checks the battlefield", choice)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	choice = keyedPick(t, g, me.ID, calamityPick.Key())
	g.WithWriteLock(func() {
		if err := (BounceToHand{Target: saddler}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err == nil {
		t.Error("a saddler that left the battlefield was accepted")
	}
}

// The second round is asked from the first copy's tail, against the
// trigger's item rebuilt from data: it still knows what Calamity attacked
// (the first thing offered) and which saddlers it has, restored or not.
func TestCalamitySecondRoundReadsTheAskingItemFromTheTail(t *testing.T) {
	g := newCatalogGame(t)
	me, defender := g.Seats[0], g.Seats[2]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)

	choice := calamityAttackPrompt(t, g, me, defender.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no pick")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	c := splitPrompt(t, g, me.ID, "token/attack-split")
	if c.AttackSplit.Targets[0] != defender.ID {
		t.Fatalf("round one offers %v first, want Calamity's defender %s", c.AttackSplit.Targets, defender.ID)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	choice = keyedPick(t, g, me.ID, calamityPick.Key())
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("round two offers %v, want the saddler", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	c = splitPrompt(t, g, me.ID, "token/attack-split")
	if c.AttackSplit.Targets[0] != defender.ID {
		t.Errorf("round two offers %v first, want Calamity's defender %s: the tail lost the trigger's item", c.AttackSplit.Targets, defender.ID)
	}
}

// A binary that lacks the pick's key, or the tail the split carries,
// refuses the file.
func TestCalamityWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calamity := pushCalamity(g, me.ID)
	saddler := pushProbeSaddler(g, me.ID)
	choice := calamityAttackPrompt(t, g, me, opp.ID, calamity, saddler)
	if choice == nil {
		t.Fatal("no pick")
	}
	refuseUnknownKey(t, g, calamityPick.Key())

	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatal(err)
	}
	splitPrompt(t, g, me.ID, "token/attack-split")
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if s := snap.PendingChoices[i].AttackSplit; s != nil && s.Tail != nil {
			s.Tail.Key = "attacking-tokens/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no split with a tail")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown tail: err = %v, want ErrUnknownEffectKey", err)
	}
}
