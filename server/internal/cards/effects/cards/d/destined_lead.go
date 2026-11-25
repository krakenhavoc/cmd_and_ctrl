package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Destined // Lead — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Destined {1}{B}: "Target creature gets +1/+0 and gains indestructible until
//	 end of turn."
//	Lead {3}{G}: "Aftermath. All creatures able to block target creature this
//	 turn do so."
//
// Two registered faces. Destined's pump and keyword share one target, both
// until end of turn. Lead is Bloodscent's lure on the targeted creature, read
// as each blocker is declared (CR 509.1c); it is cast only from a graveyard
// and exiled as it resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ebde396-6672-491a-a6ce-1de49b12379b",
		Name:         "Destined",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible")), Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Destined — +1/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Destined — indestructible"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "7ebde396-6672-491a-a6ce-1de49b12379b#1",
		Name:         "Lead",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.LureTargetForTheTurn(ctx, "Lead — all creatures able to block it do so")
		},
	})
}
