package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Princess Irulan Corrino — Legendary Creature — Human Noble Advisor
// {1}{W}{U}{B}, 3/3 (Secret Lair Drop):
//
//	"Whenever you draw your second card each turn, you gain 2 life.
//	 Whenever an opponent draws their second card each turn, they lose 2 life.
//	 {T}: You and target opponent each draw two cards."
//
// The two triggers read the per-turn draw tally as each card is drawn (one
// event per card, so a two-card draw triggers on its second card). "They" is
// the opponent who drew, read off the triggering event. The activated ability
// draws for you first, then for the target opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85f20d9d-588a-4b18-8cc3-3796150630ba",
		Name:         "Princess Irulan Corrino",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Princess Irulan Corrino — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Princess Irulan Corrino — that player loses 2 life", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You and target opponent each draw two cards.",
			Purpose: game.Purpose{Draws: 2},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
