package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wretched Confluence — Instant {3}{B}{B} (Commander 2015, #5546):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Target player draws a card and loses 1 life.
//	 • Target creature gets -2/-2 until end of turn.
//	 • Return target creature card from your graveyard to your hand."
//
// Mystic Confluence's ChooseNRepeating (CR 700.2d): each occurrence runs
// once in printed order with its own target group, so two -2/-2 bullets may
// shrink two creatures or one creature twice. The first bullet is Shadrix
// Silverquill's draw-then-lose steps on the occurrence's player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb430640-81b6-4720-84d6-b8ec3af6a2c3",
		Name:         "Wretched Confluence",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeSteps("Target player draws a card and loses 1 life.",
				effects.TargetPlayer("target player"),
				effects.Instruction(effects.ShadrixDraw), effects.Instruction(effects.ShadrixLoseLife)),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-2 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-2, -2, "Wretched Confluence")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesTargetsToHand),
		),
	})
}
