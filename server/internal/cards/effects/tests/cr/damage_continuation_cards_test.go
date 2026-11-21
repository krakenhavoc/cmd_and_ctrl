package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shielded, o2, o3 := g.Seats[1], g.Seats[2], g.Seats[3]

	effectstest.B12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", effectstest.B14CreepingBloodsuckerOracle, 1, 2)
	effectstest.B12Push(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", effectstest.B14AngrathsMaraudersOracle, 4, 4)

	effectstest.B12ToMyNextUpkeep(t, g)

	// The shield is turn-scoped (CR 514.2), so it is registered after
	// the turn it has to survive has begun — the cleanup step of the
	// previous turn sweeps anything registered earlier.
	g.WithWriteLock(func() {
		if err := (effects.PreventNextDamage{
			Target: shielded.ID,
			Amount: 1,
			Label:  "prevent the next 1 damage",
		}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Fatalf("PreventNextDamage: %v", err)
		}
	})

	mine := me.Life
	lifeBefore := []int{shielded.Life, o2.Life, o3.Life}

	effectstest.PassPriorityAroundTable(t, g)

	// The batch is sequenced through the continuation, so nothing past
	// the paused leg has happened yet and nobody has gained anything.
	if me.Life != mine {
		t.Fatalf("the controller gained %d before the drain finished", me.Life-mine)
	}
	if shielded.Life != lifeBefore[0] {
		t.Fatalf("the shielded opponent's life moved to %d before the prompt was answered", shielded.Life)
	}

	effectstest.AnswerOrderingPromptDoublerFirst(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", effectstest.B14CreepingBloodsuckerOracle, 1, 2)

	effectstest.B12ToMyNextUpkeep(t, g)

	g.WithWriteLock(func() {
		for _, p := range g.Seats[1:] {
			// A shield larger than the drain: the whole event is
			// prevented. (The uncharged "prevent that damage" form is
			// not built — ADR 0041 P8 adds a kind with its first card.)
			if err := (effects.PreventNextDamage{Target: p.ID, Amount: 100, Label: "prevent the next 100 damage"}).Apply(effects.NewContext(g, nil)); err != nil {
				t.Fatalf("PreventNextDamage: %v", err)
			}
		}
	})

	mine := me.Life
	theirs := effectstest.LifeOfOpponents(g)

	effectstest.PassPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one shield per opponent is one replacement each — nothing to order, got %d prompts",
			len(g.PendingChoices))
	}
	for i, before := range theirs {
		if got := effectstest.LifeOfOpponents(g)[i]; got != before {
			t.Errorf("opponent %d: %d → %d, want no change — the damage was prevented", i, before, got)
		}
	}
	if me.Life != mine {
		t.Errorf("controller life %d → %d, want no change — no damage was dealt this way", mine, me.Life)
	}
}
