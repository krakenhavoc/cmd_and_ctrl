package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB289ReturnThisCardTransformed is "return this card to the battlefield
// transformed under your control" (Skin Invasion): the source as the
// graveyard object it became when it left the battlefield (CR 400.7), put
// back with its back face up under its controller. A card that left the
// graveyard while the trigger waited is not returned. The back face is not an
// Aura, so nothing is attached.
func OldB289ReturnThisCardTransformed(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceCardInGraveyardAfterLeaving(ctx) {
		return nil
	}
	return ReturnFromGraveyard{
		Target:      item.SourceCardID,
		Dest:        game.ZoneBattlefield,
		Controller:  item.Controller,
		Transformed: true,
	}.Apply(ctx)
}

// OldB289ThatTargetEnchantedPlayer is "… that target enchanted player" for a
// cost modifier on a Curse: one of the announced targets (CR 601.2c, read at
// 601.2f) is the player the modifier's source is attached to. Reads q.Targets,
// so the modifier must set ReadsTargets.
func OldB289ThatTargetEnchantedPlayer() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Source.AttachedTo.Kind != game.TargetPlayer {
			return false
		}
		for _, t := range q.Targets {
			if t.Kind == game.TargetPlayer && t.ID == q.Source.AttachedTo.ID {
				return true
			}
		}
		return false
	}
}

// OldB289CreaturesYouControl counts the creatures the source's controller
// controls (Hope Against Hope): "you" is the Aura's controller, not the host's
// (CR 109.5), and the host counts itself.
func OldB289CreaturesYouControl(g *game.Game, source *game.Card) int {
	return CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
}

// OldB289HostIsHuman is "as long as enchanted creature is a Human": read on
// every recompute off the host as the earlier layers left it, with the word
// "Human" read through the Aura's text (CR 612.2).
func OldB289HostIsHuman(host *game.Card, _ *game.Game, source *game.Card) bool {
	return host.HasSubtype(source.TextSubtype("Human"))
}

// EnchantedPlayerLosesOneYouGainOne is a Curse's upkeep payoff: "that player
// loses 1 life and you gain 1 life" (Curse of Leeches, Infectious Curse). The
// player is the one whose upkeep it is, read off the trigger's event; the gain
// is not linked to the loss, so it happens whether or not the loss went
// through, and waits for any prompt the loss asks (a CR 616 order; CR 608.2c,
// #4155).
func EnchantedPlayerLosesOneYouGainOne() Effect {
	return Do(
		Instruction(func(ctx *Context) error {
			victim := ctx.Item.Trigger.Event.Actor
			if ctx.Game.PlayerByIDForEffect(victim) != nil {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, victim, -1)
			}
			return nil
		}),
		GainLife{Amount: 1},
	)
}
