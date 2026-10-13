package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// afflictDefenderLoses is the body of afflict N (CR 702.130a): "Whenever
// this creature becomes blocked, defending player loses N life." The
// defending player is the becomes-blocked event's Actor, and the loss is
// life loss, not damage. Cyberman Patrol grants it and Neheb, the Eternal
// prints it; each declares its own becomes-blocked condition.
func afflictDefenderLoses(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		defender := item.Trigger.Event.Actor
		if g.PlayerByIDForEffect(defender) == nil {
			return nil
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -n)
	}
}
