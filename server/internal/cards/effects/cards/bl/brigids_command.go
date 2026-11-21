package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6487fc1b-5c83-4d09-94d3-175725bcb922",
		Name:         "Brigid's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.KindredCommandCopyMode("Kithkin"),
			effects.ModeWithPurpose(effects.ModeDoing("Target player creates a 1/1 green and white Kithkin creature token.",
				effects.TargetPlayer("target player"), brigidsCommandKithkin),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature you control gets +3/+3 until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.BoostTheModesTarget(3, 3, "Brigid's Command")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature you control fights target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				),
				effects.FightTheModesTargets), game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}

// brigidsCommandKithkin is "target player creates a 1/1 green and white
// Kithkin creature token", under that player's control.
func brigidsCommandKithkin(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.CreateToken{Controller: t.ID, Template: effects.TokenCard("1/1 green and white Kithkin"), N: 1}.Apply(ctx)
}
