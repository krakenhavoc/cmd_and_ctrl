package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Halana and Alena, Partners — Legendary Creature — Human Ranger
// {2}{R}{G}, 2/3 (Foundations, #4465):
//
//	"Reach
//	 First strike
//	 At the beginning of combat on your turn, put X +1/+1 counters on
//	 another target creature you control, where X is Halana and Alena's
//	 power. That creature gains haste until end of turn."
//
// Sickbay's beginning-of-combat trigger on another target creature you
// control (Another, CR 109.1), with Ouroboroid's X: Halana and Alena's
// power read once as the trigger resolves (CR 608.2h), counters included,
// or as they last existed on the battlefield if they have left. A power of
// 0 or less puts no counters, and the creature gains haste either way. A
// target that is gone as the trigger resolves gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b347be2d-9aa1-42a9-b652-8fb032832f09",
		Name:            "Halana and Alena, Partners",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(
				"Halana and Alena, Partners — put X +1/+1 counters on another target creature you control, where X is their power; it gains haste until end of turn",
				Do(
					Instruction(func(ctx *Context) error {
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						x, ok := ouroboroidPower(ctx.Game, ctx)
						if !ok || x <= 0 {
							return nil
						}
						return AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
					}),
					// The haste waits for a prompt the counters ask (CR 608.2c).
					Instruction(func(ctx *Context) error {
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"haste"},
							Label:    "Halana and Alena, Partners — haste until end of turn",
						}.Apply(ctx)
					}),
				)),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
