package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leshrac's Sigil — Enchantment {B}{B} (Secret Lair Drop):
//
//	"Whenever an opponent casts a green spell, you may pay {B}{B}. If you do,
//	 look at that player's hand and choose a card from it. The player discards
//	 that card.
//	 {B}{B}: Return this enchantment to its owner's hand."
//
// The trigger is on the cast of an opponent's green spell; resolving it asks its
// controller to pay {B}{B} (MayPay), and the paid branch is Mind Slash's look-at-
// the-hand-and-choose against the caster, whose hand is private to the
// Sigil's controller. The second ability returns the Sigil as an effect (not a
// cost), so it can be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98388eba-31ab-4302-be9e-ea67ea2b8c1a",
		Name:         "Leshrac's Sigil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("G")),
				"Leshrac's Sigil — you may pay {B}{B}; if you do, look at that player's hand and they discard a card you choose",
				effects.S82OldB022LeshracsSigilOffer),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}: Return this enchantment to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}{B}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
