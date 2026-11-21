package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gobhobbler Rats — Creature — Rat {B}{R}, 2/2 (Ravnica Remastered, #5948):
//
//	"Hellbent — As long as you have no cards in hand, this creature gets
//	 +1/+0 and has '{B}: Regenerate this creature.'"
//
// Two statics that watch the hand (DependsOnHandSize, so the layers are
// rebuilt as cards enter and leave it): a layer 7c +1/+0 and a layer 6
// grant of the regeneration ability, from the bundle below, to this creature
// only. Both stop the moment its controller holds a card.
//
// No simplification.
func init() {
	const grant = "gobhobbler-rats/regenerate"
	pump := gobhobblerRatsPump(1, 0)
	effects.Register(effects.Spec{
		OracleID:     "aa254e74-583d-4681-8169-5ac0dbdb10e5",
		Name:         "Gobhobbler Rats",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{B}: Regenerate this creature.",
		}},
		Static: []game.StaticAbility{
			gobhobblerRatsHellbent(pump, "Hellbent — as long as you have no cards in hand, this creature gets +1/+0."),
			gobhobblerRatsHellbent(effects.GrantAbilities(effects.SelfOnly, grant), `Hellbent — as long as you have no cards in hand, this creature has "{B}: Regenerate this creature."`),
		},
	})
}

// gobhobblerRatsPump is the +P/+T of a self-pump: a layer 7c modify.
func gobhobblerRatsPump(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    "Hellbent — as long as you have no cards in hand, this creature gets +1/+0.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// gobhobblerRatsHellbent holds a static to the controller having an empty hand.
func gobhobblerRatsHellbent(s game.StaticAbility, label string) game.StaticAbility {
	s.Label = label
	when := s.AppliesTo
	s.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return when(target, g, source) && effects.R35YouHaveNoCardsInHand(g, source.Controller, uuid.Nil)
	}
	s.DependsOnHandSize = true
	return s
}
