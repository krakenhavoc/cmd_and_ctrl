package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:       "b1a3b2ab-d9f8-4a0a-b2e2-aaa9989b2e41",
		Name:           "Surtland Elementalist",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.RevealOrPay("a", "Giant", "{2}"),
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Surtland Elementalist — you may cast an instant or sorcery spell from your hand without paying its mana cost", func(g *game.Game, item *game.StackItem) error {
				return SurtlandElementalistOfferFreeCast(effects.NewContext(g, item), item.Controller)
			}),
		},
	})
}

// surtlandElementalistOfferFreeCast is "you may cast an instant or
// sorcery spell from your hand without paying its mana cost" —
// Kari Zev's Expertise's free cast with the clause swapped from
// "mana value 2 or less" to "instant or sorcery", judged on the face being
// cast (#4509): Bonecrusher Giant is offered, and cast free as Stomp.
func SurtlandElementalistOfferFreeCast(ctx *effects.Context, controller uuid.UUID) error {
	effects.OfferFreeHandCast(ctx, controller, surtlandElementalistFreeCast,
		"an instant or sorcery spell", effects.FreeInstantOrSorcery())
	return nil
}

// surtlandElementalistFreeCast is the offer's keyed continuation (#4246).
var surtlandElementalistFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/surtland-elementalist-cast", "Surtland Elementalist")
