package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brigid's Command — Kindred Sorcery — Kithkin {1}{G}{W} (Lorwyn
// Eclipsed, #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Kithkin you control.
//	 • Target player creates a 1/1 green and white Kithkin creature token.
//	 • Target creature you control gets +3/+3 until end of turn.
//	 • Target creature you control fights target creature an opponent
//	   controls."
//
// Ashling's Command's "choose two" template (ChooseN 2..2): each chosen
// bullet has its own targets, checked again as the spell resolves (CR
// 608.2b), and the bullets run in printed order (CR 608.2c), so a pumped
// creature fights with its +3/+3. The copy bullet is
// KindredCommandCopyMode; the Kithkin token is the printed tecl row,
// created under the target player's control; the fight bullet is Clash
// of the Eikons' two clauses (CR 701.14), fighting only while both are
// still legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6487fc1b-5c83-4d09-94d3-175725bcb922",
		Name:         "Brigid's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			KindredCommandCopyMode("Kithkin"),
			ModeWithPurpose(ModeDoing("Target player creates a 1/1 green and white Kithkin creature token.",
				TargetPlayer("target player"), brigidsCommandKithkin),
				ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			ModeWithPurpose(ModeDoing("Target creature you control gets +3/+3 until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				BoostTheModesTarget(3, 3, "Brigid's Command")),
				ForTargets(BenefitUntilEOT(0, 3, 3))),
			ModeWithPurpose(ModeDoing("Target creature you control fights target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				),
				FightTheModesTargets), game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}

// brigidsCommandKithkin is "target player creates a 1/1 green and white
// Kithkin creature token", under that player's control.
func brigidsCommandKithkin(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return CreateToken{Controller: t.ID, Template: TokenCard("1/1 green and white Kithkin"), N: 1}.Apply(ctx)
}
