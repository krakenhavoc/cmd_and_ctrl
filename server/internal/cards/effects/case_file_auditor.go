package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case File Auditor — Creature — Human Detective {2}{W}, 1/4 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters and whenever you solve a Case, look at
//	 the top six cards of your library. You may reveal an enchantment
//	 card from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order.
//	 You may spend mana as though it were mana of any color to cast Case
//	 spells."
//
// One triggered ability with two conditions (OnAny, Gadget Technician's
// shape): this creature entering, or a Case becoming solved under its
// controller (EventCaseSolved, CR 719.3, whose Actor is the Case's
// controller). The body is Dragonologist's look-and-take with six cards
// and an enchantment filter. The spending grant is Oath of Nissa's
// narrowed AnyColorSpend, covering only a cast whose spell has the Case
// subtype; it reaches no activation and no other cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c612741-2b30-43b3-bd3f-7df19662d298",
		Name:         "Case File Auditor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventCaseSolved},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					if ev.Kind == game.EventCaseSolved {
						return ev.Actor == source.Controller
					}
					return ev.CardID == source.InstanceID
				},
				"Case File Auditor — look at the top six cards; you may reveal an enchantment card and put it into your hand",
				LookAtTopThenMayTakeToHand(6, Enchantment(), 1,
					"Case File Auditor — reveal an enchantment card and put it into your hand")),
		},
		AnyColorSpend: []game.AnyColorSpendStatic{{
			Label: "You may spend mana as though it were mana of any color to cast Case spells.",
			Whose: game.AnyColorSpendYou,
			Covers: func(ctx game.ManaSpendContext) bool {
				if ctx.Purpose != game.SpendPurposeCast {
					return false
				}
				for _, s := range ctx.Subtypes {
					if strings.EqualFold(s, "Case") {
						return true
					}
				}
				return false
			},
		}},
	})
}
