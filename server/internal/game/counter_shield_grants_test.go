package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_shield_grants_test.go — ADR 0106 §4 (#1806), delivery PR 3:
// the counter gate's sources 3 and 5 (a mark on one spell, a player's
// "this turn" grant) and the one-use promise spent at CR 601.2i. The
// printed cards (Veil of Summer, Insist, Vexing Shusher …) are pinned
// end to end in cards/effects/counter_shield_grants_test.go.

func grantShield(g *Game, p *Player, gr CounterShieldGrant, label string) {
	g.WithWriteLock(func() { g.GrantCounterShieldForEffect(p.ID, gr, label, uuid.New()) })
}

// endTheTurn runs the CR 514.2 cleanup sweep: "this turn" effects end.
func endTheTurn(g *Game) {
	g.WithWriteLock(func() { g.sweepPlayerStaticsLocked(true) })
}

// castFromHand puts a card in `p`'s hand and casts it, the real cast
// path (CR 601.2i included).
func castFromHand(t *testing.T, g *Game, p *Player, name, typeLine string) uuid.UUID {
	t.Helper()
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return c.InstanceID
}

// A "this turn" grant (Veil of Summer) covers its player's spells,
// including one cast after it began (CR 611.2c), and nobody else's,
// and ends at cleanup (CR 514.2).
func TestATurnGrantCoversYourSpellsUntilCleanup(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := pushSpellFor(t, g, "Instant", me, me)
	grantShield(g, me, CounterShieldGrant{Text: "Spells you control can't be countered this turn."}, "Veil of Summer")
	after := pushSpellFor(t, g, "Sorcery", me, me)
	theirs := pushSpellFor(t, g, "Instant", opp, opp)

	if !cantBeCountered(g, before) || !cantBeCountered(g, after) {
		t.Error("a spell you control is covered, whether cast before or after the grant")
	}
	if cantBeCountered(g, theirs) {
		t.Error("an opponent's spell is not")
	}
	endTheTurn(g)
	if cantBeCountered(g, after) {
		t.Error("the grant ends at cleanup")
	}
	if len(me.Statics) != 0 {
		t.Errorf("the sweep leaves %d statics", len(me.Statics))
	}
}

// "You control" follows the spell's current controller (ADR 0104); "you
// cast" follows its caster and never covers a copy (CR 707.10).
func TestATurnGrantJudgesControlAndCastAsPrinted(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantShield(g, me, CounterShieldGrant{Whose: CounterShieldYouControl}, "Veil of Summer")
	grantShield(g, opp, CounterShieldGrant{Whose: CounterShieldYouCast, Filter: PermissionFilter{CreatureOnly: true}}, "Domri, Anarch of Bolas")

	stolen := pushSpellFor(t, g, "Creature — Bear", opp, me)
	if !cantBeCountered(g, stolen) {
		t.Error("a creature spell the opponent cast is theirs to shield by cast, and mine by control")
	}
	g.WithWriteLock(func() { g.Seats[0].Statics = nil })
	if !cantBeCountered(g, stolen) {
		t.Error("\"creature spells you cast\" still covers a spell its caster no longer controls")
	}
	theirSorcery := pushSpellFor(t, g, "Sorcery", opp, opp)
	if cantBeCountered(g, theirSorcery) {
		t.Error("Domri's grant names creature spells only")
	}
	copied := pushSpellFor(t, g, "Creature — Bear", opp, opp)
	g.StackMeta[copied].IsCopy = true
	if cantBeCountered(g, copied) {
		t.Error("a copy is not cast, so \"spells you cast\" does not cover it")
	}
}

// "OTHER spells you control" (Determined) excepts the one object.
func TestATurnGrantExceptsTheSpellThatMadeIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	own := pushSpellFor(t, g, "Instant", me, me)
	other := pushSpellFor(t, g, "Instant", me, me)
	grantShield(g, me, CounterShieldGrant{Except: ObjectRef{ID: own}}, "Determined")
	if cantBeCountered(g, own) {
		t.Error("the excepted spell is not covered")
	}
	if !cantBeCountered(g, other) {
		t.Error("every other spell you control is")
	}
	// The same card as a NEW object (CR 400.7) is a different spell.
	g.Stack.Cards[len(g.Stack.Cards)-2].ObjectEpoch++
	if !cantBeCountered(g, own) {
		t.Error("the except names an object, not a card")
	}
}

// The one-use promise (Insist) is spent at CR 601.2i by the first
// matching spell its player casts, becomes a mark on it, and is never
// read at the gate before that.
func TestAPromiseIsSpentByTheFirstMatchingSpellCast(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	promise := CounterShieldGrant{NextOnly: true, Filter: PermissionFilter{CreatureOnly: true},
		Text: "The next creature spell you cast this turn can't be countered."}
	grantShield(g, me, promise, "Insist")

	onStack := pushSpellFor(t, g, "Creature — Bear", me, me)
	if cantBeCountered(g, onStack) {
		t.Fatal("a promise is not read at the gate: a spell that was not cast after it is not covered")
	}
	g.WithWriteLock(func() {
		g.Stack.Cards = nil
		g.StackMeta = nil
	})

	sorcery := castFromHand(t, g, me, "Divination", "Sorcery")
	if cantBeCountered(g, sorcery) {
		t.Error("a sorcery does not match a creature promise")
	}
	if got := g.CounterShieldGrantsForEffect(me); len(got) != 1 || !got[0].NextOnly || got[0].SourceName != "Insist" {
		t.Fatalf("a spell that doesn't match leaves the promise: %+v", got)
	}
	g.WithWriteLock(func() {
		g.Stack.Cards = nil
		g.StackMeta = nil
	})

	// An opponent's promise is theirs: my creature spell neither spends
	// nor uses it.
	grantShield(g, opp, promise, "Their Insist")
	bear := castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	if got := g.CounterShieldGrantsForEffect(opp); len(got) != 1 {
		t.Errorf("only the promise's own player spends it: %+v", got)
	}
	if !cantBeCountered(g, bear) {
		t.Fatal("the first matching spell spends the promise and can't be countered")
	}
	marks := g.StackMeta[bear].CantBeCountered
	if len(marks) != 1 || marks[0].SourceName != "Insist" || marks[0].Label != promise.Text {
		t.Errorf("the mark: %+v", marks)
	}
	if len(me.Statics) != 0 {
		t.Errorf("the promise is gone from the player: %+v", me.Statics)
	}
	g.WithWriteLock(func() {
		g.Stack.Cards = nil
		g.StackMeta = nil
	})
	second := castFromHand(t, g, me, "Second Bear", "Creature — Bear")
	if cantBeCountered(g, second) {
		t.Error("a promise is spent once")
	}
}

// Two promises that match one spell are both spent on it; a promise
// nobody spends ends at cleanup.
func TestTwoPromisesAreBothSpentAndAnUnspentOneEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantShield(g, me, CounterShieldGrant{NextOnly: true}, "Mistrise Village")
	grantShield(g, me, CounterShieldGrant{NextOnly: true, Filter: PermissionFilter{InstantOrSorceryOnly: true}}, "Overmaster")
	grantShield(g, me, CounterShieldGrant{NextOnly: true, Filter: PermissionFilter{CreatureOnly: true}}, "Insist")

	spell := castFromHand(t, g, me, "Divination", "Sorcery")
	if got := len(g.StackMeta[spell].CantBeCountered); got != 2 {
		t.Errorf("both matching promises are spent on the sorcery: %d marks", got)
	}
	left := g.CounterShieldGrantsForEffect(me)
	if len(left) != 1 || left[0].SourceName != "Insist" {
		t.Fatalf("the creature promise is left: %+v", left)
	}
	endTheTurn(g)
	if len(me.Statics) != 0 {
		t.Error("an unspent promise ends at cleanup")
	}
}

// ADR 0106 §4 decision 4, written before the copy code was touched: a
// copy does not inherit a mark (CR 707.2 — an effect is not a copiable
// value), and copying spends no promise (CR 707.10 — a copy is not
// cast). The original keeps its mark.
func TestACopyNeitherInheritsAMarkNorSpendsAPromise(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Instant", me, me)
	g.WithWriteLock(func() {
		if !g.MarkSpellCantBeCounteredForEffect(spell, CounterShieldMark{SourceName: "Vexing Shusher"}) {
			t.Fatal("mark the spell")
		}
	})
	grantShield(g, me, CounterShieldGrant{NextOnly: true}, "Mistrise Village")
	var copyID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
		copyID = g.Stack.Cards[len(g.Stack.Cards)-1].InstanceID
	})
	if copyID == spell || !g.StackMeta[copyID].IsCopy {
		t.Fatal("no copy on top of the stack")
	}
	if len(g.StackMeta[copyID].CantBeCountered) != 0 || cantBeCountered(g, copyID) {
		t.Error("the copy does not inherit the mark")
	}
	if !cantBeCountered(g, spell) {
		t.Error("the original keeps it")
	}
	if got := g.CounterShieldGrantsForEffect(me); len(got) != 1 {
		t.Errorf("a copy is not cast, so the promise is unspent: %+v", got)
	}
}

// "Target spell can't be countered" marks the object: it survives a
// change of control (ADR 0104), refuses an ability, and goes with the
// item when it leaves the stack (CR 400.7).
func TestAMarkStaysWithTheSpellObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Sorcery", me, me)
	ability := uuid.New()
	g.WithWriteLock(func() {
		g.StackMeta[ability] = &StackItem{ID: ability, Kind: StackItemActivated, Controller: me.ID}
		if g.MarkSpellCantBeCounteredForEffect(ability, CounterShieldMark{}) {
			t.Error("an ability can't be marked: the statement is about a spell")
		}
		if !g.MarkSpellCantBeCounteredForEffect(spell, CounterShieldMark{SourceName: "Vexing Shusher"}) {
			t.Fatal("mark the spell")
		}
	})
	g.StackMeta[spell].BaseController = me.ID
	g.StackMeta[spell].Controller = opp.ID
	if !cantBeCountered(g, spell) {
		t.Error("a stolen spell keeps its mark")
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if !g.Stack.Contains(spell) {
		t.Fatal("the marked spell was countered")
	}
}

// An undo snapshot taken before the mark keeps the item as it was: the
// writer replaces the slice rather than appending into a shared array.
func TestAMarkDoesNotReachAnEarlierClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Instant", me, me)
	g.WithWriteLock(func() {
		g.MarkSpellCantBeCounteredForEffect(spell, CounterShieldMark{SourceName: "first"})
	})
	var before *Game
	before = g.Clone()
	g.WithWriteLock(func() {
		g.MarkSpellCantBeCounteredForEffect(spell, CounterShieldMark{SourceName: "second"})
	})
	if got := len(before.StackMeta[spell].CantBeCountered); got != 1 {
		t.Errorf("the clone sees %d marks, want 1", got)
	}
	if got := len(g.StackMeta[spell].CantBeCountered); got != 2 {
		t.Errorf("the live item has %d marks, want 2", got)
	}
}

// Both new fields are plain data and survive a restore point.
func TestGrantsAndMarksSurviveARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Instant", me, me)
	g.WithWriteLock(func() {
		g.MarkSpellCantBeCounteredForEffect(spell, CounterShieldMark{Source: uuid.New(), SourceName: "Vexing Shusher", Label: "Target spell can't be countered."})
	})
	grantShield(g, me, CounterShieldGrant{NextOnly: true, Filter: PermissionFilter{CreatureOnly: true}, Text: "next creature"}, "Insist")
	grantShield(g, me, CounterShieldGrant{Text: "Spells you control can't be countered this turn."}, "Veil of Summer")

	_, restored := roundTrip(t, g)
	if got := restored.StackMeta[spell].CantBeCountered; len(got) != 1 || got[0] != g.StackMeta[spell].CantBeCountered[0] {
		t.Errorf("the mark: %+v", got)
	}
	rme := restored.playerByIDLocked(me.ID)
	if len(rme.Statics) != 2 || rme.Statics[0].CantBeCountered != me.Statics[0].CantBeCountered ||
		rme.Statics[1].CantBeCountered != me.Statics[1].CantBeCountered {
		t.Errorf("the grants: %+v", rme.Statics)
	}
	other := pushSpellFor(t, restored, "Sorcery", rme, rme)
	if !cantBeCountered(restored, other) {
		t.Error("the restored turn grant still covers a new spell")
	}
}
