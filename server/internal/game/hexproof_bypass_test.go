package game

import (
	"testing"

	"github.com/google/uuid"
)

// hexproof_bypass_test.go — #1560, the engine half of "as though it
// didn't have hexproof". The catalog half (Nowhere to Run, Kaya, Bane
// of the Dead) is in cards/effects/nowhere_to_run_test.go.

const waiverOracle = "test-hexproof-waiver"

// withHexproofBypasses stubs the catalog hook for one test, returning
// a pointer to the number of times a waiver predicate was asked.
func withHexproofBypasses(t *testing.T, b HexproofBypass) *int {
	t.Helper()
	calls := new(int)
	wrapped := b
	if b.Permanent != nil {
		wrapped.Permanent = func(g *Game, target, source *Card) bool {
			*calls++
			return b.Permanent(g, target, source)
		}
	}
	prev := CatalogHexproofBypasses
	CatalogHexproofBypasses = func(key string) []HexproofBypass {
		if key == waiverOracle {
			return []HexproofBypass{wrapped}
		}
		return nil
	}
	t.Cleanup(func() { CatalogHexproofBypasses = prev })
	return calls
}

func opponentsCreaturesWaiver(yoursOnly bool) HexproofBypass {
	return HexproofBypass{
		Permanent: func(_ *Game, target, source *Card) bool {
			return target.IsCreature() && target.Controller != source.Controller
		},
		YoursOnly: yoursOnly,
	}
}

func pushWaiver(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Waiver", owner.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = waiverOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func legalCreatureTargets(g *Game, by uuid.UUID) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, id := range g.LegalTargetsForEffect(SourceChooser(by), anyCreatureSpec()).Cards {
		out[id] = true
	}
	return out
}

// "You control" (Glaring Spotlight, Kaya) waives hexproof for the
// static's controller alone; the unrestricted form (Nowhere to Run)
// for every player.
func TestHexproofBypassYoursOnlyScopesTheSource(t *testing.T) {
	for _, tc := range []struct {
		name      string
		yoursOnly bool
		thirdOK   bool
	}{
		{"spells and abilities", false, true},
		{"spells and abilities you control", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withHexproofBypasses(t, opponentsCreaturesWaiver(tc.yoursOnly))
			g := newFourPlayerActiveGame(t)
			me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
			theirs := pushProtectedCreature(g, opp, "Their Bogle", "hexproof")
			pushWaiver(g, me)

			if !legalCreatureTargets(g, me.ID)[theirs] {
				t.Error("the waiver's controller should be able to target it")
			}
			if got := legalCreatureTargets(g, third.ID)[theirs]; got != tc.thirdOK {
				t.Errorf("third player may target: got %v, want %v", got, tc.thirdOK)
			}
		})
	}
}

// The exported CanBeTargetedBy has no game and waives nothing; the
// choke point's own form does. And the waiver is asked only when
// hexproof would refuse — a board with no hexproof never walks for one.
func TestHexproofBypassIsReadOnlyAtTheChokePoint(t *testing.T) {
	calls := withHexproofBypasses(t, opponentsCreaturesWaiver(false))
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushProtectedCreature(g, opp, "Plain Bear")
	pushWaiver(g, me)

	if !legalCreatureTargets(g, me.ID)[plain] {
		t.Fatal("a plain creature must be a legal target")
	}
	if *calls != 0 {
		t.Errorf("the waiver was asked %d times about creatures without hexproof", *calls)
	}

	theirs := pushProtectedCreature(g, opp, "Their Bogle", "hexproof")
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != theirs {
			continue
		}
		if CanBeTargetedBy(c, ZoneBattlefield, SourceChooser(me.ID)) {
			t.Error("the game-less CanBeTargetedBy must not honour a waiver")
		}
		if !g.canBeTargetedByLocked(c, ZoneBattlefield, SourceChooser(me.ID)) {
			t.Error("the choke point must honour the waiver")
		}
	}
}
