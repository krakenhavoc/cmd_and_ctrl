package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scroll of Griselbrand — {1} Artifact: "{1}, Sacrifice this artifact: Target
// opponent discards a card. If you control a Demon, that player loses 3
// life."
//
// The discard is the opponent's choice. The Demon is looked for as the
// ability resolves, the Scroll having been sacrificed as a cost.
//
// No simplification.
func init() {
	const label = "{1}, Sacrifice this artifact: Target opponent discards a card. If you control a Demon, that player loses 3 life."
	effects.Register(effects.Spec{
		OracleID:     "c961373a-be0e-467c-92e0-33a0c953736a",
		Name:         "Scroll of Griselbrand",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Scroll of Griselbrand — discard a card",
					})
					if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.OfCreatureType("Demon"))) > 0 {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -3)
					}
				}
				return nil
			},
		}},
	})
}
