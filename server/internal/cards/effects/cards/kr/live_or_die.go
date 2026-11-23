package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Live or Die — Instant {3}{B}{B} (Duskmourn: House of Horror, #4466):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to the battlefield.
//	 • Destroy target creature."
//
// A modal spell (CR 700.2): the mode and its target are chosen as it is
// cast, and the chosen bullet reads its own target (ModeTarget). The two
// bullets are Inscription of Ruin's, without its mana value caps: the
// creature card returns under its owner's control, which is yours, and a
// target gone by resolution leaves the spell doing nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc367b23-d1f4-4c5c-b300-a44a726e3f74",
		Name:         "Live or Die",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to the battlefield.",
				effects.TargetCreatureInYourGraveyard(), effects.ReturnTheModesTargetToBattlefield),
			effects.ModeDoing("Destroy target creature.", effects.TargetCreature("target creature"), effects.DestroyTheModesTarget),
		),
	})
}
