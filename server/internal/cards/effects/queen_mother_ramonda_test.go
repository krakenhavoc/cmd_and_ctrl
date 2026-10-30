package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// queen_mother_ramonda_test.go — #4659: "When Queen Mother Ramonda enters,
// you become the monarch. As long as you're the monarch, creatures with
// power 2 or less can't attack you."

const queenMotherRamondaOracle = "f5619a13-b229-4202-a217-bf6509706e77"

// ramondaTable is seat 1 controlling Ramonda (the monarch) and a
// planeswalker; seat 0 (the active seat) has a power-2 Bear and a power-3
// Ogre ready to attack.
type ramondaTable struct {
	g         *game.Game
	s0, s1    *game.Player
	walker    uuid.UUID
	bear, big uuid.UUID
}

func newRamondaTable(t *testing.T) *ramondaTable {
	t.Helper()
	g := newCatalogGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want seat 0", g.Turn.ActiveSeat)
	}
	tb := &ramondaTable{g: g, s0: g.Seats[0], s1: g.Seats[1]}
	pushCatalogPermanent(g, tb.s1.ID, "Queen Mother Ramonda", "Legendary Creature — Human Noble", queenMotherRamondaOracle, false)
	tb.walker = pushWalkerForTest(g, tb.s1.ID, "Test Walker", "", 3)
	tb.bear = b12Creature(g, tb.s0.ID, "Bear", "Creature — Bear", 2, 2)
	tb.big = b12Creature(g, tb.s0.ID, "Ogre", "Creature — Ogre", 3, 3)
	return tb
}

// While Ramonda's controller is the monarch, a power-2 creature can't attack
// them, and the refusal says so, but it may attack their planeswalker ("you",
// not "you or planeswalkers you control"). A power-3 creature attacks either.
func TestQueenMotherRamondaPowerTwoCantAttackTheMonarchButPowerThreeCan(t *testing.T) {
	tb := newRamondaTable(t)
	monCrown(t, tb.g, tb.s1.ID)
	advanceToDeclareAttackersOf(t, tb.g, 0)

	small := arbTargets(t, tb.g, tb.s0.ID, tb.bear)
	if small[tb.s1.ID.String()] {
		t.Fatalf("a power-2 creature is offered the monarch: %v", small)
	}
	if !small[tb.walker.String()] {
		t.Fatalf("a power-2 creature is not offered the monarch's planeswalker: %v", small)
	}
	err := tb.g.Clone().DeclareAttacker(tb.bear, tb.s1.ID)
	var refused *game.AttackTargetRestrictionError
	if !errors.As(err, &refused) || !errors.Is(err, game.ErrIllegalAttackTarget) ||
		!strings.Contains(refused.Sentence(), "Bear can't attack "+tb.s1.Name) || strings.Contains(refused.Sentence(), "planeswalkers") {
		t.Fatalf("power-2 creature attacking the monarch: err = %v, want a players-only refusal", err)
	}
	if err := tb.g.Clone().DeclareAttacker(tb.bear, tb.walker); err != nil {
		t.Fatalf("power-2 creature attacking the planeswalker: %v", err)
	}

	big := arbTargets(t, tb.g, tb.s0.ID, tb.big)
	if !big[tb.s1.ID.String()] || !big[tb.walker.String()] {
		t.Fatalf("a power-3 creature is refused something: %v", big)
	}
	if err := tb.g.Clone().DeclareAttacker(tb.big, tb.s1.ID); err != nil {
		t.Fatalf("power-3 creature attacking the monarch: %v", err)
	}
}

// Power is judged when attackers are declared, on the layers and counters:
// a power-2 creature pumped to 3 may attack, and a power-3 one shrunk to 2
// may not.
func TestQueenMotherRamondaJudgesPowerLiveAtDeclaration(t *testing.T) {
	tb := newRamondaTable(t)
	g := tb.g
	monCrown(t, g, tb.s1.ID)
	g.WithWriteLock(func() {
		for id, mod := range map[uuid.UUID]game.Mod{tb.bear: game.ModifyPTMod(1, 0), tb.big: game.ModifyPTMod(-1, 0)} {
			if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{mod}, Duration: game.IndefiniteDuration()}).Apply(NewContext(g, nil)); err != nil {
				t.Fatal(err)
			}
		}
	})
	advanceToDeclareAttackersOf(t, g, 0)

	if got := arbTargets(t, g, tb.s0.ID, tb.bear); !got[tb.s1.ID.String()] {
		t.Fatalf("a power-2 creature pumped to 3 is refused the monarch: %v", got)
	}
	if err := g.Clone().DeclareAttacker(tb.bear, tb.s1.ID); err != nil {
		t.Fatalf("pumped creature attacking the monarch: %v", err)
	}
	if got := arbTargets(t, g, tb.s0.ID, tb.big); got[tb.s1.ID.String()] {
		t.Fatalf("a power-3 creature shrunk to 2 is offered the monarch: %v", got)
	}
	if err := g.Clone().DeclareAttacker(tb.big, tb.s1.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("shrunk creature attacking the monarch: err = %v, want ErrIllegalAttackTarget", err)
	}
}

// "As long as you're the monarch": with nobody crowned, or the crown on
// another player, a power-2 creature attacks Ramonda's controller; the
// restriction comes back when they take the crown again.
func TestQueenMotherRamondaOnlyWhileItsControllerIsTheMonarch(t *testing.T) {
	tb := newRamondaTable(t)
	g := tb.g
	advanceToDeclareAttackersOf(t, g, 0)
	if got := arbTargets(t, g, tb.s0.ID, tb.bear); !got[tb.s1.ID.String()] {
		t.Fatalf("with no monarch a power-2 creature is refused Ramonda's controller: %v", got)
	}

	monCrown(t, g, tb.s1.ID)
	if got := arbTargets(t, g, tb.s0.ID, tb.bear); got[tb.s1.ID.String()] {
		t.Fatalf("setup: a power-2 creature is offered the monarch: %v", got)
	}

	monCrown(t, g, g.Seats[2].ID)
	if got := arbTargets(t, g, tb.s0.ID, tb.bear); !got[tb.s1.ID.String()] {
		t.Fatalf("after the crown passed on, a power-2 creature is still refused Ramonda's controller: %v", got)
	}
	if err := g.Clone().DeclareAttacker(tb.bear, tb.s1.ID); err != nil {
		t.Fatalf("power-2 creature attacking after the crown passed: %v", err)
	}

	monCrown(t, g, tb.s1.ID)
	if got := arbTargets(t, g, tb.s0.ID, tb.bear); got[tb.s1.ID.String()] {
		t.Fatalf("with the crown back, a power-2 creature is offered the monarch: %v", got)
	}
}

// Ramonda's controller is the one who is protected, not the monarch: the
// crown on another player doesn't give that player the restriction, and
// Ramonda's own controller keeps being a legal target for power 3.
func TestQueenMotherRamondaProtectsOnlyItsController(t *testing.T) {
	tb := newRamondaTable(t)
	monCrown(t, tb.g, tb.g.Seats[2].ID)
	advanceToDeclareAttackersOf(t, tb.g, 0)
	if got := arbTargets(t, tb.g, tb.s0.ID, tb.bear); !got[tb.g.Seats[2].ID.String()] {
		t.Fatalf("a power-2 creature is refused another player who holds the crown: %v", got)
	}
}

// Her enters trigger makes you the monarch.
func TestQueenMotherRamondaEntersMakesYouTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Queen Mother Ramonda", "Legendary Creature — Human Noble", queenMotherRamondaOracle)
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want Ramonda's controller %v", g.Monarch, me.ID)
	}
}
