package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b644_helpers.go — shared bodies for Planeshift's enchantments, removal,
// creatures, counterspells, lands and mana, batch 644 (#5615).

// OldB644SpellsOfEitherColorCostLess is "<A> spells and <B> spells you cast cost
// {1} less to cast" (Thornscape Familiar): a spell of either colour, or both. The
// colours are words of the source's text (CR 612.2).
func OldB644SpellsOfEitherColorCostLess(label, a, b string) game.CostModifier {
	return CostsLess(1, label, YourSpell(), func(q game.CostQuery) bool {
		return q.Card.HasColor(q.Source.TextColor(a)) || q.Card.HasColor(q.Source.TextColor(b))
	})
}

// OldB644AnotherPermanentYouControlTargeted is Cloud Cover's condition: a
// permanent on the battlefield, other than the source, under the source's
// controller's control became the target of a spell or ability an opponent
// controls. EventBecomesTarget is emitted once per target slot (CR 115.3).
func OldB644AnotherPermanentYouControlTargeted(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && OnBattlefield(g, ev.CardID) && c.Controller == you
}

// OldB644BounceTriggeringPermanent returns the permanent the trigger's event named
// to its owner's hand, when it is still on the battlefield.
func OldB644BounceTriggeringPermanent(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !OnBattlefield(g, item.Trigger.Event.CardID) {
		return nil
	}
	return BounceToHand{Target: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
}

// OldB644IfNoCreaturesAttacked is an intervening "if no creatures attacked this
// turn" (CR 603.4): the trigger fires only while none has, and does nothing if one
// has by the time it resolves. Creatures that attacked and have since left still
// count (they attacked).
func OldB644IfNoCreaturesAttacked(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return g.CreaturesAttackedThisTurn() == 0 && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "no creatures attacked this turn", Test: func(ctx *Context) bool {
		return ctx.Game.CreaturesAttackedThisTurn() == 0
	}}, t)
}

// OldB644TriggerPlayerSacrifices is "that player sacrifices a <permanent> of their
// choice": the player the trigger's event named (the player whose step it is)
// chooses among the permanents `match` passes that they control. A player with
// none sacrifices nothing; one with exactly one is not asked.
func OldB644TriggerPlayerSacrifices(question string, match func(game.Card) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Trigger().Event.Actor
		var candidates []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == player && match(c) && !g.CantBeSacrificedForEffect(c.InstanceID) {
				candidates = append(candidates, c.InstanceID)
			}
		}
		if len(candidates) == 0 {
			return nil
		}
		return SacrificeChoice{Player: player, Candidates: candidates, Question: question}.Apply(ctx)
	}
}
