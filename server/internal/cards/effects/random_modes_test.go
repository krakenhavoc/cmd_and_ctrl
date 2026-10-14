package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// random_modes_test.go — #3399, CR 700.2: "choose one at random" on a
// triggered ability. The engine picks the bullet as the trigger goes on
// the stack; nobody is asked for the mode, and the same draw comes back
// after an undo.

const (
	cultOfSkaroOracle = "1cf40a30-5b42-4dce-ae4f-699bc879e180"
	umaroOracle       = "0f4350ad-dc44-4942-b097-e4033e29d659"
)

func randomModesKey(n byte) [32]byte {
	var k [32]byte
	for i := range k {
		k[i] = byte(i+1) ^ n
	}
	return k
}

// cultAttacks has a fresh Cult of Skaro attack and returns the mode the
// trigger went on the stack with.
func cultAttacks(t *testing.T, g *game.Game, cult uuid.UUID) []int {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cult, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Fatalf("a random mode asks nobody to choose: %+v", c)
	}
	item := triggerOnStack(g, cult)
	if item == nil {
		t.Fatal("Cult of Skaro's trigger is not on the stack")
	}
	return slices.Clone(item.Modes)
}

func cultGame(t *testing.T, key byte) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(key))
	cult := b12Push(g, g.Seats[0].ID, "Cult of Skaro", "Legendary Artifact Creature — Dalek", cultOfSkaroOracle, 4, 4)
	return g, cult
}

func TestCultOfSkaroPicksOneBulletWithoutAsking(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < 4; key++ {
		g, cult := cultGame(t, key)
		me, opp := g.Seats[0], g.Seats[1]
		hand, life, oppLife := me.Hand.Size(), me.Life, opp.Life
		modes := cultAttacks(t, g, cult)
		if len(modes) != 1 || modes[0] < 0 || modes[0] > 3 {
			t.Fatalf("key %d: the trigger carries exactly one of four modes, got %v", key, modes)
		}
		seen[modes[0]] = true
		passPriorityAroundTable(t, g)
		switch modes[0] {
		case 0:
			if n := counterCount(g, cult, "+1/+1"); n != 1 {
				t.Errorf("Thay: %d counters on the Cult, want 1", n)
			}
		case 1:
			if got := me.Hand.Size(); got != hand+2 {
				t.Errorf("Caan: hand %d, want %d", got, hand+2)
			}
		case 2:
			daleks := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == me.ID && c.Name == "Dalek" {
					daleks++
				}
			}
			if daleks != 1 {
				t.Errorf("Sec: %d Dalek tokens, want 1", daleks)
			}
		case 3:
			if opp.Life != oppLife-4 {
				t.Errorf("Jast: opponent at %d, want %d", opp.Life, oppLife-4)
			}
		}
		if modes[0] != 3 && opp.Life != oppLife {
			t.Errorf("key %d mode %d: the opponent lost life they should not have", key, modes[0])
		}
		if modes[0] != 1 && me.Hand.Size() != hand {
			t.Errorf("key %d mode %d: a card was drawn that the mode does not draw", key, modes[0])
		}
		_ = life
	}
	if len(seen) != 4 {
		t.Errorf("48 draws reached modes %v; a uniform pick reaches all four", seen)
	}
}

// The draw comes from the keyed stream, so an undo then a redo of the
// same attack names the same bullet.
func TestCultOfSkaroModeReplaysAfterUndo(t *testing.T) {
	for key := byte(0); key < 6; key++ {
		g, cult := cultGame(t, key)
		advanceTo(t, g, game.StepDeclareAttackers)
		pre := g.Clone()
		attack := func() []int {
			if err := g.DeclareAttacker(cult, g.Seats[1].ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			lockInAttacks(t, g)
			item := triggerOnStack(g, cult)
			if item == nil {
				t.Fatal("no trigger on the stack")
			}
			return slices.Clone(item.Modes)
		}
		first := attack()
		g.WithWriteLock(func() { g.RestoreFrom(pre) })
		second := attack()
		if !slices.Equal(first, second) {
			t.Errorf("key %d: undo then redo chose %v then %v", key, first, second)
		}
	}
}

func TestRandomModeIsLoggedAtRandom(t *testing.T) {
	g, cult := cultGame(t, 3)
	modes := cultAttacks(t, g, cult)
	want := g.Seats[0].ID
	found := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventOptionChosen && ev.CardID == cult && ev.Actor == want {
			found = true
		}
	}
	if !found {
		t.Errorf("mode %v was picked but the log has no entry for it", modes)
	}
}

// Umaro's third bullet targets: the target prompt follows the random
// pick, and only that bullet asks one.
func TestUmaroTargetsOnlyWhenTheDamageBulletIsDrawn(t *testing.T) {
	asked, skipped := 0, 0
	for key := byte(0); key < 40 && (asked == 0 || skipped == 0); key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		umaro := b12Push(g, me.ID, "Umaro, Raging Yeti", "Legendary Creature — Yeti Berserker", umaroOracle, 6, 6)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		if c := modePickChoiceFor(g, me.ID); c != nil {
			t.Fatalf("a random mode asks nobody to choose: %+v", c)
		}
		oppLife := opp.Life
		if p := latestPickTarget(g, me.ID); p != nil {
			asked++
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			passPriorityAroundTable(t, g)
			if opp.Life != oppLife-5 {
				t.Errorf("5 damage: opponent at %d, want %d", opp.Life, oppLife-5)
			}
			continue
		}
		skipped++
		item := triggerOnStack(g, umaro)
		if item == nil || len(item.Modes) != 1 || item.Modes[0] == 2 {
			t.Fatalf("with no target prompt the trigger holds a targetless bullet: %+v", item)
		}
		passPriorityAroundTable(t, g)
		if item.Modes[0] == 0 {
			if p := b12Card(t, g, bear).CurrentPower(); p != 5 {
				t.Errorf("the team pump: Bear power %d, want 5", p)
			}
			if p := b12Card(t, g, umaro).CurrentPower(); p != 6 {
				t.Errorf("Umaro pumps the others, not himself: power %d", p)
			}
		}
	}
	if asked == 0 || skipped == 0 {
		t.Errorf("40 draws should reach both the targeted bullet and a targetless one: asked %d, skipped %d", asked, skipped)
	}
}
