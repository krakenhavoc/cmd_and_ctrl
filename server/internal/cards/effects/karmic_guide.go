package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karmic Guide — Creature — Angel Spirit, {3}{W}{W}, 2/2:
//
//	"Flying, protection from black
//	 Echo {3}{W}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, return target creature card from your graveyard to the battlefield."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c31fec9-e4b3-4761-990e-7be38eb05604",
		Name:            "Karmic Guide",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
		Triggered: []game.TriggeredAbility{
			Echo("Karmic Guide", "{3}{W}{W}"),
			Targeting(WhenThisEnters("Karmic Guide — return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					reanimateSingleTarget(NewContext(g, item), item.Controller)
					return nil
				}), targetCreatureInYourGraveyard()),
		},
	})
}
