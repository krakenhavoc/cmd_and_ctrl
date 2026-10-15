package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_replacement_identity_test.go — #3482: a paused replacement
// event names its effects by the object behind them (CR 400.7), not by
// where that object sits on the battlefield.

// oneMoreLifeGainedBy is a second, different gain replacement, so a
// gain beside a Rhox Faithmender asks a CR 616 question.
func oneMoreLifeGainedBy(player uuid.UUID) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventLife && ev.LifeDelta > 0 && ev.LifePlayer == player
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.LifeDelta++
			return nil
		},
		Label: "you gain one more life",
	}
}

// pausedGainBoard puts a vanilla creature and Rhox Faithmender under
// seat 0, in that battlefield order, plus a second gain replacement,
// and starts a gain of 2 that pauses on the CR 616 ordering prompt.
func pausedGainBoard(t *testing.T) (g *game.Game, me *game.Player, vanilla, faithmender uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	vanilla = pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	faithmender = b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(oneMoreLifeGainedBy(me.ID))
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2); err != nil {
			t.Fatalf("gain: %v", err)
		}
	})
	return g, me, vanilla, faithmender
}

func moveToGraveyard(t *testing.T, g *game.Game, owner, card uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, card,
	); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
}

// The issue's reproduction. The creature before the Faithmender leaves
// while the prompt is open; the doubling has applied once (CR 614.5)
// and must not be gathered again under the index it has moved to.
// Before the fix the gain was 10: 2, doubled, one more, doubled again.
func TestPausedReplacementKeepsItsMarksWhenAnEarlierPermanentLeaves(t *testing.T) {
	g, me, vanilla, _ := pausedGainBoard(t)
	life := me.Life
	moveToGraveyard(t, g, me.ID, vanilla)
	answerTheOrderingPrompt(t, g)
	if gained := me.Life - life; gained != 5 && gained != 6 {
		t.Errorf("gained %d, want 2 doubled and one more, in either order (5 or 6)", gained)
	}
}

// The other order of the same board, so both effects are checked as
// the one applied first.
func TestPausedReplacementKeepsItsMarksInTheOtherOrder(t *testing.T) {
	g, me, vanilla, _ := pausedGainBoard(t)
	life := me.Life
	moveToGraveyard(t, g, me.ID, vanilla)
	prompt := orderingPrompt(t, g)
	ids := prompt.ReplacementEffectIDs
	reversed := []game.ReplacementEffectID{ids[1], ids[0]}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, reversed); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if gained := me.Life - life; gained != 5 && gained != 6 {
		t.Errorf("gained %d, want 5 or 6", gained)
	}
}

// A listed effect whose source has left is dropped (a static
// ability's effect exists only while its source is on the battlefield).
// It is not mapped onto the permanent that slid into its slot: here the
// bear, which a stale battlefield pointer would read as a Faithmender
// under the same controller.
func TestPausedReplacementDropsAnEffectWhoseSourceLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	faithmender := b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(oneMoreLifeGainedBy(me.ID))
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2); err != nil {
			t.Fatalf("gain: %v", err)
		}
	})
	moveToGraveyard(t, g, me.ID, faithmender)
	answerTheOrderingPrompt(t, g)
	if gained := me.Life - life; gained != 3 {
		t.Errorf("gained %d, want 3: the departed Faithmender doubles nothing", gained)
	}
}

// A Faithmender that leaves and comes back is a new object (CR 400.7).
// The old object's listed effect is dropped; the new one is gathered
// afresh once the ordered effects have run, as any effect that becomes
// applicable mid-event is (CR 616.1f), so the gain is still doubled
// exactly once.
func TestPausedReplacementTreatsAReturnedSourceAsANewObject(t *testing.T) {
	g, me, _, faithmender := pausedGainBoard(t)
	life := me.Life
	moveToGraveyard(t, g, me.ID, faithmender)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, faithmender,
	); err != nil {
		t.Fatalf("return the Faithmender: %v", err)
	}
	answerTheOrderingPrompt(t, g)
	if gained := me.Life - life; gained != 6 {
		t.Errorf("gained %d, want 6: one more, then the returned Faithmender doubles once", gained)
	}
}

// Undo into the open prompt replays the same way: the restored game
// answers with the effects the prompt listed and the marks it held.
func TestPausedReplacementReplaysAfterUndo(t *testing.T) {
	g, me, vanilla, _ := pausedGainBoard(t)
	life := me.Life
	moveToGraveyard(t, g, me.ID, vanilla)
	snap := g.Clone()
	answerTheOrderingPrompt(t, g)
	first := me.Life - life
	g.RestoreFrom(snap)
	me = g.Seats[0]
	if me.Life != life {
		t.Fatalf("undo left life at %d, want %d", me.Life, life)
	}
	answerTheOrderingPrompt(t, g)
	if again := me.Life - life; again != first || (again != 5 && again != 6) {
		t.Errorf("replayed gain %d, first answer %d; want the same 5 or 6", again, first)
	}
}

func orderingPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			if len(c.ReplacementEffectIDs) != 2 {
				t.Fatalf("prompt lists %d effects, want 2", len(c.ReplacementEffectIDs))
			}
			return c
		}
	}
	t.Fatalf("no CR 616 ordering prompt queued (have %d choices)", len(g.PendingChoices))
	return nil
}
