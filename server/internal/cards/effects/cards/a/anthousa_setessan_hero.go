package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anthousa, Setessan Hero — Legendary Creature — Human Warrior {3}{G}{G}, 4/5
// (Theros, S82 OLD b388):
//
//	"Heroic — Whenever you cast a spell that targets Anthousa, up to three target
//	 lands you control each become 2/2 Warrior creatures until end of turn.
//	 They're still lands."
//
// The lands are chosen as the trigger goes on the stack (CR 603.3d), "up to
// three" so none is allowed. Each still-legal target becomes a 2/2 Warrior
// creature in addition to being a land, until end of turn (CR 611.2c: the set is
// fixed as it resolves). A land that came under your control this turn is
// summoning sick as a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "141f952b-9e9f-48c5-9c88-b1504dfa73bc",
		Name:         "Anthousa, Setessan Hero",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Heroic("Anthousa, Setessan Hero — up to three target lands you control each become 2/2 Warrior creatures until end of turn (heroic)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						var lands []uuid.UUID
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								lands = append(lands, t.ID)
							}
						}
						for _, id := range lands {
							if err := (effects.BecomeCreatureUntilEOT{
								Target:       id,
								Types:        []string{"Creature"},
								Subtypes:     []string{"Warrior"},
								SetPower:     2,
								SetToughness: 2,
								Label:        "Anthousa, Setessan Hero — becomes a 2/2 Warrior creature",
							}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetPermanent("up to three target lands you control", effects.Land(), effects.YouControl()).WithCount(0, 3)),
		},
	})
}
