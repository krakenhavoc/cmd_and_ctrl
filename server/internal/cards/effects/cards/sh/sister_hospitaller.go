package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sister Hospitaller — Creature — Human Cleric {4}{W}{B}, 3/2:
//
//	"Medicus Ministorum — When this creature enters, return target creature
//	 card from your graveyard to the battlefield. You gain life equal to its
//	 mana value."
//
// The life is read off the card as it sat in the graveyard (CR 608.2): the
// mana value of the card, with {X} as 0 (CR 202.3e). With no legal target the
// ability does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06c39c1c-0e34-4ef0-bd06-23245cb9cc59",
		Name:         "Sister Hospitaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sister Hospitaller — return target creature card from your graveyard to the battlefield and gain life equal to its mana value",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						card, ok := effects.ReanimateSingleTarget(ctx, item.Controller)
						if !ok {
							return nil
						}
						if mv := card.ManaValue(); mv > 0 {
							return effects.GainLife{Player: item.Controller, Amount: mv}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCreatureInYourGraveyard()),
		},
	})
}
