package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Outfitter — Creature — Kor Soldier {W}{W}, 2/2 (Zendikar, S82 OLD b494):
//
//	"When this creature enters, you may attach target Equipment you control to
//	 target creature you control."
//
// Noctis, Heir Apparent's two-slot attach clause (TwoSlotAttachTargets,
// AttachClauseTargets) on an enters trigger: both targets are chosen as the
// trigger goes on the stack and each is checked again as it resolves; an
// Equipment that can't be attached to that creature stays where it was.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ae3562f-28b7-4462-96ed-be0cf7052ccc",
		Name:         "Kor Outfitter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Kor Outfitter — you may attach target Equipment you control to target creature you control", effects.AttachClauseTargets),
				effects.TwoSlotAttachTargets(
					"target Equipment you control", []effects.CardPredicate{effects.YouControl()},
					"target creature you control", []effects.CardPredicate{effects.YouControl()},
				)),
				"Kor Outfitter — attach target Equipment you control to target creature you control?"),
		},
	})
}
