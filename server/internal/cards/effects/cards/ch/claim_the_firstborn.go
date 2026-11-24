package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Claim the Firstborn — Sorcery {R} (Secret Lair Drop):
//
//	"Gain control of target creature with mana value 3 or less until end of
//	 turn. Untap that creature. It gains haste until end of turn."
//
// Captivating Crew's steal, untap and haste in printed order, behind a mana
// value 3 or less target clause (judged at announce and again as it resolves,
// CR 608.2b). A creature that is no longer a legal target by resolution makes the
// whole spell do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bb82829-05e8-4cdd-91e4-b837a620bfc8",
		Name:         "Claim the Firstborn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with mana value 3 or less", effects.ManaValueLE(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Claim the Firstborn — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Claim the Firstborn — haste until end of turn",
			}.Apply(ctx)
		},
	})
}
