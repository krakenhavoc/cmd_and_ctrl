package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_spent_total_test.go — #1735: the cards that read "the amount of
// mana spent to cast" a spell (CR 601.2h).
//
// Mockingbird reads it INSIDE the CR 614 entry window, through
// EntersAsCopyOfFromCast → game.EntryCastCountsForEffect; Mana Sculpt
// reads it off the spell it is countering, through the stack item's
// payment record. The engine half — every way the total cost is
// settled, the clone and the snapshot — is in
// game/mana_spent_total_test.go.

const manaSculptOracle = "35e2f82e-7ca3-4a92-9134-b7999eef5337"

// mvCreature puts a creature with the given printed cost on the
// battlefield under `controller` and returns it.
func mvCreature(g *game.Game, controller uuid.UUID, name, typeLine, cost string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: "oracle-" + name, Power: p, Toughness: tough,
		Owner: controller, Controller: controller,
	})
}

// mockingbirdCard is the printed card, as the deck importer builds it.
func mockingbirdCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Mockingbird", TypeLine: "Creature — Bird Bard",
		ManaCost: "{X}{U}", OracleID: mockingbirdOracle, Power: 1, Toughness: 1,
		Owner: owner, Controller: owner,
	}
}

// castMockingbird floats `pool`, casts Mockingbird for X = x and passes
// priority until the copy prompt opens. Returns the Mockingbird's ID
// and the prompt (nil if it never opened and the stack settled).
func castMockingbird(t *testing.T, g *game.Game, x int, pool string, strict bool) (uuid.UUID, *game.PendingChoice) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, pool)
	bird := mockingbirdCard(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bird) })
	if err := g.CastSpell(me.ID, bird.InstanceID, game.CastSpellParams{Strict: strict, XValue: x}); err != nil {
		t.Fatalf("cast Mockingbird: %v", err)
	}
	return bird.InstanceID, passUntilCopyPrompt(t, g)
}

// passUntilCopyPrompt passes priority until a copy-target prompt opens,
// or the stack settles with none.
func passUntilCopyPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		if pc := copyPrompt(g); pc != nil {
			return pc
		}
		if stackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return copyPrompt(g)
}

func sameIDSet(got []uuid.UUID, want ...uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[uuid.UUID]bool, len(got))
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			return false
		}
	}
	return true
}

// The ceiling is the amount SPENT, at the boundary and not over it:
// {X}{U} for X=2 spends three, so a three-drop and a two-drop are on
// the list and a four-drop is not. A token that copies nothing is mana
// value 0, so it is on the list too; a noncreature never is.
func TestMockingbirdOffersCreaturesUpToTheManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	under := mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	at := mvCreature(g, opp.ID, "Three Drop", "Creature — Elf Druid", "{2}{G}", 3, 3)
	over := mvCreature(g, opp.ID, "Four Drop", "Creature — Giant", "{3}{G}", 4, 4)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}",
		OracleID: "oracle-mind-stone", Owner: me.ID, Controller: me.ID,
	})

	_, pc := castMockingbird(t, g, 2, "UUU", true)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !sameIDSet(pc.CopyOptions, under, at, token) {
		t.Errorf("CopyOptions = %v, want the two-drop %s, the three-drop %s and the token %s (not the four-drop %s)",
			pc.CopyOptions, under, at, token, over)
	}
}

// What lands: the copied creature's values, "except it's a Bird in
// addition to its other types and it has flying" — the copied Elf Druid
// stays an Elf Druid, and a creature with no flying of its own gains it.
func TestMockingbirdCopyIsAFlyingBird(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	druid := mvCreature(g, opp.ID, "Three Drop", "Creature — Elf Druid", "{2}{G}", 3, 3)

	bird, pc := castMockingbird(t, g, 2, "UUU", true)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, druid); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	got := copyBattlefieldCard(t, g, bird)
	if got.Name != "Three Drop" {
		t.Fatalf("name = %q, want the copied creature's", got.Name)
	}
	if effectivePower(t, g, bird) != 3 || effectiveToughness(t, g, bird) != 3 {
		t.Errorf("P/T = %d/%d, want the copied 3/3", effectivePower(t, g, bird), effectiveToughness(t, g, bird))
	}
	subs := effectiveSubtypes(t, g, bird)
	for _, want := range []string{"Elf", "Druid", "Bird"} {
		if !containsString(subs, want) {
			t.Errorf("subtypes %v lack %q", subs, want)
		}
	}
	if !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the copy lacks flying")
	}
	if !got.IsCopy() {
		t.Error("Mockingbird does not report as a copy")
	}
	// The exception is part of the copiable values (CR 707.9b): the
	// original is untouched.
	if hasEffectiveKeyword(t, g, druid, "flying") || containsString(effectiveSubtypes(t, g, druid), "Bird") {
		t.Error("the copied creature itself became a flying Bird")
	}
}

// The exception is copiable (CR 707.9b, CR 707.9a): a Clone that
// copies the Mockingbird-as-a-Druid is a flying Bird Druid too.
func TestACloneOfAMockingbirdCopyIsAFlyingBirdToo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	druid := mvCreature(g, opp.ID, "Three Drop", "Creature — Elf Druid", "{2}{G}", 3, 3)
	bird, pc := castMockingbird(t, g, 2, "UUU", true)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, druid); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bird)
	if got := copyBattlefieldCard(t, g, clone); got.Name != "Three Drop" {
		t.Fatalf("the Clone is %q, want a copy of the copy", got.Name)
	}
	if !containsString(effectiveSubtypes(t, g, clone), "Bird") || !hasEffectiveKeyword(t, g, clone, "flying") {
		t.Errorf("the Clone did not copy the Bird and flying: subtypes %v", effectiveSubtypes(t, g, clone))
	}
}

// A pick over the ceiling is not a legal answer; it degrades to a
// decline rather than copying, as every copy prompt does.
func TestMockingbirdCannotBeAnsweredWithACreatureOverTheCeiling(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mvCreature(g, opp.ID, "One Drop", "Creature — Elf", "{G}", 1, 1)
	over := mvCreature(g, opp.ID, "Four Drop", "Creature — Giant", "{3}{G}", 4, 4)

	bird, pc := castMockingbird(t, g, 1, "UU", true)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, over); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := copyBattlefieldCard(t, g, bird); got.Name != "Mockingbird" {
		t.Errorf("name = %q — a four-drop was copied off two mana spent", got.Name)
	}
}

// Declining leaves the printed 1/1 Bird Bard flier.
func TestMockingbirdDeclinedIsAOneOneFlier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)

	bird, pc := castMockingbird(t, g, 1, "UU", true)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, uuid.Nil); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := copyBattlefieldCard(t, g, bird)
	if got.Name != "Mockingbird" || got.IsCopy() {
		t.Errorf("declined Mockingbird is %q (copy=%v)", got.Name, got.IsCopy())
	}
	if effectivePower(t, g, bird) != 1 || effectiveToughness(t, g, bird) != 1 {
		t.Errorf("P/T = %d/%d, want 1/1", effectivePower(t, g, bird), effectiveToughness(t, g, bird))
	}
	if !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("declined Mockingbird lost its printed flying")
	}
}

// Not cast, nothing spent: a reanimated Mockingbird can copy only a
// creature with mana value 0.
func TestMockingbirdReanimatedCopiesOnlyManaValueZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bird := mockingbirdCard(me.ID)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(bird)
		if err := g.ReturnFromGraveyardForEffect(bird.InstanceID, game.ZoneBattlefield); err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
	})
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt on the reanimated Mockingbird")
	}
	if !sameIDSet(pc.CopyOptions, token) {
		t.Errorf("CopyOptions = %v, want only the mana-value-0 token %s", pc.CopyOptions, token)
	}
}

// flickerMockingbird blinks the Mockingbird on the battlefield.
func flickerMockingbird(t *testing.T, g *game.Game, controller, bird uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: controller, SourceCardID: uuid.New()}
		if err := (Flicker{Target: bird}).Apply(ctxFor(g, item)); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
}

// Flickered, likewise: the Mockingbird that returns is a new object
// that was never cast, even though the one that left was.
func TestMockingbirdFlickeredCopiesOnlyManaValueZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bird, pc := castMockingbird(t, g, 3, "UUUU", true)
	if pc == nil {
		t.Fatal("no copy prompt on the cast")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, uuid.Nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)

	flickerMockingbird(t, g, me.ID, bird)
	pc = copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt on the flickered Mockingbird")
	}
	if !sameIDSet(pc.CopyOptions, token) {
		t.Errorf("CopyOptions = %v, want only the token %s — four mana were spent on the OLD object", pc.CopyOptions, token)
	}
}

// With no mana-value-0 creature out, a Mockingbird that was not cast
// is not asked at all and enters as itself.
func TestMockingbirdFlickeredWithNothingAtZeroReturnsAsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	bird := pushBattlefieldCardWithTimestamp(g, mockingbirdCard(me.ID))

	flickerMockingbird(t, g, me.ID, bird)
	if pc := copyPrompt(g); pc != nil {
		t.Fatalf("a flickered Mockingbird was offered %v off no mana spent", pc.CopyOptions)
	}
	if findBattlefieldByName(g, "Mockingbird") == uuid.Nil {
		t.Error("the flickered Mockingbird did not return as itself")
	}
}

// Declared caveat: a cast the engine did not charge (strict mana off)
// spent an unknown amount, which reads as zero — only a mana-value-0
// creature is offered, never "enough".
func TestMockingbirdCastWithStrictManaOffOffersOnlyManaValueZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mvCreature(g, me.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	_, pc := castMockingbird(t, g, 5, "", false)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !sameIDSet(pc.CopyOptions, token) {
		t.Errorf("CopyOptions = %v, want only the token %s — an unrecorded payment is not \"enough\"", pc.CopyOptions, token)
	}
}

// --- Mana Sculpt -------------------------------------------------------

// With a Wizard, the refund is the amount SPENT on the countered spell,
// not its mana value: a {2}{R/P} spell paid {2} and 2 life spent two
// mana, and refunds {C}{C} at the caster's next main phase.
func TestManaSculptRefundsTheManaSpentWithAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wizard", TypeLine: "Creature — Human Wizard",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatForTest(g, opp, "CC")
	victim := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{
			InstanceID: victim, Name: "Phyrexian Bolt", TypeLine: "Instant", ManaCost: "{2}{R/P}",
			Owner: opp.ID, Controller: opp.ID,
		})
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{
		Strict: true, PhyrexianLife: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}

	castCatalogSpell(t, g, "Mana Sculpt", "Instant", manaSculptOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || g.DelayedTriggers[0].At != game.StepPostcombatMain {
		t.Fatalf("want one refund at this turn's postcombat main, got %d", n)
	}
	batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C] — two mana spent, though its mana value is 3", got)
	}
}

// "If you control a Wizard": without one, a fully paid spell is still
// countered and nothing is refunded.
func TestManaSculptRefundsNothingWithoutAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	floatForTest(g, opp, "CCR")
	victim := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{
			InstanceID: victim, Name: "Big Bolt", TypeLine: "Instant", ManaCost: "{2}{R}",
			Owner: opp.ID, Controller: opp.ID,
		})
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	castCatalogSpell(t, g, "Mana Sculpt", "Instant", manaSculptOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("no Wizard, yet %d refund(s) scheduled", n)
	}
}

// Declared caveat: a countered spell whose caster had strict mana off
// spent an unknown amount, and Mana Sculpt adds nothing for it.
func TestManaSculptRefundsNothingForAWaivedPayment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wizard", TypeLine: "Creature — Human Wizard",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	victim := batch01OpponentCasts(t, g, opp, "Big Bolt", "", "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Mana Sculpt", "Instant", manaSculptOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("a waived payment scheduled %d refund(s)", n)
	}
}
