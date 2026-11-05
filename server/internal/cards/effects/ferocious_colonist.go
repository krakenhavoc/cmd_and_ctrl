package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferocious Colonist — Creature — Klingon Citizen {1}{R}, 0/1 (Star Trek
// Commander, #4974):
//
//	"Double strike
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of
//	 turn for each opponent you attacked this combat.)"
//
// Both keywords are the engine's: double strike (CR 702.4) and melee
// (CR 702.121, game/melee.go), whose trigger counts each opponent the
// controller attacked with a creature this combat as it resolves, and
// pumps this object until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd549938-8919-45ac-bf52-92492d9d1d49",
		Name:            "Ferocious Colonist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", game.KeywordMelee},
	})
}
