package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_gain_life_test.go — ADR 0107 §5 decision 6 (#1880): CR 119.7's
// "can't gain life", asked before the life window opens.

const cantGainLifeOracle = "test-cant-gain-life-static"

func withCantGainLifeStatics(t *testing.T, statics ...CantGainLifeStatic) {
	t.Helper()
	prev := CatalogCantGainLife
	CatalogCantGainLife = func(key string) []CantGainLifeStatic {
		if key == cantGainLifeOracle {
			return statics
		}
		return nil
	}
	t.Cleanup(func() { CatalogCantGainLife = prev })
}

func pushCantGainLifeSource(g *Game, controller *Player) uuid.UUID {
	c := NewCard("No Gain", controller.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = cantGainLifeOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func lifeChangesFor(g *Game, from int, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventChangeLife && ev.Target == player {
			n++
		}
	}
	return n
}

// "Players can't gain life this turn": nothing is gained, no
// EventChangeLife fires for "whenever you gain life" to see, a loss still
// happens, and the continuation is told zero.
func TestCantGainLifeStopsTheGainAndItsEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	start := me.Life
	var told = -1
	cut := len(g.Events)
	g.WithWriteLock(func() {
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeEveryone, g.UntilEndOfTurnDuration(), "Skullcrack")
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, me.ID, 4, func(_ *Game, applied int) error {
			told = applied
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	if me.Life != start {
		t.Fatalf("life %d after a forbidden gain, want %d", me.Life, start)
	}
	if told != 0 {
		t.Errorf("continuation told %d, want 0", told)
	}
	if n := lifeChangesFor(g, cut, me.ID); n != 0 {
		t.Errorf("%d EventChangeLife for a gain that can't happen, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	if me.Life != start-3 {
		t.Errorf("life %d after a loss, want %d — only gains are stopped", me.Life, start-3)
	}
}

// CR 119.7 / 119.10: a replacement that would replace a life gain does
// nothing — it never sees the gain, so it neither runs nor asks.
func TestCantGainLifeComesBeforeLifeReplacements(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	start := me.Life
	ran := false
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifeDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ran = true
				ev.LifeDelta *= 2
				return nil
			},
			Label: "doubler",
		})
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeEveryone, g.UntilEndOfTurnDuration(), "Skullcrack")
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	if ran {
		t.Error("an \"if you would gain life\" replacement ran on a gain that can't happen")
	}
	if me.Life != start {
		t.Errorf("life %d, want %d", me.Life, start)
	}
}

// "Your opponents can't gain life this turn" covers the opponents of the
// record's controller and not the controller.
func TestOpponentsCantGainLifeThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myStart, oppStart := me.Life, opp.Life
	g.WithWriteLock(func() {
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeOpponents, g.UntilEndOfTurnDuration(), "Atarka's Command")
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2)
	})
	if me.Life != myStart+2 {
		t.Errorf("controller's life %d, want %d", me.Life, myStart+2)
	}
	if opp.Life != oppStart {
		t.Errorf("opponent's life %d, want %d", opp.Life, oppStart)
	}
}

// "This turn" ends at cleanup (CR 514.2); "for the rest of the game"
// does not (CR 611.2a).
func TestCantGainLifeDurations(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myStart, oppStart := me.Life, opp.Life
	g.WithWriteLock(func() {
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeEveryone, g.UntilEndOfTurnDuration(), "Skullcrack")
		g.PlayerCantGainLifeForEffect(uuid.Nil, opp.ID, IndefiniteDuration(), "Screaming Nemesis")
		g.sweepScopedEffectsLocked(true)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2)
	})
	if me.Life != myStart+2 {
		t.Errorf("the turn grant outlived cleanup: life %d, want %d", me.Life, myStart+2)
	}
	if opp.Life != oppStart {
		t.Errorf("the rest-of-the-game grant ended at cleanup: life %d, want %d", opp.Life, oppStart)
	}
}

// The battlefield statics, each against whom it names.
func TestCantGainLifeStatics(t *testing.T) {
	cases := []struct {
		name        string
		whose       CantGainLifeWhose
		mine, their bool // whether the static's controller / its opponent can still gain
	}{
		{"players (Leyline of Punishment)", CantGainLifeEveryone, false, false},
		{"your opponents (Erebos)", CantGainLifeOpponents, true, false},
		{"enchanted player (Grievous Wound)", CantGainLifeEnchantedPlayer, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withCantGainLifeStatics(t, CantGainLifeStatic{Whose: tc.whose})
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushCantGainLifeSource(g, me)
			if tc.whose == CantGainLifeEnchantedPlayer {
				g.WithWriteLock(func() {
					i := findCardOnBattlefield(g, src)
					g.Battlefield.Cards[i].AttachedTo = TargetRef{Kind: TargetPlayer, ID: opp.ID}
				})
			}
			myStart, oppStart := me.Life, opp.Life
			g.WithWriteLock(func() {
				_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1)
				_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 1)
			})
			if got := me.Life == myStart+1; got != tc.mine {
				t.Errorf("controller gained = %v, want %v", got, tc.mine)
			}
			if got := opp.Life == oppStart+1; got != tc.their {
				t.Errorf("opponent gained = %v, want %v", got, tc.their)
			}
		})
	}
}

// Lifelink damage is still dealt; only the gain is stopped (CR 702.15b
// makes it a life gain, and CR 119.7 stops a life gain).
func TestLifelinkDamageUnderCantGainLife(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myStart, oppStart := me.Life, opp.Life
	linker := pushColouredCreature(g, me, "Lifelinker", []string{"W"}, "lifelink")
	g.WithWriteLock(func() {
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeEveryone, g.UntilEndOfTurnDuration(), "Skullcrack")
		if err := g.DealDamageToPlayerForEffect(linker, opp.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Life != oppStart-3 {
		t.Errorf("opponent's life %d, want %d — the damage is still dealt", opp.Life, oppStart-3)
	}
	if me.Life != myStart {
		t.Errorf("lifelinker's controller's life %d, want %d — no gain", me.Life, myStart)
	}
}

// "If that player would gain life this turn, that player gains no life
// instead" (Flames of the Blood Hand) is a replacement for one player.
func TestGainNoLifeThisTurnReplacesOnePlayersGains(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myStart, oppStart := me.Life, opp.Life
	cut := len(g.Events)
	g.WithWriteLock(func() {
		if !g.GainNoLifeThisTurnForEffect(uuid.Nil, opp.ID, "Flames of the Blood Hand") {
			t.Fatal("no record")
		}
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
	})
	if opp.Life != oppStart-1 {
		t.Errorf("opponent's life %d, want %d (no gain, the loss lands)", opp.Life, oppStart-1)
	}
	if me.Life != myStart+3 {
		t.Errorf("my life %d, want %d", me.Life, myStart+3)
	}
	if n := lifeChangesFor(g, cut, opp.ID); n != 1 {
		t.Errorf("%d EventChangeLife for the opponent, want 1 (the loss)", n)
	}
}
