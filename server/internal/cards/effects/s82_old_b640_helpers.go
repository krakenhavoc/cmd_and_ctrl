package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB640RevealThenDamage is Planeswalker's Fury: the target opponent reveals
// a card at random from their hand, then the enchantment deals damage equal to
// that card's mana value to that player (the mana value read as it is revealed,
// CR 608.2h). An empty hand reveals nothing and deals no damage; an opponent
// who is no longer a legal target is not revealed from or damaged (CR 608.2b).
func OldB640RevealThenDamage(ctx *Context) error {
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	id, ok := OldB639RevealRandomFromHand(ctx, opp, "Planeswalker's Fury")
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || c.ManaValue() == 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: opp, Amount: c.ManaValue()}.Apply(ctx)
}

// OldB640BlueOrBlackCreature is Doomsday Specter's "a blue or black creature
// you control": the colour words are the effect's text (CR 612.2), read through
// the layered view so a creature made blue or black counts.
func OldB640BlueOrBlackCreature(c game.Card) bool {
	return c.IsCreature() && (c.HasColorWord("U") || c.HasColorWord("B"))
}

// OldB640LookAtTheDamagedPlayersHand is "look at that player's hand and choose
// a card from it. The player discards that card" for a combat-damage trigger:
// the player is the one the damage event names, the hand is shown only to the
// trigger's controller, and an empty hand asks nothing (CR 609.3).
func OldB640LookAtTheDamagedPlayersHand(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Target
	if player == uuid.Nil {
		return nil
	}
	return ChooseFromRevealedHand{Player: player, Look: true, Label: "card"}.Apply(NewContext(g, item))
}

// OldB640DamageThenBecomesBlack is Singe: damage to the target creature, then
// the target becomes black (a layer-5 set, CR 105.3 and 613.1e) until end of
// turn. Both are done to a target still legal as the spell resolves (CR 608.2b).
func OldB640DamageThenBecomesBlack(amount int, name string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		id := item.Targets[0].ID
		if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
		return UntilEndOfTurn(ctx, id, nil, name+" — that creature becomes black",
			game.SetColorsMod(ctx.TextColors("B")...))
	}
}
