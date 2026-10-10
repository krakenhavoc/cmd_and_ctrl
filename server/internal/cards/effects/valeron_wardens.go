package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valeron Wardens — Creature — Human Monk {2}{G}, 1/3:
//
//	"Renown 2 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put two +1/+1 counters on it and it becomes
//	 renowned.)
//	 Whenever a creature you control becomes renowned, draw a card."
//
// #2049: renown is the engine's keyword trigger (game/renown.go); the
// draw watches EventBecameRenowned, which fires only for the renown
// trigger that actually made a creature renowned — the Wardens' own
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d1202573-2105-4d4d-ae7d-7164974b6d89",
		Name:            "Valeron Wardens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 2"},
		Triggered: []game.TriggeredAbility{
			WhenACreatureYouControlBecomesRenowned("Valeron Wardens — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
