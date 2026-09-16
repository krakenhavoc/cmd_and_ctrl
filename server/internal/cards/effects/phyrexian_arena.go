package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Arena — Enchantment for {1}{B}{B}:
//
//	"At the beginning of your upkeep, you lose 1 life and draw a
//	card."
//
// S19 sub-PR 5: the first "your upkeep" trigger on the auto-fire
// pipeline. EventBeginUpkeep carries the active player in Actor;
// AppliesTo gates on Actor == Controller so the Arena only fires on
// its own controller's upkeep (not every player's). Mandatory: the
// trigger goes on the stack at upkeep and the life loss + draw
// happen when it resolves.
func init() {
	Register(Spec{
		OracleID: "ee579a32-a048-4335-b966-231ba731cdea",
		Name:     "Phyrexian Arena",
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Phyrexian Arena — lose 1 life, draw a card", func(g *game.Game, item *game.StackItem) error {
				if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
