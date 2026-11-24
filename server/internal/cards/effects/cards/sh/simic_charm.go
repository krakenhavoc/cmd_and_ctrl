package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Simic Charm — Instant {G}{U} (Commander Legends):
//
//	"Choose one —
//	 • Target creature gets +3/+3 until end of turn.
//	 • Permanents you control gain hexproof until end of turn.
//	 • Return target creature to its owner's hand."
//
// Izzet Charm's shape. The hexproof goes to the permanents you control as the
// spell resolves, locked then (CR 611.2c), so one that enters later does not
// have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c767aa2e-0a95-401f-8e36-d90a87648eab",
		Name:         "Simic Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target creature gets +3/+3 until end of turn.", effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
			effects.Mode("Permanents you control gain hexproof until end of turn."),
			effects.Mode("Return target creature to its owner's hand.", effects.TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Simic Charm — +3/+3 until end of turn"}.Apply(ctx)
					}
				}
			case ctx.HasMode(1):
				return effects.GrantKeywordUntilEOT{
					Match:    effects.YouControl(),
					Keywords: []string{"hexproof"},
					Label:    "Simic Charm — permanents you control gain hexproof",
				}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.BounceEachLegalTarget(ctx.Game, item)
			}
			return nil
		},
	})
}
