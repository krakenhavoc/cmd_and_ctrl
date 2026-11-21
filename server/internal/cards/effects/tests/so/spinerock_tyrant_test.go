package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spinerock_tyrant_test.go — #4693: "those spells gain wither". The
// original and the Tyrant's copy have wither while they are on the stack,
// so their damage to a creature is -1/-1 counters (CR 702.80a, 702.80c,
// 120.3d) and to a player is life loss (CR 120.3a). A copy some other
// effect makes does not have it (CR 707.2).

const spinerockTyrantOracle = "e35d4c62-5211-4785-b683-a83719195d87"

// spinerockSetup puts the Tyrant under the active player and two 12/12s
// under the next player, and returns the caster, the opponent and the
// two creatures.
func spinerockSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Spinerock Tyrant", "Creature — Dragon", spinerockTyrantOracle, 6, 6)
	a := effectstest.B12Creature(g, opp.ID, "Big A", "Creature — Test", 12, 12)
	b := effectstest.B12Creature(g, opp.ID, "Big B", "Creature — Test", 12, 12)
	return g, me, opp, a, b
}

// spellWither reports, for each spell on the stack, whether it has
// wither, keyed by its ID, and the ID of the one copy (uuid.Nil when none).
func spellWither(t *testing.T, g *game.Game) (map[uuid.UUID]bool, []uuid.UUID) {
	t.Helper()
	out := map[uuid.UUID]bool{}
	var copies []uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for i := range g.Stack.Cards {
			c := &g.Stack.Cards[i]
			item := g.StackMeta[c.InstanceID]
			if item == nil || item.Kind != game.StackItemSpell {
				continue
			}
			out[c.InstanceID] = game.HasKeyword(c, game.KeywordWither)
			if item.IsCopy {
				copies = append(copies, c.InstanceID)
			}
		}
	})
	return out, copies
}

// answerCopyTarget answers the open "choose new targets for the copy"
// prompt with `to`.
func answerCopyTarget(t *testing.T, g *game.Game, chooser uuid.UUID, to game.TargetRef) {
	t.Helper()
	p := effectstest.CopyPromptFor(g, chooser)
	if p == nil {
		t.Fatal("no new-targets prompt for the copy is open")
	}
	if err := g.ResolvePickTarget(p.ID, chooser, to); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func creatureRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// The issue's case: a Lightning Bolt at one creature, copied at another.
// Both spells have wither on the stack, and each creature gets three
// -1/-1 counters and no marked damage.
func TestSpinerockTyrantCopiesABoltAndBothSpellsGainWither(t *testing.T) {
	g, me, _, a, b := spinerockSetup(t)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{creatureRef(a)})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	answerCopyTarget(t, g, me.ID, creatureRef(b))

	wither, copies := spellWither(t, g)
	if len(copies) != 1 {
		t.Fatalf("%d copies on the stack, want 1", len(copies))
	}
	if !wither[bolt] || !wither[copies[0]] {
		t.Fatalf("wither on the stack: original %v, copy %v; want both", wither[bolt], wither[copies[0]])
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := effectstest.MinusOneOn(t, g, id); got != 3 {
			t.Errorf("creature %s has %d -1/-1 counters, want 3", name, got)
		}
		if got := effectstest.DamageOn(t, g, id); got != 0 {
			t.Errorf("creature %s has %d marked damage, want 0", name, got)
		}
	}
}

// Wither damage to a player is ordinary life loss: no counters, no
// poison.
func TestSpinerockTyrantWitherDamageToAPlayerIsLifeLoss(t *testing.T) {
	g, me, opp, _, _ := spinerockSetup(t)
	at := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{at})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	answerCopyTarget(t, g, me.ID, at)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != game.StartingLife-6 || effectstest.PoisonOn(opp) != 0 {
		t.Errorf("opponent: %d life, %d poison; want %d and 0", opp.Life, effectstest.PoisonOn(opp), game.StartingLife-6)
	}
}

// Declining the copy leaves the original without wither: the creature
// takes marked damage.
func TestSpinerockTyrantDeclinedLeavesTheOriginalWithoutWither(t *testing.T) {
	g, me, _, a, _ := spinerockSetup(t)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{creatureRef(a)})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	if wither, copies := spellWither(t, g); wither[bolt] || len(copies) != 0 {
		t.Fatalf("after a decline: original wither %v, %d copies; want neither", wither[bolt], len(copies))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MinusOneOn(t, g, a) != 0 || effectstest.DamageOn(t, g, a) != 3 {
		t.Errorf("creature: %d counters, %d damage; want 0 and 3", effectstest.MinusOneOn(t, g, a), effectstest.DamageOn(t, g, a))
	}
}

// A spell with two targets doesn't trigger the Tyrant.
func TestSpinerockTyrantIgnoresASpellWithTwoTargets(t *testing.T) {
	g, me, _, a, b := spinerockSetup(t)
	effectstest.CastCatalogSpell(t, g, "Arc Trail", "Sorcery", effectstest.ArcTrailOracle, []game.TargetRef{creatureRef(a), creatureRef(b)})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
		t.Fatalf("the Tyrant asked about a two-target spell: %+v", p)
	}
	if effectstest.MinusOneOn(t, g, a) != 0 || effectstest.DamageOn(t, g, a) != 2 || effectstest.DamageOn(t, g, b) != 1 {
		t.Errorf("Arc Trail: A %d counters %d damage, B %d damage; want 0, 2, 1",
			effectstest.MinusOneOn(t, g, a), effectstest.DamageOn(t, g, a), effectstest.DamageOn(t, g, b))
	}
}

// A copy some other effect makes of the original is a copy of its
// copiable values only (CR 707.2): it has no wither, and deals marked
// damage, while the original and the Tyrant's copy still have it.
func TestSpinerockTyrantAnotherCopyOfTheSpellHasNoWither(t *testing.T) {
	g, me, _, a, b := spinerockSetup(t)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{creatureRef(a)})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	answerCopyTarget(t, g, me.ID, creatureRef(b))
	_, tyrantCopies := spellWither(t, g)
	if len(tyrantCopies) != 1 {
		t.Fatalf("%d copies after the Tyrant, want 1", len(tyrantCopies))
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bolt, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	wither, copies := spellWither(t, g)
	if len(copies) != 2 {
		t.Fatalf("%d copies, want 2", len(copies))
	}
	for _, id := range copies {
		want := id == tyrantCopies[0]
		if wither[id] != want {
			t.Errorf("copy %v (the Tyrant's: %v) has wither %v", id, want, wither[id])
		}
	}
	if !wither[bolt] {
		t.Error("the original lost wither")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MinusOneOn(t, g, a) != 3 || effectstest.DamageOn(t, g, a) != 3 {
		t.Errorf("A: %d counters, %d damage; want 3 (the original) and 3 (the other copy)", effectstest.MinusOneOn(t, g, a), effectstest.DamageOn(t, g, a))
	}
	if effectstest.MinusOneOn(t, g, b) != 3 || effectstest.DamageOn(t, g, b) != 0 {
		t.Errorf("B: %d counters, %d damage; want 3 and 0", effectstest.MinusOneOn(t, g, b), effectstest.DamageOn(t, g, b))
	}
}

// The original countered before the trigger resolves: the copy is made
// from last-known information and has wither; the original gains nothing.
func TestSpinerockTyrantCopiesACounteredSpellAndTheCopyHasWither(t *testing.T) {
	g, me, _, a, _ := spinerockSetup(t)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{creatureRef(a)})
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(bolt); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if effectstest.CopyPromptFor(g, me.ID) != nil {
		answerCopyTarget(t, g, me.ID, creatureRef(a))
	}
	wither, copies := spellWither(t, g)
	if len(copies) != 1 || !wither[copies[0]] {
		t.Fatalf("copies %v with wither %v; want one copy with wither", copies, wither)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MinusOneOn(t, g, a) != 3 || effectstest.DamageOn(t, g, a) != 0 {
		t.Errorf("A: %d counters, %d damage; want 3 and 0", effectstest.MinusOneOn(t, g, a), effectstest.DamageOn(t, g, a))
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

// Restore points: with the new-targets prompt for the copy open, and
// with the spell and its copy waiting on the stack. Both spells still
// have wither on the restored table and deal -1/-1 counters.
func TestSpinerockTyrantWitherSurvivesARestore(t *testing.T) {
	g, me, _, a, b := spinerockSetup(t)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{creatureRef(a)})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	answerCopyTarget(t, g, me.ID, creatureRef(b))
	g = effectstest.RestoreRoundTrip(t, g, true)

	wither, copies := spellWither(t, g)
	if len(copies) != 1 || !wither[bolt] || !wither[copies[0]] {
		t.Fatalf("restored stack: copies %v, wither %v; want the original and one copy, both with wither", copies, wither)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got, dmg := effectstest.MinusOneOn(t, g, id), effectstest.DamageOn(t, g, id); got != 3 || dmg != 0 {
			t.Errorf("creature %s: %d counters, %d damage; want 3 and 0", name, got, dmg)
		}
	}
}
