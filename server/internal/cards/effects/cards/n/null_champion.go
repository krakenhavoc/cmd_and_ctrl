package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Null Champion — {1}{B} Creature — Zombie Warrior 1/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {3}
//	LEVEL 1-3: 4/2
//	LEVEL 4+: 7/3, {B}: Regenerate this creature.
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.
//
// The regeneration is a granted activated ability held by the 4+ band.
const nullChampionGrant = "null-champion/regenerate"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfd2846f-574b-475c-bf2a-dc3e87d13476",
		Name:         "Null Champion",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       nullChampionGrant,
			Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
			Text:      "{B}: Regenerate this creature.",
		}},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{3}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 3, Power: 4, Toughness: 2},
			effects.LevelBand{From: 4, Power: 7, Toughness: 3, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, nullChampionGrant),
					`This creature has "{B}: Regenerate this creature."`),
			}},
		),
	})
}
