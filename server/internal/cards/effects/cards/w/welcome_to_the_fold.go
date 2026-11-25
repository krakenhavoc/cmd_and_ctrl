package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welcome to the Fold — Sorcery {2}{U}{U} (Shadows over Innistrad, S82 OLD):
//
//	"Madness {X}{U}{U}
//	 Gain control of target creature if its toughness is 2 or less. If this
//	 spell's madness cost was paid, instead gain control of that creature if
//	 its toughness is X or less."
//
// From Under the Floorboards' madness with a free X (the printed cost has no
// {X}, so CR 107.3b doesn't lock it): the branch reads which alternative cost
// the spell paid (CR 702.35b). Toughness is read through the layers as the
// spell resolves, so a creature pumped or shrunk in response is judged as it
// then is; the control change has no duration (New Blood's, CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8647f54-f50b-4bc6-bd44-fda25ef189fa",
		Name:         "Welcome to the Fold",
		Completeness: effects.CompletenessFull,
		Madness:      "{X}{U}{U}",
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			limit := 2
			if ctx.PaidAltCost(game.AltCostKeyMadness) {
				limit = ctx.X()
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.CurrentToughness() > limit {
					return nil
				}
				return effects.GainControl{
					Target:   t.ID,
					Duration: game.IndefiniteDuration(),
					Label:    "Welcome to the Fold — gain control",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
