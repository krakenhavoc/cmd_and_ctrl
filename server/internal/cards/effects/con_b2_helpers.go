package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// con_b2_helpers.go — the shapes Conflux batch 2 (#5587) prints more
// than once: the Shards' "if you control a <colour> permanent"
// intervening if on an enters or upkeep trigger, and "that player
// discards N cards". Its own file, as #231 asks of a batch, so a
// concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// conB2YouControlA is the CR 603.4 condition "if you control a <colour>
// permanent", checked again as the ability resolves. The colour is a
// word of the source's text (CR 612.2), so a text-changing effect moves
// it. `letter` is the colour's mana letter, `word` its name.
func conB2YouControlA(letter, word string) CondBoard {
	return CondBoard{
		Name: "you control a " + word + " permanent",
		Test: func(ctx *Context) bool {
			return b18ControlsPermanentOfColor(ctx.Game, ctx.Controller(), ctx.TextColor(letter))
		},
	}
}

// conB2WhenThisEntersIfYouControlA is "When this creature enters, if you
// control a <colour> permanent, <effect>." (Sedraxis Alchemist,
// Parasitic Strix): the condition is read as the creature enters, so it
// never triggers without one (CR 603.4), and again on resolution. With
// `targets`, the target is chosen as the trigger goes on the stack.
func conB2WhenThisEntersIfYouControlA(letter, word, label string, targets *game.TargetSpec, effect Effect) game.TriggeredAbility {
	t := On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return Self(ev, source, lki, g) && b18ControlsPermanentOfColor(g, source.Controller, source.TextColor(letter))
	}, label, effect)
	if targets != nil {
		t = Targeting(t, targets)
	}
	return InterveningIf(conB2YouControlA(letter, word), t)
}

// conB2AtYourUpkeepIfYouControlA is "At the beginning of your upkeep, if
// you control a <colour> permanent, <effect>." (Bloodhall Ooze,
// Mirror-Sigil Sergeant), with the same two checks.
func conB2AtYourUpkeepIfYouControlA(letter, word, label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ByYou(ev, source, lki, g) && b18ControlsPermanentOfColor(g, source.Controller, source.TextColor(letter))
	}, label, effect)
	return InterveningIf(conB2YouControlA(letter, word), t)
}

// conB2EnchantedDealtDamageToAPlayer is "Whenever enchanted creature
// deals damage to a player" (Elder Mastery): any damage, combat or not,
// to any player, its controller included.
func conB2EnchantedDealtDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

// conB2ThatPlayerDiscards is "that player discards N cards" for the
// player a damage trigger's event names. The player chooses the cards
// (CR 701.9a), through the ordinary discard prompt.
func conB2ThatPlayerDiscards(n int, question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   item.Trigger.Event.Target,
			Source:   item.SourceCardID,
			N:        n,
			Question: question,
		})
		return nil
	}
}
