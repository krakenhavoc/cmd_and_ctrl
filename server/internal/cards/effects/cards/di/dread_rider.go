package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Rider — Creature — Spirit Knight {5}{B}, 3/7 (Jumpstart 2022,
// #5948):
//
//	"{1}{B}, {T}, Exile a creature card from your graveyard: Target
//	 opponent loses 3 life."
//
// Moorland Haunt's graveyard-exile cost (a creature card of your own
// graveyard, paid as it is activated) on Brush with Death's opponent
// target; the life loss is not a drain, so nothing is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5c15ef5-6795-40c2-a131-a73f74142314",
		Name:         "Dread Rider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Exile a creature card from your graveyard: Target opponent loses 3 life.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return effects.GainLife{Player: item.Targets[0].ID, Amount: -3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
