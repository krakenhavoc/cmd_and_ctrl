package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kargan Dragonlord — {R}{R} Creature — Human Warrior 2/2 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {R}
//	LEVEL 4-7: 4/4, Flying
//	LEVEL 8+: 8/8, Flying, trample, {R}: This creature gets +1/+0 until end of turn.
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.
//
// The pump is a granted activated ability held by the 8+ band.
const karganDragonlordGrant = "kargan-dragonlord/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb4e6d97-f5df-4290-b193-912f44f43a67",
		Name:         "Kargan Dragonlord",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: karganDragonlordGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Kargan Dragonlord — +1/+0"),
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{R}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 4, To: 7, Power: 4, Toughness: 4, Keywords: []string{"flying"}},
			effects.LevelBand{From: 8, Power: 8, Toughness: 8, Keywords: []string{"flying", "trample"}, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, karganDragonlordGrant),
					`This creature has "{R}: This creature gets +1/+0 until end of turn."`),
			}},
		),
	})
}
