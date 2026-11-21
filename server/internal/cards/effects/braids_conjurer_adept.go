package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Braids, Conjurer Adept — Legendary Creature — Human Wizard {2}{U}{U}, 2/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"At the beginning of each player's upkeep, that player may put an
//	 artifact, creature, or land card from their hand onto the battlefield."
//
// Gate to the Aether's each-upkeep trigger (the event's actor is the
// upkeep's player), with the put-from-hand pick made by, and the permanent
// entering under the control of, that player rather than Braids' controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "359b8f2c-831e-47d0-9267-dae79cecb63a",
		Name:         "Braids, Conjurer Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Braids, Conjurer Adept — that player may put an artifact, creature, or land card from their hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					player := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					return PutFromHandOntoBattlefield{
						Player:   player,
						Match:    Or(Artifact(), Creature(), Land()),
						Optional: true,
						Label:    "Braids, Conjurer Adept — you may put an artifact, creature, or land card from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
