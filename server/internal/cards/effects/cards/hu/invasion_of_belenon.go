package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Belenon // Belenon War Anthem — Battle — Siege {2}{W}, defense
// 5, and its back face, an Enchantment (March of the Machine):
//
//	"When this Siege enters, create a 2/2 white and blue Knight creature
//	 token with vigilance."
//	back: "Creatures you control get +1/+1."
//
// The protector prompt and defense counters are engine behaviour keyed on
// the Battle type; the back face registers under the oracle ID plus "#1".
func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfBelenonOracleID,
		Name:         "Invasion of Belenon",
		Completeness: effects.CompletenessFull,
		Battle: &effects.BattleSpec{
			Defense: 5,
			Subtype: effects.BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Belenon — create a 2/2 white and blue Knight with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Knight with vigilance"), N: 1})),
			effects.DefeatedTrigger("Invasion of Belenon — defeated: exile it, then cast Belenon War Anthem", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     InvasionOfBelenonOracleID + "#1",
		Name:         "Belenon War Anthem",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.B16CreaturesYouControl, 1, 1),
		},
	})
}

const InvasionOfBelenonOracleID = "e462e78d-2d26-4a2d-8d8f-5f31abe5c924"
