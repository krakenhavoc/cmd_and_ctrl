package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// choose_a_background.go — shared shapes for the Background cards
// (#2874, ADR 0144). Append-only.
//
// "Choose a Background" itself is a deck-construction rule (CR
// 702.124k) that internal/deck reads off oracle text; nothing in this
// package implements it. What the Backgrounds print is "Commander
// creatures you own have …": an ADR 0093 layer-6 grant from the
// Background to every commander creature its controller OWNS, whoever
// controls that creature now. The granted ability is the creature's
// own, so "this creature" is the commander and "you" is whoever
// controls it (Agent of the Iron Throne's file says why that matters).
// The predicate is commanderCreatureYouOwn (agent_of_the_iron_throne.go).

// wheneverThisAttacksAPlayerNoOpponentRicher is the Baldur's Gate
// attack trigger that Guild Artisan, Sword Coast Sailor and Agent of
// the Shadow Thieves grant: "Whenever this creature attacks a player,
// if no opponent has more life than that player, …". An attack on a
// planeswalker or a battle does not fire it. The intervening if is
// read as the attack is declared and again as the trigger resolves
// (CR 603.4), against live life totals and the opponents of the
// trigger's controller.
func wheneverThisAttacksAPlayerNoOpponentRicher(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ThisAttacked(ev, source, lki, g) &&
			g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetPlayer &&
			b34NoOpponentHasMoreLifeThan(g, source.Controller, ev.Target)
	}, label, func(g *game.Game, item *game.StackItem) error {
		if !b34NoOpponentHasMoreLifeThan(g, item.Controller, item.Trigger.Event.Target) {
			return nil
		}
		return effect(g, item)
	})
}

// grantToCommanderCreaturesYouOwn is a Background's one static: the
// named bundles, granted to each commander creature its controller
// owns.
func grantToCommanderCreaturesYouOwn(keys ...string) game.StaticAbility {
	return GrantAbilities(commanderCreatureYouOwn, keys...)
}
