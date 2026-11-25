package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tazeem Raptor — Creature — Bird {2}{W}, 2/2 (Zendikar Rising, #5508):
//
//	"Flying
//	 When this creature enters, you may return a land you control to its
//	 owner's hand."
//
// ReturnOneYouControl's pick with a floor of zero: the controller may pick
// none, and the Raptor offers every land they control, tapped or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3111f8b7-36af-4390-af99-5b8976b2f241",
		Name:            "Tazeem Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tazeem Raptor — you may return a land you control to its owner's hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ChoosePermanents{
						Question:   "Tazeem Raptor — you may return a land you control to its owner's hand",
						Candidates: effects.OldB093OptionalLandYouControl,
						Keyed:      effects.BouncedChosen,
					}.Apply(ctx)
				}))),
		},
	})
}
