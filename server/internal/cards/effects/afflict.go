package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// Afflict is "Afflict N" (CR 702.130a) as a triggered ability: "Whenever
// this creature becomes blocked, defending player loses N life." `name`
// heads the stack label, "<card> — afflict N: defending player loses N
// life". It is for an ability given to a creature as a bundle (an ability
// sticker, ADR 0208); a card that prints afflict declares its own row.
func Afflict(n int, name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesBlocked},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Kind == game.EventBecomesBlocked && ev.Target == source.InstanceID
		},
		Key:    name + " — afflict " + strconv.Itoa(n) + ": defending player loses " + strconv.Itoa(n) + " life",
		Effect: afflictDefenderLoses(n),
	}
}
