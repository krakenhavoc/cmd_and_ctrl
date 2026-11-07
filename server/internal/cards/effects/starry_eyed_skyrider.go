package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starry-Eyed Skyrider — Creature — Human Scout {2}{W}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Whenever this creature attacks, another target creature you control
//	 gains flying until end of turn.
//	 Attacking tokens you control have flying."
//
// Flying is a printed keyword. The attack trigger is Starling, Aerial
// Ally's grant behind an attack condition: another creature you control,
// chosen as the trigger goes on the stack and checked again as it
// resolves. The static is Bone-Cairn Butcher's attacking-only layer 6
// grant with flying: a token has flying only while it is attacking for
// you, so a mobilized Warrior flies and a token at home does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04e615e6-4088-4170-904c-f2d4bcebed59",
		Name:            "Starry-Eyed Skyrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Starry-Eyed Skyrider — another target creature you control gains flying until end of turn",
				ebTargetCreatureGainsUntilEOT("flying", "Starry-Eyed Skyrider — flying until end of turn")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			Label:                    "Attacking tokens you control have flying.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return IsToken(*target) && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			},
		}},
	})
}
