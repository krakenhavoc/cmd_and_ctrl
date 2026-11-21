package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82AfrB09PackTactics is the shape of every pack tactics card (CR 207.2c: an
// ability word with no rules meaning of its own): "Whenever this creature
// attacks, if you attacked with creatures with total power 6 or greater this
// combat, <effect>". The intervening "if" (CR 603.4) is read as the attack is
// declared and again as the trigger resolves, Werewolf Pack Leader's reading
// (packTacticsAttackingPower, packTacticsTotalPower).
func S82AfrB09PackTactics(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
			PackTacticsAttackingPower(g, source.Controller) >= 6
	}, label, func(g *game.Game, item *game.StackItem) error {
		if PackTacticsTotalPower(g, item.Controller, NewContext(g, item).Trigger().Event.Batch) < 6 {
			return nil
		}
		return effect(g, item)
	})
}
