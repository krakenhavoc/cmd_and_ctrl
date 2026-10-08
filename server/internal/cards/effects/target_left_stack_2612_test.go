package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2612: a counter or copy whose stack object has already left does
// nothing (CR 608.2b) — it must not fail the resolution with "card is
// not on the stack".

// The spell Wild Unraveling targets is countered away before it resolves.
func TestWildUnravelingTargetGoneIsNotAnError(t *testing.T) {
	g := newCatalogGame(t)
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	if _, err := castWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", wildUnravelingOracle,
		game.CastSpellParams{CostBranch: branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(victim); err != nil {
			t.Fatalf("counter the victim: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !stackFullyEmpty(g) {
		t.Fatal("the stack did not settle")
	}
	assertNoEffectErrors(t, g)
}

// Wild Unraveling aimed at a COPY of a spell: a copy has a stack record
// and a card, and the counter has to find both.
func TestWildUnravelingCountersACopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(victim, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	var cp game.TargetRef
	for _, c := range g.Stack.Cards {
		if c.InstanceID != victim {
			cp = game.TargetRef{Kind: game.TargetCard, ID: c.InstanceID}
		}
	}
	if cp.ID == [16]byte{} {
		t.Fatal("no copy on the stack")
	}
	if _, err := castWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", wildUnravelingOracle,
		game.CastSpellParams{CostBranch: branch(1), Targets: []game.TargetRef{cp}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertNoEffectErrors(t, g)
}

// Wild Unraveling's body (counterTheTargetSpell, shared by every plain
// "counter target spell") handed an item whose target is no longer on the
// stack. The soak (seed 1051) reached this through a target that got past
// the pre-resolution fizzle check, so the body itself has to cope.
func TestCounterTheTargetSpellWithTargetGoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	item := &game.StackItem{
		Kind:       game.StackItemSpell,
		Controller: me.ID,
		Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: uuid.New()}},
	}
	g.WithWriteLock(func() {
		if err := counterTheTargetSpell(item, NewContext(g, item)); err != nil {
			t.Errorf("counterTheTargetSpell with the target gone: %v", err)
		}
	})
}

// The spell Jin-Gitaxias's trigger names leaves the stack from under it.
func TestJinGitaxiasTriggerWithSpellGoneIsNotAnError(t *testing.T) {
	for _, mine := range []bool{true, false} {
		name := "counter"
		if mine {
			name = "copy"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushCatalogPermanent(g, me.ID, "Jin-Gitaxias, Progress Tyrant", "Legendary Creature — Phyrexian Praetor", jinGitaxiasOracle, false)
			if !mine {
				for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
					if _, err := g.AdvanceStep(); err != nil {
						t.Fatalf("AdvanceStep: %v", err)
					}
				}
			}
			bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			g.WithWriteLock(func() {
				if err := g.CounterTargetForEffect(bolt); err != nil {
					t.Fatalf("counter the bolt: %v", err)
				}
			})
			for i := 0; i < 32 && !stackFullyEmpty(g); i++ {
				if p := latestPickTarget(g, me.ID); p != nil {
					pickPlayer(t, g, me.ID, opp.ID)
					continue
				}
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if !stackFullyEmpty(g) {
				t.Fatal("the stack did not settle")
			}
			assertNoEffectErrors(t, g)
		})
	}
}

// assertNoEffectErrors fails on any EventEffectError: the engine survives
// a failed effect and logs it, so a test that only checks the stack
// would pass over exactly the bug #2612 was.
func assertNoEffectErrors(t *testing.T, g *game.Game) {
	t.Helper()
	for _, ev := range g.Events {
		if ev.Kind == game.EventEffectError {
			t.Errorf("effect error during resolution: %s", ev.ErrorMsg)
		}
	}
}
