package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aim for the Head — Sorcery {2}{B} (Innistrad: Crimson Vow, #5495):
//
//	"Choose one —
//	 • Exile target Zombie.
//	 • Target opponent exiles two cards from their hand."
//
// A real mode pick. The Zombie is read through the layers, so a changeling
// is a legal target. The second bullet is an edict the opponent answers
// (Perfect Intimidation's prompt): they choose which two cards, a hand of
// two or fewer goes whole, and an empty hand is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b191ae0c-7d54-459f-abe3-aab5d792e514",
		Name:         "Aim for the Head",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Exile target Zombie.",
				effects.Harmful(effects.TargetPermanent("target Zombie", effects.Subtype("Zombie"))),
				effects.ExileTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Target opponent exiles two cards from their hand.",
				effects.TargetPlayer("target opponent", effects.Opponent()), effects.VowB07OpponentExilesTwoFromHand),
		),
	})
}
