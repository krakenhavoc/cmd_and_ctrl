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
//
// Each creature has connived once its discard and its counter are done
// (CR 701.50f), and that is game.EventConnived (#4641): emitted per
// creature, on every branch, by the discard's continuation. Iron Monger
// watches it ("whenever a creature you control connives"). Cyclonus runs
// its own draw and discard and emits it from its own continuation.
//
// Connive is an event a replacement can apply to (#4589,
// game.ConniveForEffect): every connive opens its CR 614 window first and
// draws and discards from a registered tail with what the window settled on,
// the cards a replacement has the player draw first included. Leader,
// Super-Genius is the replacement (DrawACardThenConniveInstead).

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
// continuation, the rest. Connive is an event a replacement can apply to
// (#4589, game.ConniveForEffect), so the creature connives from
// conniveEachTail, which the window runs once it has settled; the
// creatures still to go ride the tail's Carry and then the discard's (#4253),
// so a table waiting on the window or the discard is a restore point. The
// creature conniving is the prompt's source.
func conniveNext(g *game.Game, controller uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return g.ConniveForEffect(controller, ids[0], 1, game.KeyedTail{
		Ref:    conniveEachTail,
		Player: controller,
		Carry:  append([]uuid.UUID(nil), ids...),
	})
}

// conniveEachTail is one creature of a chain connived, or not: Carry is the
// creature and then the ones still to go. A tail told nothing connived (N
// is 0: the window cancelled it) goes on to the next creature. An on-disk
// key: never renamed, never reused.
//
// Registered in init, not at package level: the answer connives the next
// creature, so it refers to itself.
var conniveEachTail game.TailRef

func init() {
	conniveEachTail = game.RegisterTail("connive/each", func(g *game.Game, o game.TailOutcome) error {
		carry := o.Tail.Carry
		if len(carry) == 0 {
			return nil
		}
		if o.Tail.N <= 0 {
			return conniveNext(g, o.Tail.Player, carry[1:])
		}
		return conniveDrawAndDiscard(g, o.Tail.Player, carry[0], o.Amount, o.Tail.N,
			"Connive: discard a card", conniveDiscarded, carry[1:])
	})
}

// conniveDrawAndDiscard is connive's draw and discard once its window has
// settled (CR 701.50a): the cards a replacement has the player draw first
// ("instead you draw a card, then that creature connives"), the connive's own
// `n` cards, and then the discard of `n`, whose continuation puts the counters
// on and says the creature has connived. Shared by every conniving card.
func conniveDrawAndDiscard(g *game.Game, controller, creature uuid.UUID, extraDraws, n int, question string, then game.DiscardThen, carry []uuid.UUID) error {
	for range extraDraws {
		if err := g.DrawNForEffect(controller, 1); err != nil {
			return err
		}
	}
	if err := g.DrawNForEffect(controller, n); err != nil {
		return err
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   controller,
		Source:   creature,
		N:        n,
		Question: question,
	}, then, carry)
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
		// CR 701.50f: connived once the counter step is done, even if none
		// of it was possible. After the counter lands and before the next
		// creature draws, so each creature's event comes in turn.
		g.ConnivedForEffect(creature, r.Player)
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

// Apply opens the connive's window (game.ConniveForEffect); connive X is
// a connive, so Leader replaces it (#4589).
func (c ConniveX) Apply(ctx *Context) error {
	if c.X <= 0 {
		return nil
	}
	return ctx.Game.ConniveForEffect(c.Controller, c.Creature, c.X, game.KeyedTail{
		Ref:    conniveXTail,
		Player: c.Controller,
		Carry:  []uuid.UUID{c.Creature},
	})
}

// conniveXTail draws and discards for a connive X that happened. An on-disk
// key: never renamed, never reused.
var conniveXTail = game.RegisterTail("connive/x", func(g *game.Game, o game.TailOutcome) error {
	if o.Tail.N <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return conniveDrawAndDiscard(g, o.Tail.Player, o.Tail.Carry[0], o.Amount, o.Tail.N,
		"Connive: discard cards", conniveXDiscarded, []uuid.UUID{o.Tail.Carry[0]})
})

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
	if nonland > 0 && onBattlefield(g, creature) {
		if err := g.AddCounterForEffect(creature, game.CounterPlusOne, nonland); err != nil {
			return err
		}
	}
	g.ConnivedForEffect(creature, r.Player)
	return nil
})

// theTargetCreatureConnives is "target creature connives": the first
// creature the ability targeted connives for its controller (Valentina,
// Leader, Super-Genius). A target gone by now has no entry and nothing
// connives (CR 608.2b).
func theTargetCreatureConnives(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{t.ID}}.Apply(ctx)
		}
	}
	return nil
}

// WhenThisEntersItConnives is "When this creature enters, it connives"
// (A.I.M. Scientists, Red Room Recruit): ConniveEach for the one creature
// that entered (sourceConnives, shared with Atlantean Skirmisher).
func WhenThisEntersItConnives(label string) game.TriggeredAbility {
	return WhenThisEnters(label, Do(Instruction(sourceConnives)))
}

// ACreatureYouControlConnived is the When of "whenever a creature you
// control connives" (#4641, CR 701.50f). The event's Actor is the
// controller as the creature connived (CR 701.50b: its last controller,
// when it has left the battlefield), so a creature that died before its
// turn to connive still counts for the player who controlled it. A
// permanent that is not a creature does not (nothing in the catalog
// connives one, but the words say "creature").
func ACreatureYouControlConnived(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventConnived || ev.Actor != source.Controller {
		return false
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); ok {
		return c.IsCreature()
	}
	return true
}

// WheneverACreatureYouControlConnives — "Whenever a creature you control
// connives". One trigger per conniving creature (CR 701.50c: they connive
// one after another).
func WheneverACreatureYouControlConnives(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventConnived, ACreatureYouControlConnived, label, effect)
}

// DrawACardThenConniveInstead is "if a creature you control would connive,
// instead you draw a card, then that creature connives" (Leader,
// Super-Genius, #4589, CR 614.1a).
//
// Connive is an event (game.ConniveForEffect): the replacement adds a card
// drawn first and leaves the creature conniving, with the "N" dropped, since
// "that creature connives" is a plain connive (CR 701.50a). It gets one
// opportunity on the event and the modified event after it (CR 614.5), so one
// Leader is one extra card however the connive is made; a second Leader is a
// second effect and draws a second (CR 616.1f). It applies to a creature of
// yours conniving, whoever's effect makes it, and not to an opponent's
// creature. The event's Actor is the controller as the creature connived
// (CR 701.50b), which is whose draw it is.
func DrawACardThenConniveInstead(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventKeywordAction},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventKeywordAction || ev.KeywordAction != game.KeywordActionConnive || src == nil {
				return false
			}
			if ev.Actor != src.Controller {
				return false
			}
			if c, ok := g.LookupCardForEffect(ev.Source); ok {
				return c.IsCreature()
			}
			return true
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.ConniveExtraDraws++
			ev.KeywordActionCount = 1
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:    label,
		RowLabel: "If a creature you control would connive, instead you draw a card, then that creature connives.",
	}
}
