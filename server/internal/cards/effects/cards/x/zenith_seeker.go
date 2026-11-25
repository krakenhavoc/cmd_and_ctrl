package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zenith Seeker — {3}{U} Creature — Bird Wizard 2/2 (Amonkhet, #5535):
//
//	"Flying
//	 Whenever you cycle or discard a card, target creature gains flying until
//	 end of turn."
//
// Cycling a card is discarding it (CR 702.29a), so one discard event covers
// both words and the ability triggers once per card. The target is
// mandatory: with no creature on the battlefield the trigger is removed
// from the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0fdc716-31b9-47f6-8639-de8bb7e3629f",
		Name:            "Zenith Seeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.DiscardedByYou(ev, source)
				}, "Zenith Seeker — target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return effects.GrantKeywordUntilEOT{
								Target:   t.ID,
								Keywords: []string{"flying"},
								Label:    "Zenith Seeker — flying until end of turn",
							}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
