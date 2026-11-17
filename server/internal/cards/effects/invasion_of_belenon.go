package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     invasionOfBelenonOracleID,
		Name:         "Invasion of Belenon",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Belenon — create a 2/2 white and blue Knight with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Knight with vigilance"), N: 1})),
			DefeatedTrigger("Invasion of Belenon — defeated: exile it, then cast Belenon War Anthem", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:     invasionOfBelenonOracleID + "#1",
		Name:         "Belenon War Anthem",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(b16CreaturesYouControl, 1, 1),
		},
	})
}

const invasionOfBelenonOracleID = "e462e78d-2d26-4a2d-8d8f-5f31abe5c924"
