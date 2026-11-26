package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB345UntapEventCreature is "untap that creature" for a trigger whose
// event names the creature (Pine Walker's turned-face-up creature,
// Raggadragga's attacker): the event's card is untapped as the trigger
// resolves, and nothing happens once it has left the battlefield
// (UntapTarget's guard).
func OldB345UntapEventCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return UntapTarget{Target: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
}
