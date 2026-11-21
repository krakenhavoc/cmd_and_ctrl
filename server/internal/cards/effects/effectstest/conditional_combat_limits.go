package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_combat_limits_test.go — #1534's proof cards (ADR 0045
// amendment of 2026-09-24, Decision 47). Mirri, Weatherlight Duelist
// carries the per-defender block limit (her attack trigger) and the
// While-gated attack limit (her static); The Eternal Wanderer carries
// the per-permanent attack scope and three loyalty abilities.
//
// Every shape is pinned the way combat_limits_test.go pins #1507's:
// the verb accepts the legal declaration and refuses the illegal one,
// and the enumerator offers exactly what the verb accepts — checked by
// trying every candidate on a clone.

const (
	MirriWeatherlightDuelistOracle = "60bc9cd2-2e03-4889-b1ef-fdab9a9a6d08"
	theEternalWandererOracle       = "20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e"
)

// clAcceptedAttacksAnywhere is clAcceptedAttacks over every attack
// target the engine lists — players AND planeswalkers — so a limit on
// a planeswalker is checked against the verb too.
func clAcceptedAttacksAnywhere(g *game.Game, seat uuid.UUID) []clAttack {
	var out []clAttack
	targets := g.AttackTargetsForEffect(seat)
	for _, c := range g.Battlefield.Cards {
		if c.Controller != seat || !c.IsCreature() || c.AttackingTarget != uuid.Nil {
			continue
		}
		for _, t := range targets {
			if err := g.Clone().DeclareAttacker(c.InstanceID, t.ID); err == nil {
				out = append(out, clAttack{c.InstanceID, t.ID})
			}
		}
	}
	return clSorted(out)
}

// clAgreeOnAttacksAnywhere asserts the enumerator offers exactly what
// the verb accepts at every target, and returns the offer.
func ClAgreeOnAttacksAnywhere(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	offered, accepted := clOfferedAttacks(t, g, seat), clAcceptedAttacksAnywhere(g, seat)
	if len(offered) != len(accepted) {
		t.Fatalf("the enumerator offers %d attacks, the verb accepts %d:\noffered  %v\naccepted %v", len(offered), len(accepted), offered, accepted)
	}
	for i := range offered {
		if offered[i] != accepted[i] {
			t.Fatalf("the enumerator and the verb disagree:\noffered  %v\naccepted %v", offered, accepted)
		}
	}
	return offered
}

// --- The Eternal Wanderer ------------------------------------------------

func PushWanderer(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := PushWalkerForTest(g, owner, "The Eternal Wanderer", theEternalWandererOracle, 5)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].TypeLine = "Legendary Planeswalker"
		}
	}
	return id
}
