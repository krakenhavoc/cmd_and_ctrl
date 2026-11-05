package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #5084: the real Burnished Hart, through the real enumerator, view and
// policy, in the bot's precombat main phase with only zero-drops in hand,
// so its mana is spare. In round 4 with four Islands, with
// PriceRampCurve on the heuristic cracks the Hart and two basic Islands
// from its library enter tapped; with it off (DefaultConfig: the curve
// won no more in the arena) it keeps the Hart, as #4991's arena saw. In
// round 10 with eight Islands, facing two 7/7s, it keeps the Hart with
// the curve on.
func TestHeuristicCracksBurnishedHartEarly(t *testing.T) {
	for _, tc := range []struct {
		name         string
		round, lands int
		curve, crack bool
	}{
		{"round 4, curve on", 4, 4, true, true},
		{"round 4, curve off", 4, 4, false, false},
		{"round 10, curve on", 10, 8, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newSettledTable(t, 5084)
			advanceToStep(t, g, game.StepPrecombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			var them *game.Player
			for _, s := range g.Seats {
				if s.ID != me.ID {
					them = s
				}
			}
			var hart uuid.UUID
			g.WithWriteLock(func() {
				g.Turn.Round = tc.round
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				for range 2 {
					c := game.Card{InstanceID: uuid.New(), Name: "Ornithopter", OracleID: "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba",
						TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}", Power: 0, Toughness: 2, Keywords: []string{"flying"},
						Owner: me.ID, Controller: me.ID}
					c.AddKnowersAll(seatIDs(g))
					me.Hand.PushTop(c)
				}
				hart = seedBattlefield(g, me, game.Card{Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3",
					TypeLine: "Artifact Creature — Elk", ManaCost: "{3}", Power: 2, Toughness: 2})
				for range tc.lands {
					seedBattlefield(g, me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
				}
				if tc.round > 7 {
					for range 2 {
						seedBattlefield(g, them, game.Card{Name: "Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}",
							Power: 7, Toughness: 7, Keywords: []string{"trample"}})
					}
				}
			})
			library := stockSearchLibrary(me,
				game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
				game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
				game.Card{Name: "Drake", TypeLine: "Creature — Drake", ManaCost: "{2}{U}", Power: 3, Toughness: 3},
				game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
			)
			g.BumpLayerVersionForTest()

			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Kind == legal.KindActivate && m.Source == hart {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real Burnished Hart activation offered")
			}
			lands := countIslands(g, me.ID)
			cfg := heuristic.DefaultConfig()
			cfg.PriceRampCurve = tc.curve
			pol := heuristic.NewWithConfig(cfg)
			activated := 0
			// The main phase, with the Hart's search answered by the
			// policy when it opens.
			for i := 0; i < 6 && g.Turn.Step == game.StepPrecombatMain; i++ {
				for _, m := range playHeliodMainPhase(t, g, me, pol) {
					if m.move.Kind == legal.KindActivate && m.move.Source == hart {
						activated++
					}
				}
				if !owesChoice(g, me.ID) {
					break
				}
				driveChoices(t, g, pol, me.ID, 6)
			}
			if !tc.crack {
				if activated != 0 || !g.Battlefield.Contains(hart) {
					t.Errorf("activated the Hart %d times; want it kept", activated)
				}
				return
			}
			if activated != 1 || g.Battlefield.Contains(hart) {
				t.Fatalf("activated the Hart %d times (still on the battlefield: %v); want it cracked once", activated, g.Battlefield.Contains(hart))
			}
			found, tapped := 0, 0
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					for _, id := range library {
						if c.InstanceID == id {
							found++
							if c.Tapped {
								tapped++
							}
						}
					}
				}
			})
			if got := countIslands(g, me.ID); got != lands+2 || found != 2 || tapped != 2 {
				t.Errorf("Islands went from %d to %d, %d from the library (%d tapped); want two more, entered tapped", lands, got, found, tapped)
			}
		})
	}
}

// countIslands is how many Islands seat controls on the battlefield.
func countIslands(g *game.Game, seat uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == seat && c.Name == "Island" {
				n++
			}
		}
	})
	return n
}
