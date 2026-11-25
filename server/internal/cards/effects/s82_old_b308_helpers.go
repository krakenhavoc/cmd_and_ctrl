package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b308_helpers.go — shared shapes for the Battle for Zendikar Devoid
// cards of S82 batch 308 (#5547).

// OldB308Ingest is the ingest keyword (CR 702.115a): "whenever this creature
// deals combat damage to a player, that player exiles the top card of their
// library".
func OldB308Ingest(name string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(name+" — ingest: that player exiles the top card of their library",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return MillToZone{Player: item.Trigger.Event.Target, N: 1, To: game.ZoneExile}.Apply(NewContext(g, item))
		})
}

// OldB308ControlsAnotherColorlessCreature is the activation condition "you
// control another colorless creature" (Skitterskin): a colorless creature
// other than the source, read through the layers.
func OldB308ControlsAnotherColorlessCreature(g *game.Game, controller, source uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != source && c.IsCreature() && c.IsColorless() {
			return true
		}
	}
	return false
}

// OldB308ExilePermanentThenExtractAnyNumber is Crumble to Dust's body:
// "Exile target <permanent>. Search its controller's graveyard, hand, and
// library for any number of cards with the same name as that <permanent> and
// exile them. Then that player shuffles." It is ExilePermanentThenExtract with
// "any number" in place of "all": every match, in the graveyard too, is the
// searcher's to take or leave (game.SearchLibrarySpec Unbounded). The name and
// the controller are the permanent's as it last existed on the battlefield,
// read before it is exiled; the search happens whether or not it reached exile;
// a target that is no longer legal makes the spell do nothing (CR 608.2b).
func OldB308ExilePermanentThenExtractAnyNumber(card string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(target)
		if !ok {
			return nil
		}
		victim, names := c.Controller, ExtractionNames(c)
		return ExileTarget{
			Target:   target,
			ThenTail: WithWord(WithCarry(ctx.ItemTail(oldB308ExtractAnyNumberExiled), victim), strings.Join(append([]string{card}, names...), "\n")),
		}.Apply(ctx)
	}
}

// oldB308ExtractAnyNumberExiled is the clause after the exile: the "any number"
// search. An on-disk key.
var oldB308ExtractAnyNumberExiled = CardTail("extraction-family/after-exile-any-number", func(ctx *Context, o game.TailOutcome) error {
	parts := strings.Split(o.Tail.Word, "\n")
	card, names, victim := parts[0], parts[1:], TailCarry0(o)
	return ExtractionSearch(ctx, card, victim, names, false)
})

// OldB308ProcessACardCost is "As an additional cost to cast this spell, put a
// card an opponent owns from exile into that player's graveyard": the
// processing component of an activated ability's cost (ProcessCards) as a
// spell's additional cost.
func OldB308ProcessACardCost() *game.AdditionalCost {
	return &game.AdditionalCost{
		ExileCards: ProcessCards(1, "a card an opponent owns from exile", nil).ExileCards,
		Label:      "Put a card an opponent owns from exile into that player's graveyard",
	}
}
