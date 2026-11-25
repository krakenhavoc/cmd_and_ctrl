package game

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// tap_foreign_lands_test.go — #6204, ADR 0106 §1 amendment 2026-11-25:
// "Until end of turn, you may tap lands you don't control for mana. Spend
// this mana only to cast spells." (Piracy). CR 602.2: only an object's
// controller can activate its activated ability "unless the object
// specifically says otherwise"; CR 605.3a: a mana ability is activated
// whenever the player has priority or is paying a cost; CR 106.12: to tap
// a permanent for mana is to activate a mana ability with {T} in its cost.

func allowPiracy(g *Game, p *Player) {
	g.WithWriteLock(func() { g.AllowTapForeignLandsThisTurnForEffect(p.ID) })
}

func piracyBoard(t *testing.T) (*Game, *Player, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	mountain := pushBattlefieldForTest(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	return g, me, opp, mountain
}

// Without the permission the opponent's Mountain is not the player's to
// tap (CR 602.2).
func TestForeignLandIsNotYoursToTapWithoutThePermission(t *testing.T) {
	g, me, _, mountain := piracyBoard(t)
	err := g.ActivateManaAbility(me.ID, mountain, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("activate = %v, want ErrCardCallerMismatch", err)
	}
	if tappedOnBattlefield(t, g, mountain) {
		t.Error("a refused activation tapped the Mountain")
	}
}

// The acceptance test: after the effect, its controller taps an
// opponent's Mountain for {R}, the mana is theirs and may pay for a
// spell, and it can't pay an activated ability.
func TestForeignLandManaGoesToTheActivatorForSpellsOnly(t *testing.T) {
	g, me, opp, mountain := piracyBoard(t)
	allowPiracy(g, me)
	if err := g.ActivateManaAbility(me.ID, mountain, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !tappedOnBattlefield(t, g, mountain) {
		t.Fatal("the Mountain was not tapped")
	}
	if opp.ManaPool.Total() != 0 {
		t.Errorf("the land's controller got mana: %+v", opp.ManaPool)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("activator's pool = %d, want 1", me.ManaPool.Total())
	}
	tok := me.ManaPool[0]
	if tok.Color != "R" || !slices.Contains(tok.Restrictions, ManaRestrictCast) {
		t.Errorf("token = %+v, want {R} restricted to casting spells", tok)
	}
	cost, err := ParseCost("{R}")
	if err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("the mana can pay an activated ability's cost")
	}
	spell := NewCard("Shock", me.ID)
	spell.TypeLine, spell.ManaCost, spell.Controller = "Instant", "{R}", me.ID
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with the mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("mana left after the cast: %+v", me.ManaPool)
	}
}

// The land's controller still taps it as ever, and their mana is not
// restricted: the permission is the other player's.
func TestForeignLandPermissionLeavesTheControllersManaAlone(t *testing.T) {
	g, me, opp, mountain := piracyBoard(t)
	allowPiracy(g, me)
	if err := g.ActivateManaAbility(opp.ID, mountain, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("controller's activation: %v", err)
	}
	if len(opp.ManaPool) != 1 || len(opp.ManaPool[0].Restrictions) != 0 {
		t.Errorf("controller's pool = %+v, want one unrestricted {R}", opp.ManaPool)
	}
}

// Only lands: another player's mana creature is not reachable.
func TestForeignLandPermissionIsForLandsOnly(t *testing.T) {
	g, me, opp, _ := piracyBoard(t)
	allowPiracy(g, me)
	elf := pushIntrinsicPermanent(g, opp, "Elf", "Creature — Elf Druid",
		[]ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}, nil)
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("activate the opponent's Elf = %v, want ErrCardCallerMismatch", err)
	}
}

// "Tap … for mana" (CR 106.12): a land's mana ability without {T} stays
// its controller's.
func TestForeignLandPermissionReachesOnlyTapRows(t *testing.T) {
	g, me, opp, _ := piracyBoard(t)
	allowPiracy(g, me)
	land := pushIntrinsicPermanent(g, opp, "Odd Land", "Land", []ManaAbilityShape{
		{TapCost: true, Produced: "{C}", Label: "{T}: Add {C}"},
		{LifeCost: 1, Produced: "{B}", Label: "Pay 1 life: Add {B}"},
	}, nil)
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(land)
		abs := ManaAbilitiesForCard(*c)
		if !g.MayActivateManaLocked(me.ID, *c, ZoneBattlefield, abs[0]) {
			t.Error("the {T} row is not reachable")
		}
		if g.MayActivateManaLocked(me.ID, *c, ZoneBattlefield, abs[1]) {
			t.Error("the row without {T} is reachable")
		}
	})
	if err := g.ActivateManaAbility(me.ID, land, 1, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("activate the life row = %v, want ErrCardCallerMismatch", err)
	}
}

// The auto-tapper plans the opponent's land for a cast, and for nothing
// else while the player has no mana of their own.
func TestAutoPayTapsAForeignLandForASpell(t *testing.T) {
	t.Run("with the permission", func(t *testing.T) {
		g, me, _, mountain := piracyBoard(t)
		allowPiracy(g, me)
		if _, err := tryAutoCast(g, me, "Shock", "Instant", "{R}"); err != nil {
			t.Fatalf("auto-cast: %v", err)
		}
		if !tappedOnBattlefield(t, g, mountain) {
			t.Error("the opponent's Mountain was not tapped")
		}
	})
	t.Run("without it", func(t *testing.T) {
		g, me, _, mountain := piracyBoard(t)
		_, err := tryAutoCast(g, me, "Shock", "Instant", "{R}")
		var short *InsufficientManaError
		if !errors.As(err, &short) {
			t.Fatalf("auto-cast = %v, want insufficient mana", err)
		}
		if tappedOnBattlefield(t, g, mountain) {
			t.Error("the opponent's Mountain was tapped")
		}
	})
}

// The permission survives a snapshot round trip and ends with the turn.
func TestForeignLandPermissionSurvivesASnapshotAndEndsWithTheTurn(t *testing.T) {
	g, me, _, _ := piracyBoard(t)
	allowPiracy(g, me)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.TapForeignLandsThisTurn[me.ID] {
		t.Error("the restored game lost the permission")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.TapForeignLandsThisTurn != nil {
		t.Errorf("the turn boundary kept %v", g.TapForeignLandsThisTurn)
	}
}
