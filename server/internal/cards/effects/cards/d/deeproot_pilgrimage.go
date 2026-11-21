package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeproot Pilgrimage — Enchantment {1}{U} (EDHREC rank 4470):
//
//	"Whenever one or more nontoken Merfolk you control become tapped,
//	 create a 1/1 blue Merfolk creature token with hexproof."
//
// The Merfolk deck's engine for two mana. Every attack is a token,
// every tap ability is a token, and the tokens have HEXPROOF, so the
// board it builds does not fold to targeted removal. It is one of the
// reasons Merfolk is a real Commander archetype rather than a
// tribe with no payoff.
//
// Three clauses, all load-bearing:
//
//   - "NONTOKEN" is the anti-loop clause. The tokens it makes do not
//     make more tokens when they tap, so the card is an engine and
//     not a combo.
//   - "BECOME TAPPED" is not only the tap symbol. Attacking taps a
//     creature without going through the tap action, so the trigger
//     watches EventAttack as well as EventTapCard and asks whether
//     the creature is tapped NOW — which is exactly why a VIGILANCE
//     Merfolk attacking makes no token.
//   - "ONE OR MORE" is one trigger per batch, not one per Merfolk.
//     The engine emits an event per creature, so the later events of
//     the same batch are declined by OncePerBatch — without which a
//     four-Merfolk attack would make four tokens instead of one,
//     which is the #259 direction.
//
// A separate, LATER batch triggers again: tap two Merfolk with two
// abilities in sequence and the Pilgrimage makes two tokens, as
// printed (CR 603.2c).
//
// Hexproof is a canonical keyword and the targeting gate has honoured
// it since S23, so the tokens really are untargetable by opponents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "188656d8-4ea8-4e58-88da-b010a16eb6f2",
		Name:         "Deeproot Pilgrimage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B43NontokenMerfolkYouControlBecameTapped(ev, source, g)
				},
				"Deeproot Pilgrimage — create a 1/1 blue Merfolk with hexproof",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk with hexproof"), N: 1}))),
		},
	})
}
