package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Javelin of Lightning — Artifact — Equipment {1}{R} (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 During your turn, equipped creature gets +2/+0 and has first strike.
//	 Equip {4}"
//
// Knife's two turn-gated statics at +2/+0: the pump and the keyword are each
// re-read on every recompute against its controller's turn, so they come and
// go with the turn (CR 613). The enters trigger attaches the Equipment to the
// creature chosen as it goes on the stack; with none to choose, it stays
// unattached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3f244320-3e74-4eff-ba4e-a345700c40c9",
		Name:            "Javelin of Lightning",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			{
				Label:    "During your turn, equipped creature gets +2/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && effects.DuringSourceControllersTurn(target, g, source)
				},
				Apply: effects.PumpBy(2, 0),
			},
			effects.GrantToAttachedWhile(effects.DuringSourceControllersTurn, "first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Javelin of Lightning — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
