package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Transfer — Sorcery {1}{B}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"Choose one. If you control an artifact and an enchantment as you cast
//	 this spell, you may choose both instead.
//	 • Exile target creature or planeswalker.
//	 • Return target creature or planeswalker card from your graveyard to your
//	   hand."
//
// The mode count is Molten Collapse's conditional one (OrUpToIf(2, …)), read
// for the caster as the modes are chosen (CR 601.2b) and fixed from then on.
// With both chosen they resolve in printed order (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e46c709-5278-44f3-9f0e-f71bd9558339",
		Name:         "Soul Transfer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeDoing("Exile target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				ExileTheModesTarget),
			ModeDoing("Return target creature or planeswalker card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature or planeswalker card from your graveyard", YouOwn(), Or(Creature(), Planeswalker())),
				neoB08ReturnTheModesTargetToHand),
		).OrUpToIf(2, neoB08ModeCondition),
	})
}
