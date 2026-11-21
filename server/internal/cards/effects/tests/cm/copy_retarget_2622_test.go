package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCopyRetargetPromptAlwaysHasAnAcceptedAnswer(t *testing.T) {
	t.Run("one target, still legal", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.Pr7Creature(g, opp.ID, "Bear", 2)
		effectstest.Pr7Creature(g, me.ID, "Mine", 2)
		bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(bear)})
		effectstest.CopySpellNow(t, g, bolt, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "bolt")
		if !effectstest.HasFinal(finals, bear) {
			t.Error("re-picking the same target was not offered")
		}
	})

	// CR 707.10c: "The player may leave any number of the targets
	// unchanged, even if those targets would be illegal."
	t.Run("one target, gone", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.Pr7Creature(g, opp.ID, "Bear", 2)
		bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(bear)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
		effectstest.CopySpellNow(t, g, bolt, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "bolt")
		if !effectstest.HasFinal(finals, bear) {
			t.Error("leaving the now-illegal target unchanged was not offered (CR 707.10c)")
		}
	})

	t.Run("the chooser is not the active seat", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.Pr7Creature(g, opp.ID, "Bear", 2)
		effectstest.Pr7Creature(g, me.ID, "Mine", 2)
		bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(bear)})
		effectstest.CopySpellNow(t, g, bolt, opp.ID)
		if effectstest.CopyPromptFor(g, opp.ID) == nil {
			t.Fatal("the copy's controller was not asked")
		}
		effectstest.WalkCopyPrompt(t, g, opp.ID, 1, "bolt")
	})

	// The #2622 shape: two clauses, the first "target creature you
	// control".
	t.Run("Bite Down, two clauses", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.Pr7Creature(g, me.ID, "Mine", 3)
		mine2 := effectstest.Pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2)
		theirs2 := effectstest.Pr7Creature(g, opp.ID, "Theirs 2", 2)
		bite := effectstest.CastCopyTarget(t, g, me, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(mine, theirs)})
		effectstest.CopySpellNow(t, g, bite, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "bite")
		for _, want := range [][]uuid.UUID{{mine, theirs}, {mine2, theirs2}, {mine, theirs2}, {mine2, theirs}} {
			if !effectstest.HasFinal(finals, want...) {
				t.Errorf("the copy could not end up targeting %v", want)
			}
		}
	})

	// The copy's controller is not the original's, the original biter is
	// hexproof, and "you control" now means the copier.
	t.Run("Bite Down copied by an opponent, hexproof original", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.Pr7bCreature(g, me.ID, "Hexproof Mine", 3, "{G}", []string{"G"}, "hexproof")
		mine2 := effectstest.Pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2)
		bite := effectstest.CastCopyTarget(t, g, me, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(mine, theirs)})
		effectstest.CopySpellNow(t, g, bite, opp.ID)
		finals := effectstest.WalkCopyPrompt(t, g, opp.ID, 1, "bite")
		if !effectstest.HasFinal(finals, mine, theirs) {
			t.Error("leaving both targets unchanged was not offered (CR 707.10c)")
		}
		if !effectstest.HasFinal(finals, theirs, mine2) {
			t.Error("the copier could not aim its copy from its own creature at the caster's")
		}
	})

	t.Run("Bite Down, no new legal target anywhere", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.Pr7Creature(g, me.ID, "Mine", 3)
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2)
		bite := effectstest.CastCopyTarget(t, g, me, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(mine, theirs)})
		effectstest.CopySpellNow(t, g, bite, me.ID)
		if c := effectstest.CopyPromptFor(g, me.ID); c != nil {
			t.Errorf("prompted with nothing to change to: %q", c.Reason)
		}
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "bite")
		if !effectstest.HasFinal(finals, mine, theirs) {
			t.Errorf("the copy did not keep the original targets: %v", finals)
		}
	})

	t.Run("Dromoka's Command, targets in two modes", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2)
		effectstest.Pr7Creature(g, me.ID, "Mine 2", 2)
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 1)
		effectstest.Pr7Creature(g, opp.ID, "Theirs 2", 1)
		cmd := effectstest.CastCopyTarget(t, g, me, "Dromoka's Command", "Instant", effectstest.Pr7bDromoka, game.CastSpellParams{
			Modes: []int{2, 3},
			Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: mine, Mode: 0},
				{Kind: game.TargetCard, ID: mine, Mode: 1, Slot: 0},
				{Kind: game.TargetCard, ID: theirs, Mode: 1, Slot: 1},
			},
		})
		effectstest.CopySpellNow(t, g, cmd, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "command")
		if !effectstest.HasFinal(finals, mine, mine, theirs) {
			t.Error("leaving every target unchanged was not offered")
		}
	})

	t.Run("Switcheroo, one of two targets gone", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.Pr7Creature(g, me.ID, "A", 2)
		b := effectstest.Pr7Creature(g, opp.ID, "B", 2)
		c := effectstest.Pr7Creature(g, opp.ID, "C", 2)
		sw := effectstest.CastCopyTarget(t, g, me, "Switcheroo", "Sorcery", switcherooOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(a, b)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
		effectstest.CopySpellNow(t, g, sw, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "switcheroo")
		if !effectstest.HasFinal(finals, a, b) {
			t.Error("leaving both targets unchanged was not offered (CR 707.10c)")
		}
		if !effectstest.HasFinal(finals, c, b) {
			t.Error("replacing the gone target alone was not offered")
		}
	})

	// Only the surviving original is still legal: the copy keeps its
	// targets without a prompt nobody could answer.
	t.Run("Switcheroo, one target gone and no replacement", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.Pr7Creature(g, me.ID, "A", 2)
		b := effectstest.Pr7Creature(g, opp.ID, "B", 2)
		sw := effectstest.CastCopyTarget(t, g, me, "Switcheroo", "Sorcery", switcherooOracle,
			game.CastSpellParams{Targets: effectstest.CardTargets(a, b)})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
		effectstest.CopySpellNow(t, g, sw, me.ID)
		finals := effectstest.WalkCopyPrompt(t, g, me.ID, 1, "switcheroo")
		if !effectstest.HasFinal(finals, a, b) {
			t.Errorf("the copy did not keep its targets: %v", finals)
		}
	})
}

// CR 115.7f: choosing new targets for a copy cannot change the original
// division. The copy's answer is a set: the target it keeps holds its
// share, and the new one takes the share of the target it replaced.
func TestCopyRetargetKeepsTheDivision(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	id := effectstest.B12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", effectstest.ShatterskullOracle,
		game.CastSpellParams{XValue: 4, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3}})
	effectstest.CopySpellNow(t, g, id, me.ID)
	p := effectstest.CopyPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("no copy prompt")
	}
	if p.PickTargetMin != 2 || p.PickTargetMax != 2 {
		t.Errorf("the copy is asked for %d..%d targets, want exactly the original's 2", p.PickTargetMin, p.PickTargetMax)
	}
	// Listed with the kept target first: it stays in its own slot.
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(b, c)); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := []int{effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked, effectstest.E2Card(t, g, c).DamageMarked}
	if got[0] != 1 || got[1] != 6 || got[2] != 1 {
		t.Errorf("damage on A/B/C = %v, want [1 6 1]: the original deals 1/3 to A/B and the copy 1/3 to C/B", got)
	}
}
