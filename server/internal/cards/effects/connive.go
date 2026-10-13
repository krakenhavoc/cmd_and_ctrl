package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// connive.go — connive (CR 701.50, #3172): "draw a card, then discard a
// card. If you discarded a nonland card, put a +1/+1 counter on this
// creature."
//
// Ledger Shredder conniving itself is the one-creature case
// (ledgerShredderConnive). This is the general body for an effect that
// makes SEVERAL creatures connive, as Lethal Scheme's "each creature that
// convoked this spell connives" does.
//
// Each creature connives completely before the next begins (CR 701.50a
// is per creature, and the second draw must not come before the first
// discard has been answered, the same ordering Scry's `Then` documents),
// so the chain runs through the discard prompt's own continuation.
// The controller chooses nothing about the order beyond what the effect
// names; the list is taken in the order given.
//
// A creature that has left the battlefield by the time its turn comes
// still draws and discards (CR 701.50c: the connive still happens, there
// is just no creature to put the counter on), and the counter is skipped.

// ConniveEach makes every creature in `ids` connive in turn for
// `controller`.
type ConniveEach struct {
	Controller uuid.UUID
	Creatures  []uuid.UUID
}

// Apply starts the chain.
func (c ConniveEach) Apply(ctx *Context) error {
	return conniveChain(ctx.Game, ctx.Item, c.Controller, append([]uuid.UUID(nil), c.Creatures...))
}

func conniveChain(g *game.Game, item *game.StackItem, controller uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	creature, rest := ids[0], ids[1:]
	if err := (DrawCards{Player: controller, N: 1}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   controller,
		Source:   creature,
		N:        1,
		Question: "Connive: discard a card",
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			if len(discarded) > 0 {
				if card, ok := g.LookupCardForEffect(discarded[0]); ok && !card.IsLand() && onBattlefield(g, creature) {
					if err := g.AddCounterForEffect(creature, game.CounterPlusOne, 1); err != nil {
						return err
					}
				}
			}
			return conniveChain(g, item, controller, rest)
		},
	})
	return nil
}

// ConniveX makes ONE creature connive X (CR 701.50): draw X cards, then
// discard X cards, then put a +1/+1 counter on the creature for each
// nonland card discarded this way. Raffine, Scheming Seer is the user,
// where X is the number of attacking creatures. X is read by the caller
// when the ability resolves.
//
// The discard is one prompt for all X cards, and the counters are counted
// on the answer, so discarding two nonland cards is two counters. A
// creature that has left the battlefield by then still draws and discards
// (CR 701.50c) and gets nothing. Fewer cards in hand than X discards the
// hand (CR 701.9a), which DiscardPrompt already clamps.
type ConniveX struct {
	Controller uuid.UUID
	Creature   uuid.UUID
	X          int
}

// Apply draws, then queues the discard whose continuation adds counters.
func (c ConniveX) Apply(ctx *Context) error {
	if c.X <= 0 {
		return nil
	}
	g := ctx.Game
	if err := (DrawCards{Player: c.Controller, N: c.X}).Apply(ctx); err != nil {
		return err
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   c.Controller,
		Source:   c.Creature,
		N:        c.X,
		Question: "Connive: discard cards",
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			nonland := 0
			for _, id := range discarded {
				if card, ok := g.LookupCardForEffect(id); ok && !card.IsLand() {
					nonland++
				}
			}
			if nonland == 0 || !onBattlefield(g, c.Creature) {
				return nil
			}
			return g.AddCounterForEffect(c.Creature, game.CounterPlusOne, nonland)
		},
	})
	return nil
}
