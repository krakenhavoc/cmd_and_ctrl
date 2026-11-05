package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Cerritos, Following Up — Legendary Artifact — Spacecraft {5},
// 4/4 (Star Trek, #4973):
//
//	"When U.S.S. Cerritos enters, create two 1/1 red Officer creature
//	 tokens.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Flying"
//
// U.S.S. Enterprise-A's shape (an Officer-making enters trigger on
// Station, SpacecraftAt and ThresholdKeywords) with two Officers, a 4/4
// body and flying alone at 6+.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "472eac8f-2b82-4fa9-a40c-af11ec40f19e",
		Name:         "U.S.S. Cerritos, Following Up",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("U.S.S. Cerritos — create two 1/1 red Officer creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 2})),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 4, 4),
			ThresholdKeywords(6, "flying"),
		),
	})
}
