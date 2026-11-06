package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bone-Cairn Butcher — Creature — Demon {1}{R}{W}{B}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Mobilize 2 (Whenever this creature attacks, create two tapped and
//	 attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	 beginning of the next end step.)
//	 Attacking tokens you control have deathtouch."
//
// Mobilize is mobilize.go. The static is Ohran Frostfang's attacking-only
// layer 6 grant narrowed to tokens: it reads attacking status, so the
// Warriors (and any other token attacking for you) have deathtouch only
// while they are attacking, and lose it when combat ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "449db6bf-d766-4d09-a7bd-9c45b3d4c4c3",
		Name:         "Bone-Cairn Butcher",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mobilize("Bone-Cairn Butcher", 2)},
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			Label:                    "Attacking tokens you control have deathtouch.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return IsToken(*target) && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "deathtouch")
			},
		}},
	})
}
