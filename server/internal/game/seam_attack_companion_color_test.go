package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// seam_attack_companion_color_test.go — Scarred Puma (S82 OLD batch 664,
// #5616): "This creature can't attack unless a black or green creature also
// attacks." AttackLimit.OthersColors narrows the NotAlone companion to the
// colors named, read at declaration (CR 508.1c).

func pushColoredCompanion(t *testing.T, g *Game, owner *Player, power, toughness int, colors ...string) uuid.UUID {
	t.Helper()
	id := pushSizedCreature(t, g, owner, power, toughness)
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			// The fixtures set the layered characteristic directly, so the
			// color goes there too.
			c.Colors = colors
			if c.effective != nil {
				c.effective.Colors = colors
			}
		}
	})
	return id
}

var scarredPumaLimit = AttackLimit{Scope: AttackLimitNotAlone, Max: 1, OthersColors: []string{"B", "G"}}

func TestScarredPumaNeedsABlackOrGreenCompanion(t *testing.T) {
	t.Run("a red companion is no company", func(t *testing.T) {
		stubCombatLimits(t, []AttackLimit{scarredPumaLimit}, nil)
		g := newActiveGame(t)
		me, them := g.Seats[0], g.Seats[1]
		puma := pushCompanionRuleCreature(t, g, me, 2, 1)
		red := pushColoredCompanion(t, g, me, 3, 3, "R")
		white := pushColoredCompanion(t, g, me, 3, 3, "W")
		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(puma, them.ID); !errors.Is(err, ErrAttackLimit) {
			t.Fatalf("Puma with only red and white companions = %v, want ErrAttackLimit", err)
		}
		for _, id := range []uuid.UUID{red, white} {
			if err := g.DeclareAttacker(id, them.ID); err != nil {
				t.Fatalf("the plain creature was refused: %v", err)
			}
		}
	})

	t.Run("a green companion makes it legal, a black one too", func(t *testing.T) {
		for _, color := range []string{"G", "B"} {
			stubCombatLimits(t, []AttackLimit{scarredPumaLimit}, nil)
			g := newActiveGame(t)
			me, them := g.Seats[0], g.Seats[1]
			puma := pushCompanionRuleCreature(t, g, me, 2, 1)
			friend := pushColoredCompanion(t, g, me, 1, 1, "R", color)
			advanceIntoStep(t, g, StepDeclareAttackers)
			if err := g.DeclareAttacker(puma, them.ID); err != nil {
				t.Fatalf("Puma with a %s companion available: %v", color, err)
			}
			var e *AttackLimitError
			if err := g.PassPriority(); !errors.As(err, &e) {
				t.Fatalf("pass with Puma attacking alone = %v, want *AttackLimitError", err)
			}
			if got := e.Sentence(me.ID); got != "Kw Creature can't attack unless a black or green creature also attacks." {
				t.Errorf("sentence = %q", got)
			}
			if err := g.DeclareAttacker(friend, them.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("pass with both attacking: %v", err)
			}
			if c := findCard(g, puma); c.AttackingTarget == uuid.Nil {
				t.Error("Puma was taken out of a legal declaration")
			}
		}
	})
}
