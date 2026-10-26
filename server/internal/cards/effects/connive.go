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

func conniveChain(g *game.Game, _ *game.StackItem, controller uuid.UUID, ids []uuid.UUID) error {
	return conniveNext(g, controller, ids)
}

// conniveNext makes the first of `ids` connive and, from the discard's
// continuation, the rest. The creatures still to go ride the discard as its
// Carry (#4253), so a table waiting on the discard is a restore point; the
// creature conniving is the prompt's source.
func conniveNext(g *game.Game, controller uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	creature, rest := ids[0], ids[1:]
	if err := g.DrawNForEffect(controller, 1); err != nil {
		return err
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   controller,
		Source:   creature,
		N:        1,
		Question: "Connive: discard a card",
	}, conniveDiscarded, rest)
}

// conniveDiscarded puts the +1/+1 counter for a nonland card discarded and
// goes on to the next creature. An on-disk key: never renamed, never reused.
//
// Registered in init, not at package level: the answer connives the next
// creature, so it refers to itself.
var conniveDiscarded game.DiscardThen

func init() {
	conniveDiscarded = game.RegisterDiscardThen("choose-cards/connive-discarded", func(g *game.Game, r game.DiscardsDone) error {
		creature := r.Source
		if discarded := r.Discarded.By(r.Player); len(discarded) > 0 {
			if card, ok := g.LookupCardForEffect(discarded[0]); ok && !card.IsLand() && onBattlefield(g, creature) {
				if err := g.AddCounterForEffect(creature, game.CounterPlusOne, 1); err != nil {
					return err
				}
			}
		}
		return conniveNext(g, r.Player, r.Carry)
	})
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
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   c.Controller,
		Source:   c.Creature,
		N:        c.X,
		Question: "Connive: discard cards",
	}, conniveXDiscarded, []uuid.UUID{c.Creature})
}

// conniveXDiscarded puts a +1/+1 counter on the creature for each nonland
// card discarded. An on-disk key: never renamed, never reused.
var conniveXDiscarded = game.RegisterDiscardThen("choose-cards/connive-x-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	creature := r.Carry[0]
	nonland := 0
	for _, id := range r.Discarded.By(r.Player) {
		if card, ok := g.LookupCardForEffect(id); ok && !card.IsLand() {
			nonland++
		}
	}
	if nonland == 0 || !onBattlefield(g, creature) {
		return nil
	}
	return g.AddCounterForEffect(creature, game.CounterPlusOne, nonland)
})
