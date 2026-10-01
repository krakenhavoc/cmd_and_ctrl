package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_target_restrictions_test.go — ADR 0106 §2 (#1794), the #544
// half. A creature that "attacks each combat if able and can't attack
// its owner or planeswalkers its owner controls" is offered every other
// target and never its owner, every offered move is accepted, the
// withheld ones really are refused, and when only its owner is left to
// attack the pass comes back (CR 508.1d counts requirements "without
// disobeying any restrictions").

const legalSleeperOracle = "test-legal-sleeper-agent"

// withLegalSleeperStatics gives the test oracle Xantcha's two statics,
// written as the catalog constructors write them, on top of whatever
// the real catalog answers for every other card.
func withLegalSleeperStatics(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.IsCreature()
	}
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != legalSleeperOracle {
			if prev == nil {
				return nil
			}
			return prev(key)
		}
		return []game.StaticAbility{
			{Layer: game.Layer6Ability, AppliesTo: self, Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, s *game.Card) {
				c.AttackRequirements = append(c.AttackRequirements, game.AttackRequirement{Source: s.InstanceID, SourceName: s.Name})
			}},
			{Layer: game.Layer6Ability, AppliesTo: self, Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, s *game.Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
					Source: s.InstanceID, SourceName: s.Name, NotOwner: true, NotOwnersPlaneswalkers: true,
				})
			}},
		}
	}
}

// sleeperFor puts the test creature under `controller`, owned by
// `owner`, entered (so the layer pass reaches its statics) but not
// summoning sick.
func sleeperFor(g *game.Game, owner, controller *game.Player) uuid.UUID {
	id := enteredCard(g, controller, game.Card{
		Name: "Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		OracleID: legalSleeperOracle, Power: 5, Toughness: 5,
	})
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
			c.Owner = owner.ID
			c.SummonedThisTurn = false
		}
	}
	return id
}

func TestCantAttackOwnerMovesAgreeWithTheEngine(t *testing.T) {
	withLegalSleeperStatics(t)
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, owner := g.Seats[s], g.Seats[(s+1)%4]
	clearHand(me)
	sleeper := sleeperFor(g, owner, me)
	walker := battlefieldCard(g, owner, game.Card{
		Name: "Owner's Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if countKind(moves, legal.KindPass) != 0 {
		t.Fatalf("pass offered while the creature owes an attack: %v", labels(moves))
	}
	offered := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != sleeper {
			continue
		}
		ap := decodeAttack(t, m)
		if ap.Target == owner.ID.String() || ap.Target == walker.String() {
			t.Errorf("offered the creature at its owner or the owner's planeswalker: %q", m.Label)
		}
		if !m.AlwaysLegal {
			t.Errorf("an attack that answers the requirement is not AlwaysLegal: %q", m.Label)
		}
		offered++
	}
	if offered != 2 {
		t.Fatalf("offered %d attacks, want the two opponents who are not its owner: %v", offered, labels(moves))
	}
	for _, refused := range []uuid.UUID{owner.ID, walker} {
		if err := g.Clone().DeclareAttacker(sleeper, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}

func TestCantAttackOwnerWithOnlyItsOwnerLeftOwesNothing(t *testing.T) {
	withLegalSleeperStatics(t)
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, owner := g.Seats[s], g.Seats[(s+1)%4]
	clearHand(me)
	sleeper := sleeperFor(g, owner, me)
	for _, p := range []*game.Player{g.Seats[(s+2)%4], g.Seats[(s+3)%4]} {
		p.Eliminated = true
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if n := attacksBy(moves, sleeper); n != 0 {
		t.Errorf("offered %d attacks with only its owner to attack: %v", n, labels(moves))
	}
	if countKind(moves, legal.KindPass) != 1 {
		t.Errorf("pass withheld although the requirement can't be obeyed: %v", labels(moves))
	}
}
