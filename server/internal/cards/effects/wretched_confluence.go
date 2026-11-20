package effects

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
	Register(Spec{
		OracleID:     "fb430640-81b6-4720-84d6-b8ec3af6a2c3",
		Name:         "Wretched Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeSteps("Target player draws a card and loses 1 life.",
				TargetPlayer("target player"),
				Instruction(shadrixDraw), Instruction(shadrixLoseLife)),
			ModeWithPurpose(ModeDoing("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-2, -2, "Wretched Confluence")), ForTargets(HarmUntilEOT(0, -2, -2))),
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				ReturnTheModesTargetsToHand),
		),
	})
}
