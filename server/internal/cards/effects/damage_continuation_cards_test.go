package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_continuation_cards_test.go is #807 seen from the catalog:
// Creeping Bloodsucker — "this creature deals 1 damage to each
// opponent. You gain life equal to the damage dealt this way" — on a
// board where the damage dealt this way is NOT 1 × the number of
// opponents.
//
// The engine-side suite is server/internal/game/damage_continuation_test.go.
// What this test adds is the card sentence, on a board built entirely
// out of real cards and real primitives: Angrath's Marauders doubles
// every damage event from a source its controller controls, and a
// CR 615.8 prevention shield sits on one opponent. Two DIFFERENT damage
// replacements on that one opponent is the mixed window that still
// prompts after #800 — identical replacements collapse without asking,
// and this pair is not identical.
//
// Before #807 the card read each opponent's life total back on the line
// after damaging them. The shielded opponent's damage was still waiting
// on the CR 616 ordering prompt, so their total had not moved, they
// counted as having taken nothing, and the gain was short by everything
// they took.

const b14AngrathsMaraudersOracle = "2d4976d4-649c-4d42-ac5a-ada4b46a480c"

// TestCreepingBloodsuckerGainsWhatWasActuallyDealtAcrossACR616Prompt is
// the issue's headline repro.
//
// The board: a Creeping Bloodsucker and an Angrath's Marauders, both
// mine, and a "prevent the next 1 damage" shield on the first opponent.
// At my upkeep the Bloodsucker deals 1 to each of three opponents.
//
//   - Opponents 2 and 3 see the doubler only — one replacement, no
//     prompt, 2 damage each.
//   - Opponent 1 sees the doubler AND the shield, so CR 616.1 asks them
//     to order the two. Ordered doubler-first, 1 becomes 2 and the
//     1-point shield eats 1 of it: 1 damage dealt.
//
// "The damage dealt this way" is therefore 5, and the gain is 5 — not
// 4 (the pre-#807 answer, with the paused opponent counted as zero) and
// not 6 (1 × 3 doubled, ignoring the shield).
func TestCreepingBloodsuckerGainsWhatWasActuallyDealtAcrossACR616Prompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shielded, o2, o3 := g.Seats[1], g.Seats[2], g.Seats[3]

	b12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", b14CreepingBloodsuckerOracle, 1, 2)
	b12Push(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", b14AngrathsMaraudersOracle, 4, 4)

	b12ToMyNextUpkeep(t, g)

	// The shield is turn-scoped (CR 514.2), so it is registered after
	// the turn it has to survive has begun — the cleanup step of the
	// previous turn sweeps anything registered earlier.
	g.WithWriteLock(func() {
		if err := (PreventNextDamage{
			Target: shielded.ID,
			Amount: 1,
			Label:  "prevent the next 1 damage",
		}).Apply(NewContext(g, nil)); err != nil {
			t.Fatalf("PreventNextDamage: %v", err)
		}
	})

	mine := me.Life
	lifeBefore := []int{shielded.Life, o2.Life, o3.Life}

	passPriorityAroundTable(t, g)

	// The batch is sequenced through the continuation, so nothing past
	// the paused leg has happened yet and nobody has gained anything.
	if me.Life != mine {
		t.Fatalf("the controller gained %d before the drain finished", me.Life-mine)
	}
	if shielded.Life != lifeBefore[0] {
		t.Fatalf("the shielded opponent's life moved to %d before the prompt was answered", shielded.Life)
	}

	answerOrderingPromptDoublerFirst(t, g)

	if got := lifeBefore[0] - shielded.Life; got != 1 {
		t.Errorf("shielded opponent took %d, want 1 — 1 doubled to 2, then 1 prevented", got)
	}
	if got := lifeBefore[1] - o2.Life; got != 2 {
		t.Errorf("opponent 2 took %d, want 2 — doubled, unshielded", got)
	}
	if got := lifeBefore[2] - o3.Life; got != 2 {
		t.Errorf("opponent 3 took %d, want 2 — doubled, unshielded", got)
	}
	if got := me.Life - mine; got != 5 {
		t.Errorf("controller gained %d, want 5 — the damage actually dealt (1+2+2), "+
			"not 4 with the paused opponent counted as zero", got)
	}
}

// TestCreepingBloodsuckerGainsNothingWhenEveryOpponentIsFogged is the
// zero end of the same contract: a replacement that eats the whole
// event is a terminal outcome too, so the batch is told and the gain is
// nothing — rather than the drain hanging on an opponent who took no
// damage.
func TestCreepingBloodsuckerGainsNothingWhenEveryOpponentIsFogged(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", b14CreepingBloodsuckerOracle, 1, 2)

	b12ToMyNextUpkeep(t, g)

	g.WithWriteLock(func() {
		for _, p := range g.Seats[1:] {
			if err := (PreventNextDamage{Target: p.ID, Label: "prevent that damage"}).Apply(NewContext(g, nil)); err != nil {
				t.Fatalf("PreventNextDamage: %v", err)
			}
		}
	})

	mine := me.Life
	theirs := lifeOfOpponents(g)

	passPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one shield per opponent is one replacement each — nothing to order, got %d prompts",
			len(g.PendingChoices))
	}
	for i, before := range theirs {
		if got := lifeOfOpponents(g)[i]; got != before {
			t.Errorf("opponent %d: %d → %d, want no change — the damage was prevented", i, before, got)
		}
	}
	if me.Life != mine {
		t.Errorf("controller life %d → %d, want no change — no damage was dealt this way", mine, me.Life)
	}
}

// answerOrderingPromptDoublerFirst answers the single queued CR 616
// ordering prompt with the damage doubler applied FIRST.
//
// The order is chosen rather than taken as offered because it is the
// order that makes the test say something: doubler-then-shield leaves 1
// damage for the continuation to report, while shield-then-doubler
// prevents the whole 1 and reports zero — which is also what the bug
// reported, so it could not tell the two apart.
func answerOrderingPromptDoublerFirst(t *testing.T, g *game.Game) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
			break
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt queued (have %d choices) — "+
			"a doubler and a prevention shield on one event are two different replacements",
			len(g.PendingChoices))
	}
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("prompt offers %d replacements, want 2", len(prompt.ReplacementEffectIDs))
	}
	ordered := make([]game.ReplacementEffectID, 0, 2)
	var shield []game.ReplacementEffectID
	for _, id := range prompt.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		if strings.Contains(strings.ToLower(label), "prevent") {
			shield = append(shield, id)
			continue
		}
		ordered = append(ordered, id)
	}
	if len(shield) != 1 || len(ordered) != 1 {
		t.Fatalf("expected exactly one prevention shield and one doubler in the prompt, got %d/%d",
			len(shield), len(ordered))
	}
	ordered = append(ordered, shield...)
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, ordered); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
