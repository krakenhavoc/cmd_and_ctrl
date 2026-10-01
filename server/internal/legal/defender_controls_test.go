package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// defender_controls_test.go — ADR 0107 §2 (#1879), the #544 half. A
// creature that "can't attack unless defending player controls an Island"
// is offered the opponent with an Island and that opponent's planeswalker,
// never the others, every offered move is accepted, and the withheld ones
// really are refused.

const legalSerpentOracle = "test-legal-island-serpent"

func withLegalSerpentStatic(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != legalSerpentOracle {
			if prev == nil {
				return nil
			}
			return prev(key)
		}
		return []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, s *game.Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
					Source: s.InstanceID, SourceName: s.Name,
					DefenderMustControl: []game.PermanentQuery{{Subtypes: []string{"Island"}}},
				})
			},
		}}
	}
}

func TestDefenderMustControlMovesAgreeWithTheEngine(t *testing.T) {
	withLegalSerpentStatic(t)
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, islands, dry := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+2)%4]
	clearHand(me)
	serpent := enteredCard(g, me, game.Card{
		Name: "Sea Serpent", TypeLine: "Creature — Serpent", OracleID: legalSerpentOracle, Power: 5, Toughness: 5,
	})
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == serpent {
			c.SummonedThisTurn = false
		}
	}
	battlefieldCard(g, islands, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	islandsWalker := battlefieldCard(g, islands, game.Card{
		Name: "Islands' Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	dryWalker := battlefieldCard(g, dry, game.Card{
		Name: "Dry Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	offered := map[string]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != serpent {
			continue
		}
		offered[decodeAttack(t, m).Target] = true
	}
	if len(offered) != 2 || !offered[islands.ID.String()] || !offered[islandsWalker.String()] {
		t.Fatalf("offered %v, want the Island player and their planeswalker: %v", offered, labels(moves))
	}
	for _, refused := range []uuid.UUID{dry.ID, dryWalker, g.Seats[(s+3)%4].ID} {
		if err := g.Clone().DeclareAttacker(serpent, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}
