package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_target_restrictions_test.go — the two catalog constructors of
// ADR 0106 §2 (#1794), on test specs in Xantcha's and Alexios's shapes
// (neither card lands until its other seams do). The engine half is
// pinned in game/attack_target_restrictions_test.go.

const (
	testSleeperOracle  = "test-effects-sleeper-agent"
	testChampionOracle = "test-effects-exiled-champion"
)

func registerOwnerRestrictionSpecs(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{OracleID: testSleeperOracle, Name: "Sleeper Agent", Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.AttacksEachCombat(), effects.CantAttackItsOwnerOrItsOwnersPlaneswalkers()}})
	effectstest.RegisterForTest(t, effects.Spec{OracleID: testChampionOracle, Name: "Exiled Champion", Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.CantAttackItsOwner()}})
}

func TestCantAttackItsOwnerConstructors(t *testing.T) {
	registerOwnerRestrictionSpecs(t)
	g := effectstest.NewCatalogGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
	sleeper := effectstest.PushStolenForTest(g, owner.ID, me.ID, "Sleeper Agent", testSleeperOracle)
	champ := effectstest.PushStolenForTest(g, owner.ID, me.ID, "Exiled Champion", testChampionOracle)
	walker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: walker, Name: "Owner's Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: owner.ID, Controller: owner.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	sc := effectstest.FindBattlefieldCardForTest(g, sleeper).Effective()
	if len(sc.AttackTargetRestrictions) != 1 || !sc.AttackTargetRestrictions[0].NotOwnersPlaneswalkers ||
		sc.AttackTargetRestrictions[0].SourceName != "Sleeper Agent" {
		t.Errorf("sleeper restrictions = %+v", sc.AttackTargetRestrictions)
	}
	if len(sc.AttackRequirements) != 1 {
		t.Errorf("sleeper requirements = %+v, want attacks each combat", sc.AttackRequirements)
	}
	cc := effectstest.FindBattlefieldCardForTest(g, champ).Effective()
	if len(cc.AttackTargetRestrictions) != 1 || cc.AttackTargetRestrictions[0].NotOwnersPlaneswalkers {
		t.Errorf("champion restrictions = %+v, want the owner only", cc.AttackTargetRestrictions)
	}

	for _, id := range []uuid.UUID{sleeper, champ} {
		if err := g.Clone().DeclareAttacker(id, owner.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("declare %s at its owner: err = %v, want ErrIllegalAttackTarget", id, err)
		}
	}
	if err := g.Clone().DeclareAttacker(sleeper, walker); !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Errorf("sleeper at its owner's planeswalker: err = %v", err)
	}
	if err := g.Clone().DeclareAttacker(champ, walker); err != nil {
		t.Errorf("champion at its owner's planeswalker: %v", err)
	}
	if err := g.DeclareAttacker(sleeper, other.ID); err != nil {
		t.Errorf("sleeper at another opponent: %v", err)
	}
}
