package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garruk's Harbinger — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "2dbe384b-71c4-43f9-be35-ba2894021b85", Name: "Garruk's Harbinger", Completeness: CompletenessFull,
		PrintedKeywords: []string{game.HexproofFromColor("B")},
		Triggered: []game.TriggeredAbility{On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
				return false
			}
			if g.PlayerByIDForEffect(ev.Target) != nil {
				return true
			}
			c, ok := g.LookupCardForEffect(ev.Target)
			return ok && c.IsPlaneswalker()
		}, "Garruk's Harbinger — look at that many cards and take a creature or Garruk", func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return LookAtTopThenMayTakeToHand(item.Trigger.Event.Amount, Or(Creature(), And(Planeswalker(), HasSubtype("Garruk"))), 1, "Garruk's Harbinger — reveal a creature or Garruk planeswalker and put it into your hand")(g, item)
		})},
	})
}
