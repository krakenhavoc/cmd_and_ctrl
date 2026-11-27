package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB562GainThatMuchLife is "you gain that much life" as a damage
// trigger's effect: the trigger's controller gains the amount of the damage
// event that fired it (Tamanoa).
func S82OldB562GainThatMuchLife(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}
