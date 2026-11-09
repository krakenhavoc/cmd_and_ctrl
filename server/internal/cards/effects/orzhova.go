package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orzhova — Plane — Ravnica (Planechase Anthology, #4870):
//
//	"When you planeswalk away from Orzhova, each player returns all
//	 creature cards from their graveyard to the battlefield.
//	 Whenever chaos ensues, for each opponent, exile up to one target
//	 creature card from that player's graveyard."
//
// A plane (ADR 0216 §6). The planeswalk-away trigger returns every
// player's creature cards together, each under its owner's control. The
// chaos ability is one "up to one" clause per possible opponent of the
// planar controller, each bound to that opponent's graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7241e4b-d2c7-4801-96a3-91bb998953a1",
		Name:         "Orzhova",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkAwayFromThis("Orzhova — each player returns all creature cards from their graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated {
							ids = append(ids, graveyardCardIDs(ctx, p.ID, func(c game.Card) bool { return c.IsCreature() })...)
						}
					}
					return ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
				}),
			Targeting(WheneverChaosEnsues("Orzhova — for each opponent, exile up to one target creature card from that player's graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				b9PerOpponentGraveyardClauses("creature", func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return isCreatureCard(c) })),
		},
	})
}
