package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_target_restrictions_test.go — the bot half of ADR 0106 §2
// (#1794). A creature on a bot's side that "attacks each combat if
// able and can't attack its owner or planeswalkers its owner controls"
// attacks, never attacks its owner — even when its owner is the most
// tempting target on the table, at 5 life with no blockers — and the
// combat ends. The bot's whole world is the enumerator, which never
// offers the owner, so there is nothing for any policy to get wrong;
// this pins that no policy finds a way round it.

const botSleeperOracle = "test-bot-sleeper-agent"

func withBotSleeperStatics(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.IsCreature()
	}
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != botSleeperOracle {
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

func TestBotsNeverAttackTheOwnerOfACreatureThatCantAttackItsOwner(t *testing.T) {
	withBotSleeperStatics(t)
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			s := g.Turn.ActiveSeat
			active, owner := g.Seats[s], g.Seats[(s+1)%4]
			// The owner is the obvious target: 5 life, nothing to block
			// with. Every other opponent keeps its two 2/2s.
			kept := g.Battlefield.Cards[:0]
			for _, c := range g.Battlefield.Cards {
				if c.Controller != owner.ID {
					kept = append(kept, c)
				}
			}
			g.Battlefield.Cards = kept
			owner.Life = 5

			sleeper := limitPush(g, active, "Sleeper Agent", "Legendary Creature — Phyrexian Minion", botSleeperOracle, 5, 5)
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: sleeper, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.InstanceID == sleeper {
						c.Owner = owner.ID
						c.SummonedThisTurn = false
					}
				}
			})

			attackedAt := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.InstanceID == sleeper && c.AttackingTarget != uuid.Nil {
						attackedAt[c.AttackingTarget] = true
					}
				}
			})
			if len(attackedAt) == 0 {
				t.Fatal("the creature that attacks each combat never attacked")
			}
			if attackedAt[owner.ID] {
				t.Errorf("the creature attacked its owner: %v", attackedAt)
			}
		})
	}
}
