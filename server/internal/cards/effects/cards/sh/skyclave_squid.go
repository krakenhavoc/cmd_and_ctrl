package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyclave Squid — Creature — Squid {1}{U}, 3/2 (Zendikar Rising, #5508):
//
//	"Defender
//	 Landfall — Whenever a land you control enters, this creature can attack
//	 this turn as though it didn't have defender."
//
// Assault Formation's one-turn defender waiver (CR 702.3b, read at the
// declaration of attackers), given by each land that enters under its
// controller's control. A Squid that has left the battlefield by the time the
// trigger resolves is not given anything.
//
// No simplification.
func init() {
	const label = "Skyclave Squid — can attack this turn as though it didn't have defender"
	effects.Register(effects.Spec{
		OracleID:        "301d36e7-06ff-48ed-a7e8-c7d3c72e1e2f",
		Name:            "Skyclave Squid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Skyclave Squid — landfall: can attack this turn as though it didn't have defender",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					return effects.WaiveDefenderUntilEOT(effects.NewContext(g, item), item.SourceCardID, label)
				}),
		},
	})
}
