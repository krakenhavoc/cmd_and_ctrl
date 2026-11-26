package cardsd

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deicide — Instant {1}{W} (Journey into Nyx, S82 OLD, #5558):
//
//	"Exile target enchantment. If the exiled card is a God card, search its
//	 controller's graveyard, hand, and library for any number of cards with the
//	 same name as that card and exile them. Then that player shuffles."
//
// Crumble to Dust's exile and "any number" extraction, gated on the card that
// reached exile being a God card: the clause reads the exiled card's printed
// subtypes, where it now is, so a permanent that was only a God through an
// effect does not count. The name and the controller are the enchantment's as
// it last existed on the battlefield, read before it is exiled. The shuffle
// only happens with the search (it is part of the same "if").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e421d2bd-740a-40dd-8a10-205edd4991ee",
		Name:         "Deicide",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			victim, names := c.Controller, effects.ExtractionNames(c)
			return effects.ExileTarget{
				Target:   target,
				ThenTail: effects.WithWord(effects.WithCarry(ctx.ItemTail(deicideExiled), victim), strings.Join(append([]string{"Deicide"}, names...), "\n")),
			}.Apply(ctx)
		},
	})
}

// deicideExiled is the clause after the exile: only a God card reaching exile
// opens the search. An on-disk key.
var deicideExiled = effects.CardTail("s82-old-b370/deicide-after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	exiled, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !exiled.HasSubtypeWord("God") {
		return nil
	}
	parts := strings.Split(o.Tail.Word, "\n")
	return effects.ExtractionSearch(ctx, parts[0], effects.TailCarry0(o), parts[1:], false)
})
