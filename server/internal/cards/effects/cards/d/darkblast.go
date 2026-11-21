package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkblast — Instant {B} (Ravnica: City of Guilds, #5600):
//
//	"Target creature gets -1/-1 until end of turn.
//	 Dredge 3 (If you would draw a card, you may mill three cards instead.
//	 If you do, return this card from your graveyard to your hand.)"
//
// Locust Spray's shrink with Dakmor Salvage's dredge row: Dredge(3) works
// from the graveyard (CR 702.52a) and needs three cards in the library
// (CR 702.52b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64446130-46d0-4129-a718-ed6a04218100",
		Name:         "Darkblast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.HarmUntilEOT(0, -1, -1)), Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(-1, -1, "Darkblast — -1/-1"),
		Replacements: []game.ReplacementEffect{effects.Dredge(3)},
	})
}
