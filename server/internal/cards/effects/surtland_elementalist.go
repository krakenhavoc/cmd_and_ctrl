package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surtland Elementalist — Creature — Giant Wizard {5}{U}{U}, 8/8:
//
//	"As an additional cost to cast this spell, reveal a Giant card from your hand or pay {2}.
//	 Whenever this creature attacks, you may cast an instant or sorcery spell from your hand without paying its mana cost."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b1a3b2ab-d9f8-4a0a-b2e2-aaa9989b2e41",
		Name:           "Surtland Elementalist",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("a", "Giant", "{2}"),
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Surtland Elementalist — you may cast an instant or sorcery spell from your hand without paying its mana cost", func(g *game.Game, item *game.StackItem) error {
				return surtlandElementalistOfferFreeCast(NewContext(g, item), item.Controller)
			}),
		},
	})
}

// surtlandElementalistOfferFreeCast is "you may cast an instant or
// sorcery spell from your hand without paying its mana cost" —
// Kari Zev's Expertise's free cast with the predicate swapped from
// "mana value 2 or less" to "instant or sorcery".
func surtlandElementalistOfferFreeCast(ctx *Context, controller uuid.UUID) error {
	pred := Or(Instant(), Sorcery())
	var candidates []uuid.UUID
	for _, id := range allHandCardIDs(ctx.Game, controller) {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !pred(ctx.Game, controller, c) {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   ctx.Source(),
		Question: "Surtland Elementalist — cast an instant or sorcery spell from your hand without paying its mana cost?",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			g.GrantCastPermissionOverCardForEffect(picked[0], game.CastPermission{
				Player:   controller,
				Cost:     "{0}",
				CastOnly: true,
				Label:    "Surtland Elementalist — cast it without paying its mana cost",
			})
			return nil
		},
	})
	return nil
}
