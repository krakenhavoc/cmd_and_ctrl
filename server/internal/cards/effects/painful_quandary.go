package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Painful Quandary — Enchantment {3}{B}{B}:
//
//	"Whenever an opponent casts a spell, that player loses 5 life
//	 unless they discard a card."
//
// Rhystic Study's shape with a card instead of mana, which is exactly
// why it waited: PayUnless is welded to a parsed MANA cost, and there
// was no prompt that could put a non-mana "unless" in front of another
// seat. MayChoice addressed to the caster is that prompt (#796/#568).
//
// The two branches are the card's own words rather than Yes / No — the
// question is a choice between two consequences, and a player reading
// "Yes" would have to guess which one it bought.
//
// # Hellbent asks nothing
//
// A caster with an empty hand cannot discard, so there is no choice to
// make and CR 608.2's "as much as possible" takes the life directly.
// Asking anyway would put a prompt whose only answer is the one the
// engine already knows in front of a player mid-cast — and, worse,
// offer a "discard a card" branch the resolver would have to refuse,
// which is the #544 wedge.
//
// Unlike Rhystic Study's tax this prompt BLOCKS the table
// (choice_gate.go), because the life loss is part of the resolving
// trigger rather than a payment settled afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c37051cc-6683-4dbb-b5ff-5c3a5bdab1df",
		Name:         "Painful Quandary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: AnOpponentCast(nil),
				Key:       painfulQuandaryLabel,
				// The caster is captured in Build — ev.Actor for a
				// cast event, the same read Rhystic Study makes —
				// because a trigger's effect is handed the item and
				// not the event that fired it.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					caster := ev.Actor
					if caster == uuid.Nil {
						return nil
					}
					return game.NewTriggeredItem(source, painfulQuandaryLabel,
						func(g *game.Game, item *game.StackItem) error {
							return painfulQuandaryAsk(g, item, caster)
						})
				},
			},
		},
	})
}

// painfulQuandaryLifeLoss is the printed 5.
const painfulQuandaryLifeLoss = 5

// painfulQuandaryLabel is the stack-overlay copy.
const painfulQuandaryLabel = "Painful Quandary — discard a card or lose 5 life"

// painfulQuandaryAsk is the trigger's body: ask the caster.
//
// Caller holds g.mu.
func painfulQuandaryAsk(g *game.Game, item *game.StackItem, caster uuid.UUID) error {
	ctx := NewContext(g, item)
	p := g.PlayerByIDForEffect(caster)
	if p == nil || p.Eliminated {
		return nil
	}
	if p.Hand == nil || p.Hand.Size() == 0 {
		// Nothing to discard: the "unless" cannot be satisfied.
		return g.ChangePlayerLifeForEffect(ctx.Source(), caster, -painfulQuandaryLifeLoss)
	}
	return MayChoice{
		Player:   caster,
		Question: "Painful Quandary — discard a card, or lose 5 life?",
		YesLabel: "Discard a card",
		NoLabel:  "Lose 5 life",
		OnYes:    painfulQuandaryDiscard(caster),
		OnNo:     painfulQuandaryLoseLife(caster),
	}.Apply(ctx)
}

// painfulQuandaryDiscard is the "unless they discard a card" branch.
// A hand that emptied between the question and the answer degrades to
// the life loss, which is the only legal outcome left.
func painfulQuandaryDiscard(caster uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		p := ctx.Game.PlayerByIDForEffect(caster)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), caster, -painfulQuandaryLifeLoss)
		}
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   caster,
			Source:   ctx.Source(),
			N:        1,
			Question: "Painful Quandary — discard a card",
		})
		return nil
	}
}

// painfulQuandaryLoseLife is the default branch.
func painfulQuandaryLoseLife(caster uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), caster, -painfulQuandaryLifeLoss)
	}
}
