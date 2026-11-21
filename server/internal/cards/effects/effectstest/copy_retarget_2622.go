package effectstest

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
func CopyPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser && strings.Contains(c.Reason, "for the copy") {
			return c
		}
	}
	return nil
}

// stackCopies counts the spell copies on the stack.
func StackCopies(g *game.Game) int {
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
func WalkCopyPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, wantCopies int, path string) [][]game.TargetRef {
	t.Helper()
	c := CopyPromptFor(g, chooser)
	if c == nil {
		if got := StackCopies(g); got != wantCopies {
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
		if still := CopyPromptFor(cl, chooser); still != nil && still.ID == c.ID {
			t.Errorf("%s: answer %q accepted but the prompt is still open", path, m.Label)
			continue
		}
		finals = append(finals, WalkCopyPrompt(t, cl, chooser, wantCopies, path+" > "+m.Label)...)
	}
	// The live prompt must be untouched by the answers given on clones.
	if again := CopyPromptFor(g, chooser); again == nil || again.ID != c.ID {
		t.Fatalf("%s: answering on a clone changed the live game's prompt", path)
	}
	return finals
}

// hasFinal reports whether one of the final target lists names exactly
// ids, in order.
func HasFinal(finals [][]game.TargetRef, ids ...uuid.UUID) bool {
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

func CopySpellNow(t *testing.T, g *game.Game, spell, copier uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, copier, true, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
}

func CastCopyTarget(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func CardTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}
