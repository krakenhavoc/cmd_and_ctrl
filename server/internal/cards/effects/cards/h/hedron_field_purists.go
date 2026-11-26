package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hedron-Field Purists — {2}{W} Creature — Human Cleric 0/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {2}{W}
//	LEVEL 1-4: 1/4, If a source would deal damage to you or a creature you control, prevent 1 of that damage.
//	LEVEL 5+: 2/5, If a source would deal damage to you or a creature you control, prevent 2 of that damage.
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecc3e2ac-6534-4350-92f4-2c1f578d854e",
		Name:         "Hedron-Field Purists",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}{W}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 4, Power: 1, Toughness: 4},
			effects.LevelBand{From: 5, Power: 2, Toughness: 5},
		),
		Replacements: []game.ReplacementEffect{
			effects.OldB478PreventDamageToYouAndYourCreatures("Hedron-Field Purists", 1, 1, 4),
			effects.OldB478PreventDamageToYouAndYourCreatures("Hedron-Field Purists", 2, 5, 0),
		},
	})
}
