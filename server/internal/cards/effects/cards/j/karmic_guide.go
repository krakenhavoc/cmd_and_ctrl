package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karmic Guide — Creature — Angel Spirit, {3}{W}{W}, 2/2:
//
//	"Flying, protection from black
//	 Echo {3}{W}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, return target creature card from your graveyard to the battlefield."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c31fec9-e4b3-4761-990e-7be38eb05604",
		Name:            "Karmic Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Karmic Guide", "{3}{W}{W}"),
			effects.Targeting(effects.WhenThisEnters("Karmic Guide — return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
					return nil
				}), effects.TargetCreatureInYourGraveyard()),
		},
	})
}
