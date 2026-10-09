package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// copy_retarget_2622_test.go — #2622. A catalog-soak bot table stalled
// on "Choose new target creature you control for the copy (or re-pick
// the same)": the CR 707.10c prompt a copy of Bite Down opens. Every
// case below copies a real catalog spell with "you may choose new
// targets for the copy" and asserts the two things a seat needs while
// the prompt is up:
//
//	(a) the legal enumerator offers the chooser at least one answer;
//	(b) every answer it offers is accepted when dispatched (on a clone).
//
// A copy prompt for a spell of several target clauses asks one clause
// at a time, so a walk that opens a further prompt is followed down
// every branch.

// copyPromptFor is the open CR 707.10c prompt addressed to chooser.
func copyPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser && strings.Contains(c.Reason, "for the copy") {
			return c
		}
	}
	return nil
}

// stackCopies counts the spell copies on the stack.
func stackCopies(g *game.Game) int {
	n := 0
	for _, it := range g.StackMeta {
		if it != nil && it.IsCopy {
			n++
		}
	}
	return n
}

// copyAnswers is every answer the enumerator offers for the open copy
// prompt.
func copyAnswers(t *testing.T, g *game.Game, chooser uuid.UUID, choice uuid.UUID) []legal.Move {
	t.Helper()
	var out []legal.Move
	for _, m := range legal.EnumerateForWithOptions(g, chooser, legal.Options{MaxExpansionPerSource: 1000}) {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.ChoiceID == choice.String() {
			out = append(out, m)
		}
	}
	return out
}

// walkCopyPrompt asserts (a) and (b) for the open copy prompt, then
// follows each accepted answer into the next clause's prompt until the
// copy is on the stack. It returns every final target list reached.
func walkCopyPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, wantCopies int, path string) [][]game.TargetRef {
	t.Helper()
	c := copyPromptFor(g, chooser)
	if c == nil {
		if got := stackCopies(g); got != wantCopies {
			t.Errorf("%s: no prompt open and %d copies on the stack, want %d", path, got, wantCopies)
		}
		var finals [][]game.TargetRef
		for _, it := range g.StackMeta {
			if it != nil && it.IsCopy {
				finals = append(finals, it.Targets)
			}
		}
		return finals
	}
	moves := copyAnswers(t, g, chooser, c.ID)
	if len(moves) == 0 {
		t.Fatalf("%s: %q is open and the enumerator offers %s no answer (players %v, cards %v, min %d, max %d)",
			path, c.Reason, chooser, c.PickTargetPlayers, c.PickTargetCards, c.PickTargetMin, c.PickTargetMax)
	}
	var finals [][]game.TargetRef
	for _, m := range moves {
		cl := g.Clone()
		if err := actions.Dispatch(cl, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: chooser, Params: m.Params}); err != nil {
			t.Errorf("%s: offered answer %q refused: %v", path, m.Label, err)
			continue
		}
		if still := copyPromptFor(cl, chooser); still != nil && still.ID == c.ID {
			t.Errorf("%s: answer %q accepted but the prompt is still open", path, m.Label)
			continue
		}
		finals = append(finals, walkCopyPrompt(t, cl, chooser, wantCopies, path+" > "+m.Label)...)
	}
	// The live prompt must be untouched by the answers given on clones.
	if again := copyPromptFor(g, chooser); again == nil || again.ID != c.ID {
		t.Fatalf("%s: answering on a clone changed the live game's prompt", path)
	}
	return finals
}

// hasFinal reports whether one of the final target lists names exactly
// ids, in order.
func hasFinal(finals [][]game.TargetRef, ids ...uuid.UUID) bool {
	for _, f := range finals {
		if len(f) != len(ids) {
			continue
		}
		ok := true
		for i := range f {
			if f[i].ID != ids[i] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func copySpellNow(t *testing.T, g *game.Game, spell, copier uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, copier, true, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
}

func castCopyTarget(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func cardTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}

func TestCopyRetargetPromptAlwaysHasAnAcceptedAnswer(t *testing.T) {
	t.Run("one target, still legal", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := pr7Creature(g, opp.ID, "Bear", 2)
		pr7Creature(g, me.ID, "Mine", 2)
		bolt := castCopyTarget(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle,
			game.CastSpellParams{Targets: cardTargets(bear)})
		copySpellNow(t, g, bolt, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "bolt")
		if !hasFinal(finals, bear) {
			t.Error("re-picking the same target was not offered")
		}
	})

	// CR 707.10c: "The player may leave any number of the targets
	// unchanged, even if those targets would be illegal."
	t.Run("one target, gone", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := pr7Creature(g, opp.ID, "Bear", 2)
		bolt := castCopyTarget(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle,
			game.CastSpellParams{Targets: cardTargets(bear)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
		copySpellNow(t, g, bolt, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "bolt")
		if !hasFinal(finals, bear) {
			t.Error("leaving the now-illegal target unchanged was not offered (CR 707.10c)")
		}
	})

	t.Run("the chooser is not the active seat", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := pr7Creature(g, opp.ID, "Bear", 2)
		pr7Creature(g, me.ID, "Mine", 2)
		bolt := castCopyTarget(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle,
			game.CastSpellParams{Targets: cardTargets(bear)})
		copySpellNow(t, g, bolt, opp.ID)
		if copyPromptFor(g, opp.ID) == nil {
			t.Fatal("the copy's controller was not asked")
		}
		walkCopyPrompt(t, g, opp.ID, 1, "bolt")
	})

	// The #2622 shape: two clauses, the first "target creature you
	// control".
	t.Run("Bite Down, two clauses", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := pr7Creature(g, me.ID, "Mine", 3)
		mine2 := pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := pr7Creature(g, opp.ID, "Theirs", 2)
		theirs2 := pr7Creature(g, opp.ID, "Theirs 2", 2)
		bite := castCopyTarget(t, g, me, "Bite Down", "Instant", b26BiteDownOracle,
			game.CastSpellParams{Targets: cardTargets(mine, theirs)})
		copySpellNow(t, g, bite, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "bite")
		for _, want := range [][]uuid.UUID{{mine, theirs}, {mine2, theirs2}, {mine, theirs2}, {mine2, theirs}} {
			if !hasFinal(finals, want...) {
				t.Errorf("the copy could not end up targeting %v", want)
			}
		}
	})

	// The copy's controller is not the original's, the original biter is
	// hexproof, and "you control" now means the copier.
	t.Run("Bite Down copied by an opponent, hexproof original", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := pr7bCreature(g, me.ID, "Hexproof Mine", 3, "{G}", []string{"G"}, "hexproof")
		mine2 := pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := pr7Creature(g, opp.ID, "Theirs", 2)
		bite := castCopyTarget(t, g, me, "Bite Down", "Instant", b26BiteDownOracle,
			game.CastSpellParams{Targets: cardTargets(mine, theirs)})
		copySpellNow(t, g, bite, opp.ID)
		finals := walkCopyPrompt(t, g, opp.ID, 1, "bite")
		if !hasFinal(finals, mine, theirs) {
			t.Error("leaving both targets unchanged was not offered (CR 707.10c)")
		}
		if !hasFinal(finals, theirs, mine2) {
			t.Error("the copier could not aim its copy from its own creature at the caster's")
		}
	})

	t.Run("Bite Down, no new legal target anywhere", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := pr7Creature(g, me.ID, "Mine", 3)
		theirs := pr7Creature(g, opp.ID, "Theirs", 2)
		bite := castCopyTarget(t, g, me, "Bite Down", "Instant", b26BiteDownOracle,
			game.CastSpellParams{Targets: cardTargets(mine, theirs)})
		copySpellNow(t, g, bite, me.ID)
		if c := copyPromptFor(g, me.ID); c != nil {
			t.Errorf("prompted with nothing to change to: %q", c.Reason)
		}
		finals := walkCopyPrompt(t, g, me.ID, 1, "bite")
		if !hasFinal(finals, mine, theirs) {
			t.Errorf("the copy did not keep the original targets: %v", finals)
		}
	})

	t.Run("Dromoka's Command, targets in two modes", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := pr7Creature(g, me.ID, "Mine", 2)
		pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := pr7Creature(g, opp.ID, "Theirs", 1)
		pr7Creature(g, opp.ID, "Theirs 2", 1)
		cmd := castCopyTarget(t, g, me, "Dromoka's Command", "Instant", pr7bDromoka, game.CastSpellParams{
			Modes: []int{2, 3},
			Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: mine, Mode: 0},
				{Kind: game.TargetCard, ID: mine, Mode: 1, Slot: 0},
				{Kind: game.TargetCard, ID: theirs, Mode: 1, Slot: 1},
			},
		})
		copySpellNow(t, g, cmd, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "command")
		if !hasFinal(finals, mine, mine, theirs) {
			t.Error("leaving every target unchanged was not offered")
		}
	})

	t.Run("Switcheroo, one of two targets gone", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := pr7Creature(g, me.ID, "A", 2)
		b := pr7Creature(g, opp.ID, "B", 2)
		c := pr7Creature(g, opp.ID, "C", 2)
		sw := castCopyTarget(t, g, me, "Switcheroo", "Sorcery", switcherooOracle,
			game.CastSpellParams{Targets: cardTargets(a, b)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
		copySpellNow(t, g, sw, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "switcheroo")
		if !hasFinal(finals, a, b) {
			t.Error("leaving both targets unchanged was not offered (CR 707.10c)")
		}
		if !hasFinal(finals, c, b) {
			t.Error("replacing the gone target alone was not offered")
		}
	})

	// Only the surviving original is still legal: the copy keeps its
	// targets without a prompt nobody could answer.
	t.Run("Switcheroo, one target gone and no replacement", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := pr7Creature(g, me.ID, "A", 2)
		b := pr7Creature(g, opp.ID, "B", 2)
		sw := castCopyTarget(t, g, me, "Switcheroo", "Sorcery", switcherooOracle,
			game.CastSpellParams{Targets: cardTargets(a, b)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
		copySpellNow(t, g, sw, me.ID)
		finals := walkCopyPrompt(t, g, me.ID, 1, "switcheroo")
		if !hasFinal(finals, a, b) {
			t.Errorf("the copy did not keep its targets: %v", finals)
		}
	})
}

// CR 115.7f: choosing new targets for a copy cannot change the original
// division. The copy's answer is a set: the target it keeps holds its
// share, and the new one takes the share of the target it replaced.
func TestCopyRetargetKeepsTheDivision(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	id := b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3}})
	copySpellNow(t, g, id, me.ID)
	p := copyPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("no copy prompt")
	}
	if p.PickTargetMin != 2 || p.PickTargetMax != 2 {
		t.Errorf("the copy is asked for %d..%d targets, want exactly the original's 2", p.PickTargetMin, p.PickTargetMax)
	}
	// Listed with the kept target first: it stays in its own slot.
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(b, c)); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := []int{e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked, e2Card(t, g, c).DamageMarked}
	if got[0] != 1 || got[1] != 6 || got[2] != 1 {
		t.Errorf("damage on A/B/C = %v, want [1 6 1]: the original deals 1/3 to A/B and the copy 1/3 to C/B", got)
	}
}
