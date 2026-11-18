package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twinferno — Instant {1}{R} (Dominaria United):
//
//	"Choose one —
//	 • When you cast your next instant or sorcery spell this turn, copy
//	   that spell. You may choose new targets for the copy.
//	 • Target creature you control gains double strike until end of turn."
//
// The first bullet is Doublecast's delayed trigger (CR 603.7b, WhenYouNextCast
// with the copy-the-spell body): it waits for the next instant or sorcery its
// controller casts this turn and ends with the turn if none comes. Twinferno
// is itself an instant, but it is already cast when the trigger is made, so
// it does not copy itself. The second bullet is a keyword grant until end of
// turn on the one target (CR 611.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93989dd7-2d3e-46e2-8e92-8d0479796087",
		Name:         "Twinferno",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("When you cast your next instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return WhenYouNextCast(
						"Twinferno — copy that spell",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						copyTheSpellBody,
					).Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Target creature you control gains double strike until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				func(item *game.StackItem, ctx *Context, _ int) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   item.Targets[0].ID,
						Keywords: []string{"double strike"},
						Label:    "Twinferno — double strike",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike"))}),
		),
	})
}
