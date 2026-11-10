package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unwelcome Sprite — Creature — Faerie Rogue {1}{U}, 2/1 (Lorwyn
// Eclipsed, #4458):
//
//	"Flying
//	 Whenever you cast a spell during an opponent's turn, surveil 2."
//
// Katara, Waterbending Master's cast condition: the caster is the Sprite's
// controller and is not the active player. Each such cast is its own
// trigger; the surveil is the footprinted primitive (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0a38e8c-c42c-4c99-8a44-d390ff2511ff",
		Name:            "Unwelcome Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, stdB64YouCastDuringAnOpponentsTurn,
				"Unwelcome Sprite — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
